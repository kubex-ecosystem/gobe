// Package interfaces defines the IEnvironment interface for managing environment variables and system information.
package interfaces

import "context"

type IEnvironment interface {
	Mu() IMutexes
	CPUCount() int
	MemTotal() int
	Hostname() string
	Os() string
	Kernel() string
	LoadEnvFile(watchFunc func(ctx context.Context, chanCbArg chan any) <-chan any) error
	GetEnvFilePath() string
	Getenv(key string) string
	Setenv(key, value string) error
	GetEnvCache() map[string]string
	ParseEnvVar(s string) (string, string)
	LoadEnvFromShell() error
	MemAvailable() int
	GetShellName(s string) (string, int)
	BackupEnvFile() error
	EncryptEnvFile() error
	DecryptEnvFile() (string, error)
	EncryptEnv(value string) (string, error)
	DecryptEnv(encryptedValue string) (string, error)
	IsEncrypted(envFile string) bool
	IsEncryptedValue(value string) bool
	EnableEnvFileEncryption() error
	DisableEnvFileEncryption() error
}
