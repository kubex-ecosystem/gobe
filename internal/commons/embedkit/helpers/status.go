// Package helpers defines types used for system information and status embeds.
package helpers

import (
	"strconv"
	"strings"
)

// StatusColor returns a color code based on status
// This is a wrapper to handle with more complex logic if needed
func StatusColor(status string, exitCode int) string {
	var color int
	switch status {
	case "success", "healthy", "ok", "completed", "done", "finished":
		if exitCode == 0 {
			color = getColorCode("success")
		} else {
			color = getColorCode("warning") // warning for success with non-zero exit
		}
	case "failed", "error", "critical", "panic", "aborted", "crashed", "cancelled":
		color = getColorCode("failed")
	case "timeout", "timed_out", "expired", "interrupted", "halted", "stopped":
		color = getColorCode("warning")
	case "running", "in_progress", "pending", "waiting", "starting", "initializing":
		color = getColorCode("running")
	case "warning", "degraded", "unstable", "slow", "lagging", "overloaded":
		color = getColorCode("warning")
	default:
		color = getColorCode("unknown")
	}
	return strconv.Itoa(color)
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

// StatusText returns a human-readable text based on status and sanitizes
// the input to avoid issues with invalid UTF-8 sequences.
func StatusText(status string, exitCode int) string {
	switch strings.ToLower(strings.ToValidUTF8(status, "replacement")) {
	case "success":
		if exitCode == 0 {
			return "Success"
		}
		return "Warning" // warning for success with non-zero exit
	case "failed", "error":
		return "Failed"
	case "timeout":
		return "Timeout"
	case "running", "in_progress":
		return "Running"
	default:
		return "Unknown"
	}
}

// getColorCode maps status to a hex color code
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
