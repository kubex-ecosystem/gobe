package helpers

import (
	"fmt"
	"time"
)

// ExecResult represents command execution result
type ExecResult struct {
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	Stdout    string        `json:"stdout"`
	Stderr    string        `json:"stderr"`
	Truncated bool          `json:"truncated"`
	Status    string        `json:"status"`
	UserID    string        `json:"user_id"`
	Channel   string        `json:"channel"`
}

// ExecResultEmbed creates a formatted execution result embed
func ExecResultEmbed(result ExecResult) map[string]interface{} {
	status := result.Status
	if status == "" {
		if result.ExitCode == 0 {
			status = "success"
		} else {
			status = "failed"
		}
	}

	// Build command line
	cmdLine := result.Command
	if len(result.Args) > 0 {
		for _, arg := range result.Args {
			// Simple quoting for display
			if containsSpace(arg) {
				cmdLine += fmt.Sprintf(" \"%s\"", arg)
			} else {
				cmdLine += " " + arg
			}
		}
	}

	// Format output sections
	fields := []map[string]interface{}{
		{
			"name":   "⚡ Command",
			"value":  fmt.Sprintf("```bash\n%s\n```", cmdLine),
			"inline": false,
		},
		{
			"name":   "⏱️ Duration",
			"value":  FormatDuration(result.Duration),
			"inline": true,
		},
		{
			"name":   "🚪 Exit Code",
			"value":  fmt.Sprintf("%d", result.ExitCode),
			"inline": true,
		},
	}

	// Add stdout if present
	if result.Stdout != "" {
		stdout := result.Stdout
		if len(stdout) > 1000 {
			stdout = stdout[:1000] + "..."
		}
		fields = append(fields, map[string]interface{}{
			"name":  "📤 Output",
			"value": fmt.Sprintf("```\n%s\n```", stdout),
		})
	}

	// Add stderr if present
	if result.Stderr != "" {
		stderr := result.Stderr
		if len(stderr) > 1000 {
			stderr = stderr[:1000] + "..."
		}
		fields = append(fields, map[string]interface{}{
			"name":  "⚠️ Error Output",
			"value": fmt.Sprintf("```\n%s\n```", stderr),
		})
	}

	// Add truncation warning
	if result.Truncated {
		fields = append(fields, map[string]interface{}{
			"name":  "⚠️ Notice",
			"value": "Output was truncated due to size limits",
		})
	}

	embed := map[string]interface{}{
		"title":  fmt.Sprintf("%s Command Execution", StatusEmoji(status, result.ExitCode)),
		"color":  StatusColor(status, result.ExitCode),
		"fields": fields,
		"footer": map[string]interface{}{
			"text": fmt.Sprintf("Executed by %s in %s", result.UserID, result.Channel),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return embed
}

// Helper functions

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' {
			return true
		}
	}
	return false
}
