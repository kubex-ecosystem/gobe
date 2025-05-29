package tests

// import (
// 	"context"
// 	"fmt"
// 	gb "github.com/rafa-mori/gobe"
// 	ci "github.com/rafa-mori/gobe/internal/interfaces"
// 	at "github.com/rafa-mori/gobe/internal/types"
// 	gl "github.com/rafa-mori/gobe/logger"
// 	l "github.com/faelmori/logz"
// 	"os"
// 	"testing"
// 	"time"
// )

// var bgmErr error

// func getGoBEInstance(logFile, configFile string, isConfidential bool) (ci.IGoBE, error) {
// 	return gb.NewGoBE(
// 		"Test",
// 		"8666",
// 		"0.0.0.0",
// 		logFile,
// 		configFile,
// 		isConfidential,
// 		l.GetLogger("GoBE"),
// 		false,
// 	)
// }

// func TestLoadEnvFileWithValidFileLoadsSuccessfully(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/test.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"

// 	gbmT, bgmErr = getGoBEInstance(logFile, envFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}

// 	if gbmT == nil {
// 		gl.Log("fatal", "GoBE instance is nil")
// 	}

// 	value := gbmT.Environment().Getenv("KEY")
// 	if value != "VALUE" {
// 		gl.Log("error", fmt.Sprintf("expected 'KEY' to be 'VALUE', got '%s'", value))
// 		t.Fatalf("expected 'KEY' to be 'VALUE', got '%s'", value)
// 	}
// }

// func TestLoadEnvFileWithValidEncryptedFileLoadsSuccessfully(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/go_kubex.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"

// 	gbmT, bgmErr = getGoBEInstance(logFile, envFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if gbmT == nil {
// 		gl.Log("fatal", "GoBE instance is nil")
// 	}

// 	expect := "faelmori@gmail.com"

// 	value := gbmT.Environment().Getenv("EMAIL_USR")

// 	gl.Log("notice", fmt.Sprintf("Expecting 'EMAIL_USR' to be '%s'", expect))
// 	gl.Log("notice", fmt.Sprintf("Config file: '%s'", gbmT.Environment().GetEnvFilePath()))
// 	gl.Log("notice", fmt.Sprintf("Is encrypted: '%t'", gbmT.Environment().IsEncrypted(gbmT.Environment().GetEnvFilePath())))
// 	gl.Log("notice", fmt.Sprintf("%d env vars loaded", len(gbmT.Environment().GetEnvCache())))
// 	gl.Log("notice", fmt.Sprintf("Env var 'EMAIL_USR': '%s'", value))

// 	if value != expect {
// 		gl.Log("warn", fmt.Sprintf("expected 'EMAIL_USR' to be '%s', got '%s'", expect, gbmT.Environment().Getenv("EMAIL_USR")))
// 		t.Fatalf("expected 'EMAIL_USR' to be '%s', got '%s'", expect, value)
// 	}
// }

// func TestLoadEnvFileWithInvalidFileReturnsError(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/nonexistent.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, false)
// 	if bgmErr != nil {
// 		if bgmErr.Error() != "environment is nil" {
// 			t.Fatalf("expected 'environment is nil', got '%v'", bgmErr.Error())
// 		} else {
// 			t.Logf("expected 'environment is nil', got '%v'", bgmErr.Error())
// 			return
// 		}
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		if err.Error() != "environment is nil" {
// 			t.Fatalf("expected 'environment is nil', got '%v'", err.Error())
// 		} else {
// 			t.Logf("expected 'environment is nil', got '%v'", err.Error())
// 			return
// 		}
// 	}

// 	err := gbm.Environment().LoadEnvFile(nil)

// 	if err == nil {
// 		t.Fatal("expected an error, got nil")
// 	}
// }

// func TestEncryptEnvFileWithUnencryptedFileEncryptsSuccessfully(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/unencrypted.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}

// 	writeErr := os.WriteFile(envFile, []byte("KEY=VALUE\n"), 0644)
// 	if writeErr != nil {
// 		t.Fatalf("failed to write test file: %v", writeErr)
// 		return
// 	}
// 	defer func(name string) {
// 		_ = os.Remove(name)
// 	}(envFile)

// 	err := gbm.Environment().EncryptEnvFile()
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	if !gbm.Environment().IsEncrypted(envFile) {
// 		t.Errorf("expected file to be encrypted")
// 	}
// }

// func TestEncryptEnvFileWithAlreadyEncryptedFileDoesNothing(t *testing.T) {
// 	envFile := "encrypted.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}

// 	writeErr := os.WriteFile(envFile, []byte("KEY=VALUE\n"), 0644)
// 	if writeErr != nil {
// 		t.Fatalf("failed to write test file: %v", writeErr)
// 		return
// 	}
// 	defer func(name string) {
// 		_ = os.Remove(name)
// 	}(envFile)

// 	encryptErr := gbm.Environment().EncryptEnvFile()
// 	if encryptErr != nil {
// 		t.Fatalf("unexpected error: %v", encryptErr)
// 		return
// 	} // Encrypt once
// 	err := gbm.Environment().EncryptEnvFile() // Encrypt again
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	if !gbm.Environment().IsEncrypted(envFile) {
// 		t.Errorf("expected file to remain encrypted")
// 	}
// }

// func TestDecryptEnvWithValidEncryptedValueReturnsOriginalValue(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/go_kubex.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}
// 	originalValue := "VALUE"
// 	encryptedValue, err := gbm.Environment().EncryptEnv(originalValue)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	decryptedValue, err := gbm.Environment().DecryptEnv(encryptedValue)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	if decryptedValue != originalValue {
// 		t.Errorf("expected '%s', got '%s'", originalValue, decryptedValue)
// 	}
// }

// func TestDecryptEnvWithInvalidEncryptedValueReturnsError(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/go_kubex.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}
// 	_, err := gbm.Environment().DecryptEnv("invalid-encrypted-value")
// 	if err == nil {
// 		t.Fatal("expected an error, got nil")
// 	}
// }

// func TestLoadEnvFromShellLoadsEnvironmentVariables(t *testing.T) {
// 	envFile := "/srv/apps/projects/gobe/tests/env/go_kubex.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}
// 	env, err := at.NewEnvironment("", false, nil)
// 	if err != nil {
// 		t.Fatalf("failed to create environment: %v", err)
// 	}
// 	err = env.LoadEnvFromShell()

// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	if env.Getenv("PATH") == "" {
// 		t.Errorf("expected 'PATH' to be set, got empty value")
// 	}
// }

// func TestBackupEnvFileCreatesBackupSuccessfully(t *testing.T) {
// 	envFile := "test.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}

// 	writeErr := os.WriteFile(envFile, []byte("KEY=VALUE\n"), 0644)
// 	if writeErr != nil {
// 		t.Fatalf("failed to write test file: %v", writeErr)
// 		return
// 	}
// 	defer func(name string) {
// 		_ = os.Remove(name)
// 	}(envFile)

// 	err := gbm.Environment().BackupEnvFile()
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	backupFile := envFile + ".backup"
// 	defer func(name string) {
// 		_ = os.Remove(name)
// 	}(backupFile)

// 	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
// 		t.Errorf("expected backup file to exist")
// 	}
// }

// func TestLoadEnvFileWithTimeoutReturnsError(t *testing.T) {
// 	envFile := "test.env"
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := envFile

// 	gbm, bgmErr := getGoBEInstance(logFile, configFile, true)
// 	if bgmErr != nil {
// 		gl.Log("fatal", fmt.Sprintf("Failed to initialize GoBE: %v", bgmErr.Error()))
// 		return
// 	}
// 	if err := gbm.Initialize(); err != nil {
// 		t.Fatalf("failed to initialize: %v", err)
// 	}

// 	writeErr := os.WriteFile(envFile, []byte("KEY=VALUE\n"), 0644)
// 	if writeErr != nil {
// 		t.Fatalf("failed to write test file: %v", writeErr)
// 		return
// 	}
// 	defer func(name string) {
// 		_ = os.Remove(name)
// 	}(envFile)

// 	env, err := at.NewEnvironment(envFile, false, nil)
// 	_, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
// 	defer cancel()

// 	err = env.LoadEnvFile(func(ctx context.Context, chanCbArg chan any) <-chan any {
// 		time.Sleep(2 * time.Millisecond)
// 		return nil
// 	})

// 	if err == nil {
// 		t.Fatal("expected an error, got nil")
// 	}
// }
