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
