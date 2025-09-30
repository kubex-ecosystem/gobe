// Package testsembedkit contains unit tests for the embedkit package.
package testsembedkit

import (
	"testing"
	"time"

	embedkit "github.com/kubex-ecosystem/gobe/internal/commons/embedkit/components"
	"github.com/kubex-ecosystem/gobe/internal/commons/embedkit/helpers"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			want:     "500ms",
		},
		{
			name:     "less than millisecond",
			duration: 100 * time.Microsecond,
			want:     "0ms",
		},
		{
			name:     "seconds",
			duration: 2500 * time.Millisecond,
			want:     "2.5s",
		},
		{
			name:     "minutes",
			duration: 90 * time.Second,
			want:     "1.5m",
		},
		{
			name:     "hours",
			duration: 2*time.Hour + 30*time.Minute,
			want:     "2.5h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpers.FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		exitCode int
		want     string
	}{
		{
			name:     "success with exit 0",
			status:   "success",
			exitCode: 0,
			want:     "65280",
		},
		{
			name:     "success with non-zero exit",
			status:   "success",
			exitCode: 1,
			want:     "16776960",
		},
		{
			name:     "failed status",
			status:   "failed",
			exitCode: 1,
			want:     "16711680",
		},
		{
			name:     "error status",
			status:   "error",
			exitCode: 2,
			want:     "16711680",
		},
		{
			name:     "timeout status",
			status:   "timeout",
			exitCode: 124,
			want:     "16776960",
		},
		{
			name:     "running status",
			status:   "running",
			exitCode: 0,
			want:     "39423",
		},
		{
			name:     "unknown status",
			status:   "unknown",
			exitCode: 0,
			want:     "8421504",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpers.StatusColor(tt.status, tt.exitCode)
			if got != tt.want {
				t.Errorf("StatusColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusEmoji(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		exitCode int
		want     string
	}{
		{
			name:     "success with exit 0",
			status:   "success",
			exitCode: 0,
			want:     "✅",
		},
		{
			name:     "success with non-zero exit",
			status:   "success",
			exitCode: 1,
			want:     "⚠️",
		},
		{
			name:     "failed status",
			status:   "failed",
			exitCode: 1,
			want:     "❌",
		},
		{
			name:     "timeout status",
			status:   "timeout",
			exitCode: 124,
			want:     "⏰",
		},
		{
			name:     "running status",
			status:   "running",
			exitCode: 0,
			want:     "⏳",
		},
		{
			name:     "unknown status",
			status:   "unknown",
			exitCode: 0,
			want:     "ℹ️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpers.StatusEmoji(tt.status, tt.exitCode)
			if got != tt.want {
				t.Errorf("StatusEmoji() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLinkButtons(t *testing.T) {
	tests := []struct {
		name  string
		links map[string]string
		want  string
	}{
		{
			name:  "empty links",
			links: map[string]string{},
			want:  "",
		},
		{
			name: "single link",
			links: map[string]string{
				"GitHub": "https://github.com",
			},
			want: "[GitHub](https://github.com)",
		},
		{
			name: "multiple links",
			links: map[string]string{
				"GitHub": "https://github.com",
				"Docs":   "https://docs.example.com",
			},
			want: "[GitHub](https://github.com) | [Docs](https://docs.example.com)",
		},
		{
			name:  "nil links",
			links: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embedkit.NewLinkButtons(tt.links)
			// Note: map iteration order is not guaranteed, so we check both possible orders for multiple links
			if tt.name == "multiple links" {
				alt := "[Docs](https://docs.example.com) | [GitHub](https://github.com)"
				if got != tt.want && got != alt {
					t.Errorf("NewLinkButtons() = %v, want %v or %v", got, tt.want, alt)
				}
			} else if got != tt.want {
				t.Errorf("NewLinkButtons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusEmbed(t *testing.T) {
	info := embedkit.SystemInfo{
		Hostname:    "test-server",
		Uptime:      2*time.Hour + 30*time.Minute,
		CPUUsage:    45.5,
		MemoryUsage: 67.8,
		DiskUsage:   23.1,
		Services: map[string]string{
			"database": "running",
			"web":      "success",
			"cache":    "failed",
		},
		Timestamp: time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC),
	}

	embed := embedkit.StatusEmbed(info)

	// Check that required fields exist
	if embed["title"] == nil {
		t.Error("StatusEmbed() missing title field")
	}

	if embed["color"] == nil {
		t.Error("StatusEmbed() missing color field")
	}

	if embed["fields"] == nil {
		t.Error("StatusEmbed() missing fields field")
	}

	// Check title contains hostname
	title, ok := embed["title"].(string)
	if !ok {
		t.Error("StatusEmbed() title is not a string")
	} else if !containsString(title, info.Hostname) {
		t.Errorf("StatusEmbed() title %q should contain hostname %q", title, info.Hostname)
	}

	// Check fields structure
	fields, ok := embed["fields"].([]map[string]interface{})
	if !ok {
		t.Error("StatusEmbed() fields is not the expected type")
	} else if len(fields) == 0 {
		t.Error("StatusEmbed() should have at least one field")
	}

	// Verify timestamp format
	timestamp, ok := embed["timestamp"].(string)
	if !ok {
		t.Error("StatusEmbed() timestamp should be a string")
	} else {
		_, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("StatusEmbed() timestamp %q is not valid RFC3339: %v", timestamp, err)
		}
	}
}

func TestExecResultEmbed(t *testing.T) {
	result := helpers.ExecResult{
		Command:   "ls",
		Args:      []string{"-la", "/tmp"},
		ExitCode:  0,
		Duration:  150 * time.Millisecond,
		Stdout:    "total 0\ndrwxrwxrwt 2 root root 40 Dec 25 12:00 .\ndrwxr-xr-x 3 root root 60 Dec 25 11:59 ..",
		Stderr:    "",
		Truncated: false,
		Status:    "success",
		UserID:    "user123",
		Channel:   "#general",
	}

	embed := helpers.ExecResultEmbed(result)

	// Check that required fields exist
	if embed["title"] == nil {
		t.Error("ExecResultEmbed() missing title field")
	}

	if embed["color"] == nil {
		t.Error("ExecResultEmbed() missing color field")
	}

	if embed["fields"] == nil {
		t.Error("ExecResultEmbed() missing fields field")
	}

	// Check fields structure
	fields, ok := embed["fields"].([]map[string]interface{})
	if !ok {
		t.Error("ExecResultEmbed() fields is not the expected type")
	}

	// Should have at least command, duration, and exit code fields
	if len(fields) < 3 {
		t.Errorf("ExecResultEmbed() should have at least 3 fields, got %d", len(fields))
	}

	// Check footer contains user info
	footer, ok := embed["footer"].(map[string]interface{})
	if !ok {
		t.Error("ExecResultEmbed() footer should be a map")
	} else {
		footerText, ok := footer["text"].(string)
		if !ok {
			t.Error("ExecResultEmbed() footer text should be a string")
		} else if !containsString(footerText, result.UserID) || !containsString(footerText, result.Channel) {
			t.Errorf("ExecResultEmbed() footer %q should contain user and channel info", footerText)
		}
	}
}

func TestExecResultEmbedWithError(t *testing.T) {
	result := helpers.ExecResult{
		Command:   "nonexistent-cmd",
		Args:      []string{},
		ExitCode:  127,
		Duration:  50 * time.Millisecond,
		Stdout:    "",
		Stderr:    "command not found: nonexistent-cmd",
		Truncated: false,
		Status:    "failed",
		UserID:    "user456",
		Channel:   "#test",
	}

	embed := helpers.ExecResultEmbed(result)

	// Check fields include stderr
	fields, ok := embed["fields"].([]map[string]interface{})
	if !ok {
		t.Error("ExecResultEmbed() fields is not the expected type")
		return
	}

	// Look for stderr field
	hasStderrField := false
	for _, field := range fields {
		if name, ok := field["name"].(string); ok && containsString(name, "Error") {
			hasStderrField = true
			if value, ok := field["value"].(string); ok && !containsString(value, result.Stderr) {
				t.Errorf("ExecResultEmbed() stderr field should contain error message")
			}
			break
		}
	}

	if !hasStderrField {
		t.Error("ExecResultEmbed() should include stderr field when stderr is present")
	}
}

func TestExecResultEmbedWithTruncation(t *testing.T) {
	result := helpers.ExecResult{
		Command:   "cat",
		Args:      []string{"largefile.txt"},
		ExitCode:  0,
		Duration:  2 * time.Second,
		Stdout:    "some output",
		Stderr:    "",
		Truncated: true,
		Status:    "success",
		UserID:    "user789",
		Channel:   "#ops",
	}

	embed := helpers.ExecResultEmbed(result)

	fields, ok := embed["fields"].([]map[string]interface{})
	if !ok {
		t.Error("ExecResultEmbed() fields is not the expected type")
		return
	}

	// Look for truncation notice
	hasTruncationField := false
	for _, field := range fields {
		if name, ok := field["name"].(string); ok && containsString(name, "Notice") {
			hasTruncationField = true
			if value, ok := field["value"].(string); ok && !containsString(value, "truncated") {
				t.Errorf("ExecResultEmbed() truncation field should mention truncation")
			}
			break
		}
	}

	if !hasTruncationField {
		t.Error("ExecResultEmbed() should include truncation notice when truncated is true")
	}
}

// Helper function for tests
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && haystack[:len(needle)] == needle ||
		len(haystack) > len(needle) && haystack[len(haystack)-len(needle):] == needle ||
		containsSubstring(haystack, needle)
}

func containsSubstring(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
