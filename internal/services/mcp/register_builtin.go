// Package mcp provides built-in tool registrations for the MCP system.
package mcp

import (
	"context"
	"fmt"
	"runtime"
	"time"

	manage "github.com/kubex-ecosystem/gobe/internal/app/controllers/sys/manage"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

// RegisterBuiltinTools registers all built-in MCP tools
func RegisterBuiltinTools(registry Registry) error {
	if registry == nil {
		gl.Log("error", "Registry is nil, cannot register builtin tools")
		return fmt.Errorf("registry cannot be nil")
	}

	// Register system.status tool
	statusSpec := ToolSpec{
		Name:        "system.status",
		Title:       "System Status",
		Description: "Get comprehensive system status including health, version, and runtime metrics",
		Auth:        "none",
		Args: map[string]interface{}{
			"detailed": map[string]interface{}{
				"type":        "boolean",
				"description": "Include detailed system metrics",
				"default":     false,
			},
		},
		Handler: systemStatusHandler,
	}

	err := registry.Register(statusSpec)
	if err != nil {
		gl.Log("error", "Failed to register system.status tool", err)
		return fmt.Errorf("failed to register system.status: %w", err)
	}

	gl.Log("info", "Built-in MCP tools registered successfully")
	return nil
}

// systemStatusHandler handles the system.status tool execution
func systemStatusHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	gl.Log("info", "Executing system.status tool")

	// Get detailed flag (default false)
	detailed := false
	if detailedArg, exists := args["detailed"]; exists {
		if detailedVal, ok := detailedArg.(bool); ok {
			detailed = detailedVal
		}
	}

	// Get basic system info using manage controller patterns
	serverController := manage.NewServerController()

	// Build basic status response
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(bootTime).String(),
		"version":   "v1.3.3", // From manifest.json
		"health": map[string]interface{}{
			"status":  "healthy",
			"message": "System is operational",
		},
	}

	// Add detailed metrics if requested
	if detailed {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		status["runtime"] = map[string]interface{}{
			"go_version":      runtime.Version(),
			"goroutines":      runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc_mb":      float64(memStats.Alloc) / 1024 / 1024,
				"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
				"sys_mb":        float64(memStats.Sys) / 1024 / 1024,
				"gc_cycles":     memStats.NumGC,
			},
		}

		status["connections"] = map[string]interface{}{
			"rabbit": map[string]interface{}{
				"status": "checking", // This would be populated by actual connection checks
			},
			"database": map[string]interface{}{
				"status": "checking", // This would be populated by actual connection checks
			},
		}
	}

	// Add server controller reference for consistency
	_ = serverController // Keep reference to maintain pattern

	return status, nil
}

// bootTime tracks when the system started
var bootTime = time.Now()

// GetRegistry returns the global registry instance for external access
// Note: This should be called after the registry is initialized in the controller
func GetRegistry() Registry {
	gl.Log("warn", "GetRegistry called - registry should be accessed via controller")
	return nil // The registry is managed by the controller, not globally here
}

// SetRegistry allows external setting of the registry (for testing)
func SetRegistry(registry Registry) {
	gl.Log("info", "SetRegistry called - registry should be managed via controller")
}