// Package types provides types and methods for managing environment variables,
package types

//// go:build !windows
//// +build !windows

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	ci "github.com/rafa-mori/gobe/internal/interfaces"
	crp "github.com/rafa-mori/gobe/internal/security/crypto"
	sci "github.com/rafa-mori/gobe/internal/security/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type EnvCache struct {
	m map[string]string
}

func NewEnvCache() *EnvCache {
	return &EnvCache{
		m: make(map[string]string),
	}
}

type Environment struct {
	isConfidential bool

	Logger l.Logger

	*Reference

	*EnvCache

	*Mutexes

	cpuCount int
	memTotal int
	hostname string
	os       string
	kernel   string
	envFile  string

	// For lazy loading if needed
	properties map[string]any

	mapper ci.IMapper[map[string]string]
}

func newEnvironment(envFile string, isConfidential bool, logger l.Logger) (*Environment, error) {
	if logger == nil {
		logger = l.GetLogger("Environment")
	}
	if envFile == "" {
		envFile = ".env"
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			if createErr := os.WriteFile(envFile, []byte(""), 0644); createErr != nil {
				gl.Log("error", fmt.Sprintf("Error creating env file: %s", createErr.Error()))
				return nil, fmt.Errorf("error creating env file: %s", createErr.Error())
			}
		}
	} else {
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			gl.Log("error", fmt.Sprintf("Error checking env file: %s", err.Error()))
			//return nil, fmt.Errorf("error checking env file: %s", err.Error())
			if createErr := os.WriteFile(envFile, []byte(""), 0644); createErr != nil {
				gl.Log("error", fmt.Sprintf("Error creating env file: %s", createErr.Error()))
				return nil, fmt.Errorf("error creating env file: %s", createErr.Error())
			}
		}
	}

	gl.Log("notice", "Creating new Environment instance")
	cpuCount := runtime.NumCPU()
	memTotal := syscall.Sysinfo_t{}.Totalram
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		gl.Log("error", fmt.Sprintf("Error getting hostname: %s", hostnameErr.Error()))
		return nil, fmt.Errorf("error getting hostname: %s", hostnameErr.Error())
	}
	oos := runtime.GOOS
	kernel := runtime.GOARCH
	name := filepath.Base(envFile)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if name == "" {
		name = "default"
	}
	name = strings.Join(filepath.SplitList(name), "_")

	env := &Environment{
		isConfidential: isConfidential,
		Logger:         logger,
		Reference:      NewReference(name).GetReference(),
		Mutexes:        NewMutexesType(),
		cpuCount:       cpuCount,
		memTotal:       int(memTotal),
		hostname:       hostname,
		os:             oos,
		kernel:         kernel,
		envFile:        envFile,
	}

	env.EnvCache = NewEnvCache()
	env.EnvCache.m = make(map[string]string)

	envs := os.Environ()
	for _, ev := range envs {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		env.EnvCache.m[key] = value
	}
	env.EnvCache.m["ENV_FILE"] = envFile
	env.EnvCache.m["ENV_CONFIDENTIAL"] = fmt.Sprintf("%t", isConfidential)
	env.EnvCache.m["ENV_HOSTNAME"] = env.Hostname()
	env.EnvCache.m["ENV_OS"] = env.Os()
	env.EnvCache.m["ENV_KERNEL"] = env.Kernel()
	env.EnvCache.m["ENV_CPU_COUNT"] = fmt.Sprintf("%d", env.CPUCount())
	env.EnvCache.m["ENV_MEM_TOTAL"] = fmt.Sprintf("%d", env.MemTotal())
	env.EnvCache.m["ENV_MEM_AVAILABLE"] = fmt.Sprintf("%d", env.MemAvailable())
	env.EnvCache.m["ENV_MEM_USED"] = fmt.Sprintf("%d", env.MemTotal()-env.MemAvailable())

	env.mapper = NewMapperTypeWithObject(&env.EnvCache.m, env.envFile)
	_, err := env.mapper.DeserializeFromFile("env")
	if err != nil {
		return nil, fmt.Errorf("error loading file: %s", err.Error())
	}

	return env, nil
}
func NewEnvironment(envFile string, isConfidential bool, logger l.Logger) (ci.IEnvironment, error) {
	return newEnvironment(envFile, isConfidential, logger)
}
func NewEnvironmentType(envFile string, isConfidential bool, logger l.Logger) (*Environment, error) {
	return newEnvironment(envFile, isConfidential, logger)
}

func (e *Environment) Mu() ci.IMutexes {
	if e.Mutexes == nil {
		e.Mutexes = NewMutexesType()
	}
	return e.Mutexes
}
func (e *Environment) CPUCount() int {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	if e.cpuCount == 0 {
		e.cpuCount = runtime.NumCPU()
	}
	return e.cpuCount
}
func (e *Environment) MemTotal() int {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	if e.memTotal == 0 {
		var mem syscall.Sysinfo_t
		err := syscall.Sysinfo(&mem)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting memory info: %s", err.Error()))
			return 0
		}
		totalRAM := mem.Totalram * uint64(mem.Unit) / (1024 * 1024)
		e.memTotal = int(totalRAM)
	}
	return e.memTotal
}
func (e *Environment) Hostname() string {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	if e.hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting hostname: %s", err.Error()))
			return ""
		}
		e.hostname = hostname
	}
	return e.hostname
}
func (e *Environment) Os() string {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	if e.os == "" {
		e.os = runtime.GOOS
	}
	return e.os
}
func (e *Environment) Kernel() string {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	if e.kernel == "" {
		e.kernel = runtime.GOARCH
	}
	return e.kernel
}
func (e *Environment) Getenv(key string) string {
	if val, exists := e.EnvCache.m[key]; exists {
		if val == "" {
			gl.Log("info", fmt.Sprintf("'%s' found in cache, but value is empty", key))
			return ""
		}
		isEncryptedValue := e.IsEncryptedValue(val)
		if isEncryptedValue {
			gl.Log("debug", fmt.Sprintf("'%s' found in cache, value is encrypted", key))
			decryptedVal, err := e.DecryptEnv(val)
			if err != nil {
				gl.Log("error", fmt.Sprintf("Error decrypting value for key '%s': %v", key, err))
				gl.Log("error", fmt.Sprintf("Value for key %s: %s", key, val))
				return ""
			}
			gl.Log("debug", fmt.Sprintf("Decrypted value for key '%s': %s", key, decryptedVal))
			return decryptedVal
		}
		if err := e.Setenv(key, val); err != nil {
			gl.Log("error", fmt.Sprintf("Error setting environment variable '%s': %v", key, err))
			return ""
		}
		return val
	}
	gl.Log("debug", fmt.Sprintf("'%s' not found in cache, checking system env...", key))
	return os.Getenv(key)
}
func (e *Environment) Setenv(key, value string) error {
	if e.EnvCache.m == nil {
		e.EnvCache.m = make(map[string]string)
	}
	isEncrypted := e.IsEncryptedValue(value)
	if e.isConfidential {
		if isEncrypted {
			e.EnvCache.m[key] = value
		} else {
			encryptedValue, err := e.EncryptEnv(value)
			if err != nil {
				gl.Log("error", fmt.Sprintf("Error encrypting value for key '%s': %v", key, err))
				return err
			}
			e.EnvCache.m[key] = encryptedValue
		}
	} else {
		if isEncrypted {
			decryptedValue, err := e.DecryptEnv(value)
			if err != nil {
				gl.Log("error", fmt.Sprintf("Error decrypting value for key '%s': %v", key, err))
			} else if decryptedValue != "" {
				e.EnvCache.m[key] = decryptedValue
			}
		}
		e.EnvCache.m[key] = value
	}

	gl.Log("debug", fmt.Sprintf("Key '%s' value: %s", key, value))

	return os.Setenv(key, value)
}
func (e *Environment) GetEnvCache() map[string]string {
	if e.EnvCache.m == nil {
		gl.Log("debug", "EnvCache is nil, initializing...")
		e.EnvCache.m = make(map[string]string)
	}

	return e.EnvCache.m
}
func (e *Environment) ParseEnvVar(s string) (string, string) {
	name, length := e.GetShellName(s)
	if length == 0 {
		return "", ""
	}
	value := os.Getenv(name)
	return name, value
}
func (e *Environment) LoadEnvFromShell() error {
	cmd := exec.Command("bash", "-c", "env")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("erro ao carregar env via shell: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		e.EnvCache.m[parts[0]] = parts[1]
		if setEnvErr := os.Setenv(parts[0], parts[1]); setEnvErr != nil {
			return setEnvErr
		}
	}

	gl.Log("debug", "Environment variables loaded from shell")
	return nil
}
func (e *Environment) MemAvailable() int {
	e.Mutexes.MuRLock()
	defer e.Mutexes.MuRUnlock()

	var mem syscall.Sysinfo_t
	if err := syscall.Sysinfo(&mem); err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao obter RAM disponível: %v", err))
		return -1
	}
	return int(mem.Freeram * uint64(mem.Unit) / (1024 * 1024))
}
func (e *Environment) GetShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && IsShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2
				}
				return s[1:i], i + 1
			}
		}
		return "", 1
	case IsShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	var i int
	for i = 0; i < len(s) && IsAlphaNum(s[i]); i++ {
	}
	return s[:i], i
}
func (e *Environment) GetEnvFilePath() string { return e.envFile }

func (e *Environment) BackupEnvFile() error {
	backupFile := e.envFile + ".backup"
	if _, err := os.Stat(backupFile); err == nil {
		return nil
	}

	return asyncCopyFile(e.envFile, backupFile)
}
func (e *Environment) EncryptEnvFile() error {
	if !e.isConfidential {
		gl.Log("debug", "Environment is not confidential, skipping encryption")
		return nil
	}
	isEncrypted := e.IsEncrypted(e.envFile)
	if isEncrypted {
		return nil
	}

	if err := e.BackupEnvFile(); err != nil {
		return err
	}

	data, err := os.ReadFile(e.envFile)
	if err != nil {
		return err
	}

	encryptedData, err := e.EncryptEnv(string(data))
	if err != nil {
		return err
	}

	return os.WriteFile(e.envFile, []byte(encryptedData), 0644)
}
func (e *Environment) DecryptEnvFile() (string, error) {
	isEncrypted := e.IsEncrypted(e.envFile)
	if !isEncrypted {
		gl.Log("debug", "Env file is not encrypted, skipping decryption")
		return "", nil
	}

	data, err := os.ReadFile(e.envFile)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error reading env file: %v", err))
		return "", err
	}
	if len(data) == 0 {
		gl.Log("error", "Env file is empty")
		return "", fmt.Errorf("env file is empty")
	}

	return e.DecryptEnv(string(data))
}
func (e *Environment) EncryptEnv(value string) (string, error) {
	if !e.isConfidential || e.IsEncryptedValue(value) {
		return value, nil
	}

	cryptoService, key, err := getKey(e)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error getting key: %v", err))
		return "", err
	}

	if cryptoService == nil {
		gl.Log("error", "CryptoService is nil")
		return "", fmt.Errorf("cryptoService is nil")
	}

	isEncrypted := cryptoService.IsEncrypted([]byte(value))
	if isEncrypted {
		return value, nil
	}

	encrypt, _, err := cryptoService.Encrypt([]byte(value), key)
	if err != nil {
		return "", err
	}

	encoded := cryptoService.EncodeBase64([]byte(encrypt))
	if len(encoded) == 0 {
		gl.Log("error", "Failed to encode the encrypted value")
		return "", fmt.Errorf("failed to encode the encrypted value")
	}

	return encoded, nil
}
func (e *Environment) DecryptEnv(encryptedValue string) (string, error) {
	if !e.isConfidential {
		if !e.IsEncryptedValue(encryptedValue) {
			return encryptedValue, nil
		}
	} else {
		if !e.IsEncryptedValue(encryptedValue) {
			gl.Log("debug", "Value is not encrypted")
			return encryptedValue, nil
		}
	}

	cryptoService, key, err := getKey(e)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error getting key: %v", err))
		return "", err
	}

	isEncrypted := e.IsEncryptedValue(encryptedValue)
	if !isEncrypted {
		return encryptedValue, nil
	}

	isEncoded := cryptoService.IsBase64String(encryptedValue)
	var decodedData string
	if isEncoded {
		decodedBytes, decryptedBytesErr := cryptoService.DecodeBase64(encryptedValue)
		if decryptedBytesErr != nil {
			gl.Log("error", fmt.Sprintf("Error decoding base64 string: %v", decryptedBytesErr))
			return "", decryptedBytesErr
		}
		decodedData = strings.TrimSpace(string(decodedBytes))
	} else {
		decodedData = strings.TrimSpace(encryptedValue)
	}
	trimmedDataBytes := bytes.TrimSpace([]byte(decodedData))

	decrypted, _, err := cryptoService.Decrypt(trimmedDataBytes, key)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error decrypting value: %v", err))
		return "", err
	}

	if len(decrypted) == 0 {
		gl.Log("error", "Decrypted value is empty")
		return "", fmt.Errorf("decrypted value is empty")
	}

	return strings.TrimSpace(string(decrypted)), nil
}
func (e *Environment) IsEncrypted(envFile string) bool {
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		gl.Log("error", fmt.Sprintf("Arquivo não encontrado: %v", err))
		return false
	}
	cryptoService, _, err := getKey(e)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error getting key: %v", err))
		return false
	}
	if cryptoService == nil {
		gl.Log("error", "CryptoService is nil")
		return false
	}
	data, err := os.ReadFile(envFile)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error reading file: %v", err))
		return false
	}
	if len(data) == 0 {
		gl.Log("error", "File is empty")
		return false
	}
	return cryptoService.IsEncrypted(data)
}
func (e *Environment) IsEncryptedValue(value string) bool {
	if arrB, arrBErr := base64.URLEncoding.DecodeString(value); arrBErr != nil || len(arrB) == 0 {
		return false
	} else {
		return len(arrB) > 0 && arrB[0] == 0x00
	}
}
func (e *Environment) EnableEnvFileEncryption() error {
	if e.isConfidential {
		gl.Log("debug", "Environment is already confidential, skipping encryption")
		return nil
	}

	e.isConfidential = true

	if err := e.EncryptEnvFile(); err != nil {
		return err
	}

	return nil
}
func (e *Environment) DisableEnvFileEncryption() error {
	if !e.isConfidential {
		gl.Log("debug", "Environment is not confidential, skipping decryption")
		return nil
	}

	e.isConfidential = false

	if err := e.EncryptEnvFile(); err != nil {
		return err
	}

	return nil
}
func (e *Environment) LoadEnvFile(watchFunc func(ctx context.Context, chanCbArg chan any) <-chan any) error {
	timeout := 10 * time.Second
	chanErr := make(chan error, 3)
	chanDone := make(chan bool, 3)
	chanCb := make(chan any, 10)

	var contextWithCancel context.Context
	var cancel context.CancelFunc
	if watchFunc != nil {
		gl.Log("debug", "Callback function provided, executing...")
		contextWithCancel, cancel = context.WithTimeout(context.Background(), timeout)
		watchFunc(contextWithCancel, chanCb)
	} else {
		gl.Log("debug", "No callback function provided")
		contextWithCancel, cancel = context.WithTimeout(context.Background(), timeout)
	}

	go func(cancel context.CancelFunc, chanErr chan error, chanDone chan bool) {
		defer func(chanErr chan error, chanDone chan bool, chanCb chan any) {
			cancel()
			close(chanErr)
			close(chanDone)
			close(chanCb)
		}(chanErr, chanDone, chanCb)

		gl.Log("debug", "Loading env file...")
		for {
			select {
			case <-contextWithCancel.Done():
				if err := contextWithCancel.Err(); err != nil {
					gl.Log("error", fmt.Sprintf("Error loading env file: %v", err))
					return
				}
				return
			case <-time.After(timeout):
				if chanErr != nil {
					chanErr <- fmt.Errorf("timeout loading env file")
				}
				return
			case <-chanDone:
				gl.Log("debug", "Env file loaded successfully")
				return
			default:
				continue
			}
		}
	}(cancel, chanErr, chanDone)

	// Will add a wait group to wait for the readEnvFile function to finish inside
	// the goroutine, inside the readEnvFile function and wait for the goroutine to finish here.
	go readEnvFile(e, contextWithCancel, e.MuCtxWg)
	e.MuCtxWg.Wait()

	return nil
}

func readEnvFile(e *Environment, ctx context.Context, wg *sync.WaitGroup) {
	if e.GetEnvFilePath() == "" || e.GetEnvFilePath() == ".env" {
		gl.Log("error", "Env file path is empty or default")
		return
	}

	wg.Add(1)
	go func(e *Environment, ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		defer ctx.Done()

		// Read the env file
		fileData, err := os.ReadFile(e.GetEnvFilePath())
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error reading env file: %v", err))
			ctx.Value(fmt.Errorf("error reading env file: %v", err))
			return
		}
		if len(fileData) == 0 {
			gl.Log("error", "Env file is empty")
			ctx.Value(fmt.Errorf("env file is empty"))
			return
		}
		// Check if the env file is encrypted, if so, decrypt it
		isEncrypted := e.IsEncryptedValue(string(fileData))
		if isEncrypted {
			gl.Log("debug", "Env file is encrypted, decrypting...")
			var decryptedData string
			decryptedData, err = e.DecryptEnv(string(fileData))
			if err != nil {
				gl.Log("debug", fmt.Sprintf("Error decrypting env file: %v", err))
				return
			}
			if len(decryptedData) == 0 {
				gl.Log("error", "Decrypted env file is empty")
				return
			}
			fileData = []byte(decryptedData)
			if len(fileData) == 0 {
				gl.Log("error", "Decrypted env file is empty")
				return
			}
		}
		// Create a temp copy of the env file with Mktemp with decrypted data
		tmpFile, err := os.CreateTemp("", "env_*.tmp")
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error creating temp file: %v", err))
			return
		}

		defer func(tmpFile *os.File) {
			gl.Log("debug", "Closing temp file")
			if closeErr := tmpFile.Close(); closeErr != nil {
				gl.Log("error", fmt.Sprintf("Error closing temp file: %v", closeErr.Error()))
				return
			}
			gl.Log("debug", "Removing temp file")
			if err := os.Remove(tmpFile.Name()); err != nil {
				gl.Log("error", fmt.Sprintf("Error removing temp file: %v", err))
			}
		}(tmpFile)

		if _, err := tmpFile.Write(fileData); err != nil {
			gl.Log("error", fmt.Sprintf("Error writing to temp file: %v", err))
			return
		}

		var ext any
		existing := make(map[string]string)
		mapper := NewMapperTypeWithObject(&existing, tmpFile.Name())
		extT, existingErr := mapper.DeserializeFromFile("env")
		if existingErr != nil {
			gl.Log("error", fmt.Sprintf("Error deserializing env file: %v", existingErr))
			return
		}
		if extT == nil {
			gl.Log("error", "Error loading file: nil value")
		} else {
			ext = reflect.ValueOf(extT).Elem().Interface()
		}
		if oldMap, ok := ext.(map[string]string); ok {
			for key, value := range oldMap {
				gl.Log("debug", fmt.Sprintf("Key '%s' value: %s", key, value))
				if setEnvErr := e.Setenv(key, value); setEnvErr != nil {
					gl.Log("error", fmt.Sprintf("Erro ao definir variável de ambiente '%s': %v", key, setEnvErr))
					continue
				}
			}
			e.EnvCache.m = oldMap
			if err := os.Remove(tmpFile.Name()); err != nil {
				gl.Log("error", fmt.Sprintf("Error removing temp file: %v", err))
				return
			}
			gl.Log("debug", "Temp file removed successfully")
			gl.Log("debug", "Env file read successfully")
			return
		} else {
			gl.Log("error", "Error casting to map[string]string")
			return
		}
	}(e, ctx, wg)

	gl.Log("success", "Env file read successfully")
}
func getKey(e *Environment) (sci.ICryptoService, []byte, error) {
	if e.properties["cryptoService"] == nil {
		cryptoService := crp.NewCryptoService()
		if cryptoService == nil {
			gl.Log("error", "Failed to create crypto service")
			return nil, nil, fmt.Errorf("failed to create crypto service")
		}
		if e.properties == nil {
			e.properties = make(map[string]any)
		}
		e.properties["cryptoService"] = NewProperty("cryptoService", &cryptoService, false, nil)
		if e.properties["cryptoService"] == nil {
			return nil, nil, fmt.Errorf("failed to get crypto service")
		}
	}
	cryptoServiceProperty, ok := e.properties["cryptoService"].(ci.IProperty[sci.ICryptoService])
	if !ok {
		gl.Log("error", "Failed to cast crypto service")
		return nil, nil, fmt.Errorf("failed to cast crypto service")
	}
	cryptoService := cryptoServiceProperty.GetValue()
	if e.properties["key"] == nil {
		key, err := cryptoService.GenerateKey()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate key: %v", err)
		}
		e.properties["key"] = NewProperty("key", &key, false, nil)
		if e.properties["key"] == nil {
			return nil, nil, fmt.Errorf("failed to get key")
		}
	}
	key := e.properties["key"].(*Property[[]byte]).GetValue()
	if key == nil {
		gl.Log("error", "Key is nil")
		return nil, nil, fmt.Errorf("key is nil")
	}
	return cryptoService, key, nil
}
