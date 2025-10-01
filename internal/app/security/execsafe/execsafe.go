// Package execsafe provides utilities for executing system commands safely.
package execsafe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ---------- parsing & sanitation ----------

var (
	// bloqueios hard de shell-metachar
	metaBad = regexp.MustCompile(`(?s)[;&|><` + "`" + `]`)
	// múltiplos espaços/quebras => 1 espaço
	spaceRx = regexp.MustCompile(`\s+`)
)

// tokenize estilo shlex simples (sem dependências). Preserva aspas "..."
func shlexSplit(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	inQuote := false
	esc := false
	for _, r := range s {
		switch {
		case esc:
			cur.WriteRune(r)
			esc = false
		case r == '\\':
			esc = true
		case r == '"':
			inQuote = !inQuote
		case unicode.IsSpace(r) && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if inQuote {
		return nil, errors.New("unclosed quote")
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out, nil
}

// ExtractShellCommand Extrai comando do texto livre (pt/en); retorna vazio se não achar.
func ExtractShellCommand(content string) string {
	s := norm.NFKC.String(content)
	low := strings.ToLower(s)
	triggers := []string{"executar ", "execute ", "rodar ", "run ", "exec ", "executa "}
	idx := -1
	for _, t := range triggers {
		if j := strings.Index(low, t); j != -1 {
			idx = j + len(t)
			break
		}
	}
	if idx == -1 || idx >= len(s) {
		return ""
	}
	cmd := strings.TrimSpace(s[idx:])
	cmd = spaceRx.ReplaceAllString(cmd, " ")
	return cmd
}

// ---------- registry & validation ----------

type ArgValidator func(args []string) error

type CommandSpec struct {
	Binary       string        // executável, ex: "ls"
	ArgsValidate ArgValidator  // valida args
	Timeout      time.Duration // ex: 3s
	WorkDir      string        // opcional
	MaxOutputKB  int           // truncar saída (por stream)
	EnvAllowList []string      // nomes de env que podem vazar
}

type Registry struct {
	allow map[string]CommandSpec // chave: nome lógico (e.g., "ls", "ps")
}

func NewRegistry() *Registry { return &Registry{allow: map[string]CommandSpec{}} }

func (r *Registry) Register(name string, spec CommandSpec) {
	r.allow[strings.ToLower(name)] = spec
}

func (r *Registry) Get(name string) (CommandSpec, bool) {
	sp, ok := r.allow[strings.ToLower(name)]
	return sp, ok
}

// Helpers de validação

func RegexValidator(rx *regexp.Regexp) ArgValidator {
	return func(args []string) error {
		for _, a := range args {
			if !rx.MatchString(a) {
				return fmt.Errorf("arg inválido: %q", a)
			}
			if metaBad.MatchString(a) {
				return fmt.Errorf("arg contém metachar proibido: %q", a)
			}
		}
		return nil
	}
}

func OneOfFlags(allowed ...string) ArgValidator {
	set := map[string]struct{}{}
	for _, f := range allowed {
		set[f] = struct{}{}
	}
	return func(args []string) error {
		for _, a := range args {
			if strings.HasPrefix(a, "-") {
				if _, ok := set[a]; !ok {
					return fmt.Errorf("flag não permitida: %s", a)
				}
			}
		}
		return nil
	}
}

func Chain(validators ...ArgValidator) ArgValidator {
	return func(args []string) error {
		for _, v := range validators {
			if v == nil {
				continue
			}
			if err := v(args); err != nil {
				return err
			}
		}
		return nil
	}
}

// ---------- runner ----------

type ExecResult struct {
	Cmd       string
	Args      []string
	ExitCode  int
	Duration  time.Duration
	Stdout    string
	Stderr    string
	Truncated bool
}

// Result represents the outcome of an Exec call.
type Result struct {
	Stdout      string
	Stderr      string
	ExitCode    int
	DurationMs  int64
	ResolvedBin string
	Truncated   bool
}

// Options defines execution parameters for Exec.
type Options struct {
	CWD       string
	Env       []string
	Timeout   time.Duration
	UseShell  bool
	Allowlist []string
	MaxBytes  int
}

// Exec executes a command with hardened defaults (no shell by default, allowlist enforcement,
// process-group kill on timeout, bounded stdout/stderr capture).
func Exec(ctx context.Context, bin string, args []string, opt Options) (Result, error) {
	start := time.Now()
	if opt.MaxBytes <= 0 {
		opt.MaxBytes = 256 << 10 // 256KB default per stream
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 10 * time.Second
	}

	bin = strings.TrimSpace(bin)
	if bin == "" {
		return Result{}, errors.New("command required")
	}

	base := filepath.Base(bin)
	if len(opt.Allowlist) > 0 {
		allowed := false
		for _, candidate := range opt.Allowlist {
			if candidate == base {
				allowed = true
				break
			}
		}
		if !allowed {
			return Result{}, fmt.Errorf("command '%s' not allowed", base)
		}
	}

	resolved := resolveBinaryPath(bin)
	if opt.UseShell {
		// When shell is enabled, allow shell expansions but still prefer the resolved path when available.
		if resolved == "" {
			resolved = bin
		}
	} else {
		if resolved == "" {
			resolved = bin
		}
	}

	cmd := buildCommand(resolved, args, opt.UseShell)
	if opt.CWD != "" {
		cmd.Dir = opt.CWD
	}
	if len(opt.Env) > 0 {
		cmd.Env = append(os.Environ(), opt.Env...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutBuf := newLimitedBuffer(opt.MaxBytes)
	stderrBuf := newLimitedBuffer(opt.MaxBytes)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	tctx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	if err := cmd.Start(); err != nil {
		res := Result{
			Stdout:      stdoutBuf.StringWithNotice(),
			Stderr:      stderrBuf.StringWithNotice(),
			ExitCode:    startExitCode(err),
			DurationMs:  time.Since(start).Milliseconds(),
			ResolvedBin: resolved,
			Truncated:   stdoutBuf.Truncated() || stderrBuf.Truncated(),
		}
		return res, fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var waitErr error
	timedOut := false

	select {
	case waitErr = <-done:
	case <-tctx.Done():
		timedOut = true
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // best-effort kill of process group
		waitErr = <-done
	}

	dur := time.Since(start)
	stdoutStr := stdoutBuf.StringWithNotice()
	stderrStr := stderrBuf.StringWithNotice()
	truncated := stdoutBuf.Truncated() || stderrBuf.Truncated()
	res := Result{
		Stdout:      stdoutStr,
		Stderr:      stderrStr,
		ExitCode:    waitExitCode(waitErr, timedOut),
		DurationMs:  dur.Milliseconds(),
		ResolvedBin: resolved,
		Truncated:   truncated,
	}

	if timedOut {
		return res, fmt.Errorf("timeout after %s", opt.Timeout)
	}
	if waitErr != nil {
		return res, fmt.Errorf("exit=%d: %s", res.ExitCode, SanitizeOneLine(stderrStr))
	}
	return res, nil
}

func RunSafe(ctx context.Context, reg *Registry, name string, args []string) (*ExecResult, error) {
	spec, ok := reg.Get(name)
	if !ok {
		return nil, fmt.Errorf("comando não permitido: %s", name)
	}
	// bloqueios globais
	for _, a := range args {
		if metaBad.MatchString(a) {
			return nil, fmt.Errorf("metachar proibido em argumentos")
		}
	}

	if spec.ArgsValidate != nil {
		if err := spec.ArgsValidate(args); err != nil {
			return nil, err
		}
	}

	tmo := spec.Timeout
	if tmo <= 0 {
		tmo = 3 * time.Second
	}

	cctx, cancel := context.WithTimeout(ctx, tmo)
	defer cancel()

	cmd := exec.CommandContext(cctx, spec.Binary, args...) // SEM shell
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}
	// env controlado
	if len(spec.EnvAllowList) > 0 {
		base := []string{}
		for _, k := range spec.EnvAllowList {
			if v, ok := os.LookupEnv(k); ok {
				base = append(base, fmt.Sprintf("%s=%s", k, v))
			}
		}
		cmd.Env = append(base, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	runErr := cmd.Run()
	dur := time.Since(start)

	res := &ExecResult{
		Cmd:      spec.Binary,
		Args:     args,
		Duration: dur,
		ExitCode: exitCodeOf(runErr),
	}

	maxKB := spec.MaxOutputKB
	if maxKB <= 0 {
		maxKB = 256
	} // default 256KB por stream
	res.Stdout, res.Truncated = TruncateKB(outBuf.String(), maxKB)
	stderrStr, trunc2 := TruncateKB(errBuf.String(), maxKB)
	res.Stderr = stderrStr
	res.Truncated = res.Truncated || trunc2

	// contexto cancelado vira timeout
	if errors.Is(runErr, context.DeadlineExceeded) {
		return res, fmt.Errorf("timeout após %s", dur)
	}
	// exit code != 0 vira erro com stderr
	if runErr != nil {
		return res, fmt.Errorf("exit=%d: %s", res.ExitCode, SanitizeOneLine(stderrStr))
	}
	return res, nil
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func newLimitedBuffer(limit int) *limitedBuffer {
	if limit <= 0 {
		return &limitedBuffer{limit: 0}
	}
	return &limitedBuffer{limit: limit}
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	if l.limit <= 0 {
		return l.buf.Write(p)
	}
	remaining := l.limit - l.buf.Len()
	if remaining <= 0 {
		l.truncated = true
		return len(p), nil
	}
	if len(p) <= remaining {
		return l.buf.Write(p)
	}
	l.buf.Write(p[:remaining])
	l.truncated = true
	return len(p), nil
}

func (l *limitedBuffer) Truncated() bool { return l.truncated }

func (l *limitedBuffer) StringWithNotice() string {
	str := l.buf.String()
	if l.truncated {
		return str + "\n…(truncated)…"
	}
	return str
}

func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// TruncateKB truncates string to specified KB limit (exported for testing)
func TruncateKB(s string, kb int) (string, bool) {
	lim := kb * 1024
	if len(s) <= lim {
		return s, false
	}
	return s[:lim] + "\n…(truncated)…", true
}

// SanitizeOneLine sanitizes string to single line (exported for testing)
func SanitizeOneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return spaceRx.ReplaceAllString(strings.TrimSpace(s), " ")
}

// ---------- high-level: parse + run ----------

type Parsed struct {
	Name string
	Args []string
}

func ParseUserCommand(text string) (*Parsed, error) {
	raw := ExtractShellCommand(text)
	if raw == "" {
		return nil, errors.New("nenhum comando encontrado")
	}
	if metaBad.MatchString(raw) {
		return nil, errors.New("uso de metachar proibido")
	}
	toks, err := shlexSplit(raw)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		return nil, errors.New("comando vazio")
	}
	return &Parsed{Name: filepath.Base(toks[0]), Args: toks[1:]}, nil
}

func resolveBinaryPath(bin string) string {
	bin = strings.TrimSpace(bin)
	if bin == "" {
		return ""
	}
	if strings.Contains(bin, "/") {
		return bin
	}
	if path, err := exec.LookPath(bin); err == nil {
		return path
	}
	fallbacks := []string{
		filepath.Join("/bin", bin),
		filepath.Join("/usr/bin", bin),
	}
	for _, candidate := range fallbacks {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return bin
}

func buildCommand(bin string, args []string, useShell bool) *exec.Cmd {
	if useShell {
		script := shEscape(bin, args...)
		return exec.Command("/bin/sh", "-lc", script)
	}
	return exec.Command(bin, args...)
}

func startExitCode(err error) int {
	if errors.Is(err, exec.ErrNotFound) {
		return 127
	}
	return 1
}

func waitExitCode(err error, timedOut bool) int {
	if timedOut {
		return 124
	}
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus()
		}
		return exitErr.ExitCode()
	}
	if errors.Is(err, exec.ErrNotFound) {
		return 127
	}
	return 1
}

func shEscape(bin string, args ...string) string {
	esc := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return "''"
		}
		if strings.ContainsAny(s, " '\"$`\\") {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
		return s
	}
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, esc(bin))
	for _, arg := range args {
		parts = append(parts, esc(arg))
	}
	return strings.Join(parts, " ")
}
