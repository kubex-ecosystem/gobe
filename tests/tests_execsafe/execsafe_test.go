package tests_execsafe

import (
	"strings"
	"testing"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/app/security/execsafe"
)

func TestTruncateKB(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		kb          int
		wantTrunc   bool
		wantContent bool // whether output should contain truncation message
	}{
		{
			name:        "small string, no truncation",
			input:       "hello world",
			kb:          1,
			wantTrunc:   false,
			wantContent: false,
		},
		{
			name:        "large string, with truncation",
			input:       strings.Repeat("a", 2000),
			kb:          1,
			wantTrunc:   true,
			wantContent: true,
		},
		{
			name:        "exact limit, no truncation",
			input:       strings.Repeat("b", 1024),
			kb:          1,
			wantTrunc:   false,
			wantContent: false,
		},
		{
			name:        "one byte over limit, with truncation",
			input:       strings.Repeat("c", 1025),
			kb:          1,
			wantTrunc:   true,
			wantContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, truncated := execsafe.TruncateKB(tt.input, tt.kb)

			if truncated != tt.wantTrunc {
				t.Errorf("TruncateKB() truncated = %v, want %v", truncated, tt.wantTrunc)
			}

			if tt.wantContent && !strings.Contains(result, "…(truncated)…") {
				t.Error("TruncateKB() should contain truncation message but doesn't")
			}

			if !tt.wantContent && strings.Contains(result, "…(truncated)…") {
				t.Error("TruncateKB() shouldn't contain truncation message but does")
			}

			// Check that we don't exceed the limit (plus truncation message)
			maxExpectedLen := tt.kb*1024 + 20 // 20 chars for truncation message
			if len(result) > maxExpectedLen {
				t.Errorf("TruncateKB() result too long: got %d, max expected %d", len(result), maxExpectedLen)
			}
		})
	}
}

func TestSanitizeOneLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "newlines replaced with spaces",
			input:    "hello\nworld\ntest",
			expected: "hello world test",
		},
		{
			name:     "multiple spaces collapsed",
			input:    "hello   world    test",
			expected: "hello world test",
		},
		{
			name:     "leading and trailing spaces trimmed",
			input:    "   hello world   ",
			expected: "hello world",
		},
		{
			name:     "mixed whitespace normalized",
			input:    "  hello\n\n  world\t\ttest   \n",
			expected: "hello world test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := execsafe.SanitizeOneLine(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeOneLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseUserCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantArgs    []string
		wantErr     bool
		description string
	}{
		{
			name:        "execute command in English",
			input:       "execute ls -la",
			wantName:    "ls",
			wantArgs:    []string{"-la"},
			wantErr:     false,
			description: "basic execute command",
		},
		{
			name:        "executar command in Portuguese",
			input:       "executar pwd",
			wantName:    "pwd",
			wantArgs:    []string{},
			wantErr:     false,
			description: "Portuguese execute command",
		},
		{
			name:        "run command",
			input:       "run date",
			wantName:    "date",
			wantArgs:    []string{},
			wantErr:     false,
			description: "run command variant",
		},
		{
			name:        "exec with arguments",
			input:       "exec ps -aux",
			wantName:    "ps",
			wantArgs:    []string{"-aux"},
			wantErr:     false,
			description: "exec with arguments",
		},
		{
			name:        "command with quoted arguments",
			input:       "execute echo \"hello world\"",
			wantName:    "echo",
			wantArgs:    []string{"hello world"},
			wantErr:     false,
			description: "command with quoted arguments",
		},
		{
			name:        "no command found",
			input:       "just some text without command",
			wantName:    "",
			wantArgs:    nil,
			wantErr:     true,
			description: "no trigger word found",
		},
		{
			name:        "command with shell metacharacters",
			input:       "execute ls | grep test",
			wantName:    "",
			wantArgs:    nil,
			wantErr:     true,
			description: "shell metacharacters should be rejected",
		},
		{
			name:        "unclosed quote",
			input:       "execute echo \"unclosed quote",
			wantName:    "",
			wantArgs:    nil,
			wantErr:     true,
			description: "unclosed quote should cause error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := execsafe.ParseUserCommand(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseUserCommand() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseUserCommand() unexpected error: %v", err)
				return
			}

			if result.Name != tt.wantName {
				t.Errorf("ParseUserCommand() name = %q, want %q", result.Name, tt.wantName)
			}

			if len(result.Args) != len(tt.wantArgs) {
				t.Errorf("ParseUserCommand() args length = %d, want %d", len(result.Args), len(tt.wantArgs))
			}

			for i, arg := range result.Args {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("ParseUserCommand() args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestExtractShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "execute command",
			input:    "please execute ls -la",
			expected: "ls -la",
		},
		{
			name:     "executar in Portuguese",
			input:    "por favor executar pwd",
			expected: "pwd",
		},
		{
			name:     "run command",
			input:    "can you run date for me?",
			expected: "date for me?",
		},
		{
			name:     "exec command",
			input:    "exec ps aux",
			expected: "ps aux",
		},
		{
			name:     "no command trigger",
			input:    "just some normal text",
			expected: "",
		},
		{
			name:     "case insensitive",
			input:    "Please EXECUTE ls",
			expected: "ls",
		},
		{
			name:     "multiple spaces normalized",
			input:    "execute   ls    -la",
			expected: "ls -la",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := execsafe.ExtractShellCommand(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractShellCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRegistryOperations(t *testing.T) {
	registry := execsafe.NewRegistry()

	// Test registration
	spec := execsafe.CommandSpec{
		Binary:      "test-binary",
		Timeout:     5 * time.Second,
		MaxOutputKB: 100,
	}

	registry.Register("test", spec)

	// Test retrieval
	retrievedSpec, ok := registry.Get("test")
	if !ok {
		t.Error("Registry.Get() should return true for registered command")
	}

	if retrievedSpec.Binary != spec.Binary {
		t.Errorf("Registry.Get() binary = %q, want %q", retrievedSpec.Binary, spec.Binary)
	}

	// Test case insensitive retrieval
	retrievedSpec2, ok2 := registry.Get("TEST")
	if !ok2 {
		t.Error("Registry.Get() should be case insensitive")
	}

	if retrievedSpec2.Binary != spec.Binary {
		t.Error("Registry.Get() case insensitive should return same spec")
	}

	// Test non-existent command
	_, ok3 := registry.Get("nonexistent")
	if ok3 {
		t.Error("Registry.Get() should return false for non-existent command")
	}
}