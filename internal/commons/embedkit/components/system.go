// Package components defines types used for system information and status embeds.
package components

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/commons/embedkit/helpers"
)

// SystemInfo represents system information for status embeds
type SystemInfo struct {
	Hostname    string            `json:"hostname"`
	Uptime      time.Duration     `json:"uptime"`
	CPUUsage    float64           `json:"cpu_usage"`
	MemoryUsage float64           `json:"memory_usage"`
	DiskUsage   float64           `json:"disk_usage"`
	Services    map[string]string `json:"services"` // service_name -> status
	Timestamp   time.Time         `json:"timestamp"`
}

// NewStatusEmbedBuilder assembles a status embed builder so callers can enrich it further.
func NewStatusEmbedBuilder(info SystemInfo) *EmbedBuilder {
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

	// Format services status deterministically for readability.
	serviceLines := make([]string, 0, len(info.Services))
	if len(info.Services) > 0 {
		serviceNames := make([]string, 0, len(info.Services))
		for name := range info.Services {
			serviceNames = append(serviceNames, name)
		}
		sort.Strings(serviceNames)
		for _, serviceName := range serviceNames {
			serviceStatus := info.Services[serviceName]
			emoji := helpers.StatusEmoji(serviceStatus, 0)
			serviceLines = append(serviceLines, fmt.Sprintf("%s **%s**: %s", emoji, serviceName, serviceStatus))
		}
	}

	builder := NewEmbedBuilder(fmt.Sprintf("%s System Status - %s", helpers.StatusEmoji(status, 0), info.Hostname)).
		WithColor(helpers.StatusColor(status, 0)).
		WithTimestamp(info.Timestamp).
		WithFooter("System monitoring • Updated")

	builder.AddInlineField("⏱️ Uptime", helpers.FormatDuration(info.Uptime))
	builder.AddInlineField("🖥️ CPU Usage", fmt.Sprintf("%.1f%%", info.CPUUsage))
	builder.AddInlineField("🧠 Memory Usage", fmt.Sprintf("%.1f%%", info.MemoryUsage))
	builder.AddInlineField("💾 Disk Usage", fmt.Sprintf("%.1f%%", info.DiskUsage))
	builder.AddInlineField("📊 Health Score", fmt.Sprintf("%.0f/100", healthScore))

	if len(serviceLines) > 0 {
		builder.AddField("🔧 Services", strings.Join(serviceLines, "\n"), false)
	}

	return builder
}

// StatusEmbed creates a formatted status embed map for compatibility callers.
func StatusEmbed(info SystemInfo) map[string]interface{} {
	return NewStatusEmbedBuilder(info).Build()
}
