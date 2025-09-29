// Package embedkit provides helpers for creating rich embeds and formatting
package embedkit

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// StatusColor returns a color code based on status
func StatusColor(status string, exitCode int) string {
	switch status {
	case "success":
		if exitCode == 0 {
			return "🟢" // green
		}
		return "🟡" // yellow for success with non-zero exit
	case "failed", "error":
		return "🔴" // red
	case "timeout":
		return "🟡" // yellow
	case "running", "in_progress":
		return "🔵" // blue
	default:
		return "⚪" // white/neutral
	}
}

// StatusEmoji returns an emoji based on status
func StatusEmoji(status string, exitCode int) string {
	switch status {
	case "success":
		if exitCode == 0 {
			return "✅"
		}
		return "⚠️" // warning for success with non-zero exit
	case "failed", "error":
		return "❌"
	case "timeout":
		return "⏰"
	case "running", "in_progress":
		return "⏳"
	default:
		return "ℹ️"
	}
}

// NewLinkButtons creates formatted link buttons for embeds
func NewLinkButtons(links map[string]string) string {
	if len(links) == 0 {
		return ""
	}

	var buttons string
	for label, url := range links {
		if buttons != "" {
			buttons += " | "
		}
		buttons += fmt.Sprintf("[%s](%s)", label, url)
	}
	return buttons
}

// SystemInfo represents system information for status embeds
type SystemInfo struct {
	Hostname    string            `json:"hostname"`
	Uptime      time.Duration     `json:"uptime"`
	CPUUsage    float64          `json:"cpu_usage"`
	MemoryUsage float64          `json:"memory_usage"`
	DiskUsage   float64          `json:"disk_usage"`
	Services    map[string]string `json:"services"` // service_name -> status
	Timestamp   time.Time         `json:"timestamp"`
}

// StatusEmbed creates a formatted status embed
func StatusEmbed(info SystemInfo) map[string]interface{} {
	status := "healthy"
	if info.CPUUsage > 80 || info.MemoryUsage > 90 || info.DiskUsage > 95 {
		status = "warning"
	}

	// Calculate overall health score
	healthScore := 100.0
	if info.CPUUsage > 50 {
		healthScore -= (info.CPUUsage - 50) * 0.5
	}
	if info.MemoryUsage > 70 {
		healthScore -= (info.MemoryUsage - 70) * 1.0
	}
	if info.DiskUsage > 80 {
		healthScore -= (info.DiskUsage - 80) * 2.0
	}

	if healthScore < 50 {
		status = "critical"
	} else if healthScore < 75 {
		status = "warning"
	}

	// Format services status
	servicesText := ""
	for serviceName, serviceStatus := range info.Services {
		emoji := StatusEmoji(serviceStatus, 0)
		if servicesText != "" {
			servicesText += "\n"
		}
		servicesText += fmt.Sprintf("%s **%s**: %s", emoji, serviceName, serviceStatus)
	}

	embed := map[string]interface{}{
		"title": fmt.Sprintf("%s System Status - %s", StatusEmoji(status, 0), info.Hostname),
		"color": getColorCode(status),
		"fields": []map[string]interface{}{
			{
				"name":   "⏱️ Uptime",
				"value":  FormatDuration(info.Uptime),
				"inline": true,
			},
			{
				"name":   "🖥️ CPU Usage",
				"value":  fmt.Sprintf("%.1f%%", info.CPUUsage),
				"inline": true,
			},
			{
				"name":   "🧠 Memory Usage",
				"value":  fmt.Sprintf("%.1f%%", info.MemoryUsage),
				"inline": true,
			},
			{
				"name":   "💾 Disk Usage",
				"value":  fmt.Sprintf("%.1f%%", info.DiskUsage),
				"inline": true,
			},
			{
				"name":   "📊 Health Score",
				"value":  fmt.Sprintf("%.0f/100", healthScore),
				"inline": true,
			},
			{
				"name":  "🔧 Services",
				"value": servicesText,
			},
		},
		"timestamp": info.Timestamp.Format(time.RFC3339),
		"footer": map[string]interface{}{
			"text": "System monitoring • Updated",
		},
	}

	return embed
}

// ExecResult represents command execution result
type ExecResult struct {
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	ExitCode  int          `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	Stdout    string       `json:"stdout"`
	Stderr    string       `json:"stderr"`
	Truncated bool         `json:"truncated"`
	Status    string       `json:"status"`
	UserID    string       `json:"user_id"`
	Channel   string       `json:"channel"`
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
		"color":  getColorCode(status),
		"fields": fields,
		"footer": map[string]interface{}{
			"text": fmt.Sprintf("Executed by %s in %s", result.UserID, result.Channel),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return embed
}

// Helper functions

func getColorCode(status string) int {
	switch status {
	case "success", "healthy":
		return 0x00ff00 // green
	case "warning":
		return 0xffff00 // yellow
	case "failed", "error", "critical":
		return 0xff0000 // red
	case "running", "in_progress":
		return 0x0099ff // blue
	default:
		return 0x808080 // gray
	}
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' {
			return true
		}
	}
	return false
}