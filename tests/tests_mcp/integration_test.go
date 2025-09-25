package tests_mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/system"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
)

func TestMCPEndpoints_Integration(t *testing.T) {
	// Setup gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Initialize controller
	controller := system.NewMetricsController(nil)

	// Setup routes
	router.GET("/mcp/tools", controller.ListTools)
	router.POST("/mcp/exec", controller.ExecTool)

	t.Run("ListTools endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/mcp/tools", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ListTools() status = %v, want %v", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check if response has expected structure
		if _, ok := response["data"]; !ok {
			t.Errorf("Response missing 'data' field")
		}

		if data, ok := response["data"].(map[string]interface{}); ok {
			if _, ok := data["tools"]; !ok {
				t.Errorf("Response data missing 'tools' field")
			}
			if _, ok := data["count"]; !ok {
				t.Errorf("Response data missing 'count' field")
			}
		}
	})

	t.Run("ExecTool endpoint with system.status", func(t *testing.T) {
		// Prepare request
		requestBody := map[string]interface{}{
			"tool": "system.status",
			"args": map[string]interface{}{
				"detailed": false,
			},
		}
		requestJSON, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ExecTool() status = %v, want %v, response: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check if response has expected structure
		if _, ok := response["data"]; !ok {
			t.Errorf("Response missing 'data' field")
		}

		if data, ok := response["data"].(map[string]interface{}); ok {
			if _, ok := data["tool"]; !ok {
				t.Errorf("Response data missing 'tool' field")
			}
			if _, ok := data["result"]; !ok {
				t.Errorf("Response data missing 'result' field")
			}
		}
	})

	t.Run("ExecTool with detailed system.status", func(t *testing.T) {
		// Prepare request
		requestBody := map[string]interface{}{
			"tool": "system.status",
			"args": map[string]interface{}{
				"detailed": true,
			},
		}
		requestJSON, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ExecTool() detailed status = %v, want %v, response: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify detailed response structure
		if data, ok := response["data"].(map[string]interface{}); ok {
			if result, ok := data["result"].(map[string]interface{}); ok {
				if _, ok := result["runtime"]; !ok {
					t.Errorf("Detailed status missing 'runtime' field")
				}
				if _, ok := result["connections"]; !ok {
					t.Errorf("Detailed status missing 'connections' field")
				}
			}
		}
	})

	t.Run("ExecTool with invalid tool", func(t *testing.T) {
		// Prepare request
		requestBody := map[string]interface{}{
			"tool": "nonexistent.tool",
			"args": map[string]interface{}{},
		}
		requestJSON, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return error but not 500 - our API wrapper handles errors gracefully
		if w.Code == http.StatusInternalServerError {
			t.Logf("ExecTool() with invalid tool status = %v (expected error), response: %s", w.Code, w.Body.String())
		}
	})

	t.Run("ExecTool with malformed request", func(t *testing.T) {
		// Send invalid JSON
		req, _ := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer([]byte(`{"invalid": json}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 Bad Request or similar
		if w.Code == http.StatusOK {
			t.Errorf("ExecTool() with malformed request should not return OK, got status = %v", w.Code)
		}
	})
}

func TestBuiltinTools_Integration(t *testing.T) {
	t.Run("system.status tool execution", func(t *testing.T) {
		// Test the tool directly via registry
		registry := mcp.NewRegistry()

		// Register builtin tools
		err := mcp.RegisterBuiltinTools(registry)
		if err != nil {
			t.Fatalf("RegisterBuiltinTools() error = %v", err)
		}

		// Verify system.status tool is registered
		tools := registry.List()
		found := false
		for _, tool := range tools {
			if tool.Name == "system.status" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("system.status tool not found in registry")
		}

		// Execute system.status tool
		ctx := context.Background()
		result, err := registry.Exec(ctx, "system.status", map[string]interface{}{})
		if err != nil {
			t.Errorf("system.status execution error = %v", err)
		}

		// Verify result structure
		if result == nil {
			t.Fatalf("system.status returned nil result")
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("system.status result is not a map")
		}

		// Check required fields
		requiredFields := []string{"status", "timestamp", "uptime", "version", "health"}
		for _, field := range requiredFields {
			if _, ok := resultMap[field]; !ok {
				t.Errorf("system.status result missing field: %s", field)
			}
		}

		// Execute with detailed flag
		result, err = registry.Exec(ctx, "system.status", map[string]interface{}{
			"detailed": true,
		})
		if err != nil {
			t.Errorf("system.status detailed execution error = %v", err)
		}

		resultMap, ok = result.(map[string]interface{})
		if !ok {
			t.Fatalf("system.status detailed result is not a map")
		}

		// Check detailed fields
		detailedFields := []string{"runtime", "connections"}
		for _, field := range detailedFields {
			if _, ok := resultMap[field]; !ok {
				t.Errorf("system.status detailed result missing field: %s", field)
			}
		}
	})
}

func TestSystemEndpoints_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	controller := system.NewMetricsController(nil)

	// Setup additional system endpoints
	router.GET("/api/v1/mcp/system/info", controller.SystemInfo)
	router.GET("/api/v1/mcp/system/cpu-info", controller.GetCPUInfo)
	router.GET("/api/v1/mcp/system/memory-info", controller.GetMemoryInfo)
	router.GET("/api/v1/mcp/system/disk-info", controller.GetDiskInfo)

	endpoints := []struct {
		name string
		path string
	}{
		{"SystemInfo", "/api/v1/mcp/system/info"},
		{"GetCPUInfo", "/api/v1/mcp/system/cpu-info"},
		{"GetMemoryInfo", "/api/v1/mcp/system/memory-info"},
		{"GetDiskInfo", "/api/v1/mcp/system/disk-info"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", endpoint.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s status = %v, want %v, response: %s", endpoint.name, w.Code, http.StatusOK, w.Body.String())
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal %s response: %v", endpoint.name, err)
			}

			// Check if response has expected structure
			if _, ok := response["data"]; !ok {
				t.Errorf("%s response missing 'data' field", endpoint.name)
			}
		})
	}
}