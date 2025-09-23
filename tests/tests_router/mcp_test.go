package tests_router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/system"
	"github.com/kubex-ecosystem/gobe/internal/services/mcp"
)

func TestMCPEndpoints(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create a test controller
	controller := system.NewMetricsController(nil) // nil DB for test

	// Setup router
	router := gin.New()
	router.GET("/mcp/tools", controller.ListTools)
	router.POST("/mcp/exec", controller.ExecTool)

	t.Run("GET /mcp/tools - should return tools list", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/mcp/tools", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Check status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("ListTools returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check response structure
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}

		// Check if data contains tools
		data, ok := response["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be an object")
		}

		tools, ok := data["tools"].([]interface{})
		if !ok {
			t.Fatal("Expected tools to be an array")
		}

		// Should have at least system.status tool
		if len(tools) < 1 {
			t.Errorf("Expected at least 1 tool, got %d", len(tools))
		}

		// Check if system.status tool exists
		hasSystemStatus := false
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			if toolMap["name"] == "system.status" {
				hasSystemStatus = true
				break
			}
		}

		if !hasSystemStatus {
			t.Error("Expected system.status tool to be registered")
		}
	})

	t.Run("POST /mcp/exec - should execute system.status tool", func(t *testing.T) {
		// Create request body
		requestBody := map[string]interface{}{
			"tool": "system.status",
			"args": map[string]interface{}{
				"detailed": false,
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Check status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("ExecTool returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check response structure
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}

		// Check if data contains result
		data, ok := response["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be an object")
		}

		if data["tool"] != "system.status" {
			t.Errorf("Expected tool 'system.status', got %v", data["tool"])
		}

		result, ok := data["result"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be an object")
		}

		// Check if result has expected fields
		if result["status"] != "ok" {
			t.Errorf("Expected result status 'ok', got %v", result["status"])
		}

		if _, exists := result["timestamp"]; !exists {
			t.Error("Expected result to have timestamp")
		}

		if _, exists := result["version"]; !exists {
			t.Error("Expected result to have version")
		}
	})

	t.Run("POST /mcp/exec - should handle invalid tool", func(t *testing.T) {
		// Create request body
		requestBody := map[string]interface{}{
			"tool": "nonexistent.tool",
			"args": map[string]interface{}{},
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return bad request for nonexistent tool
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("ExecTool with invalid tool returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check response structure
		if response["status"] != "error" {
			t.Errorf("Expected status 'error', got %v", response["status"])
		}
	})

	t.Run("POST /mcp/exec - should handle malformed request", func(t *testing.T) {
		// Create malformed request body (missing required tool field)
		requestBody := map[string]interface{}{
			"args": map[string]interface{}{},
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/mcp/exec", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return bad request for malformed request
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("ExecTool with malformed request returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check response structure
		if response["status"] != "error" {
			t.Errorf("Expected status 'error', got %v", response["status"])
		}
	})
}

func TestSystemStatusTool(t *testing.T) {
	t.Run("system.status tool execution", func(t *testing.T) {
		// Create a registry
		registry := mcp.NewRegistry()

		// Register built-in tools
		err := mcp.RegisterBuiltinTools(registry)
		if err != nil {
			t.Fatalf("Failed to register built-in tools: %v", err)
		}

		// Test basic execution
		result, err := registry.Exec(nil, "system.status", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to execute system.status: %v", err)
		}

		// Check result structure
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be a map")
		}

		if resultMap["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", resultMap["status"])
		}

		// Test detailed execution
		result, err = registry.Exec(nil, "system.status", map[string]interface{}{
			"detailed": true,
		})
		if err != nil {
			t.Fatalf("Failed to execute system.status with detailed=true: %v", err)
		}

		// Check result structure
		resultMap, ok = result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be a map")
		}

		if resultMap["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", resultMap["status"])
		}

		// Should have runtime info when detailed=true
		if _, exists := resultMap["runtime"]; !exists {
			t.Error("Expected runtime info in detailed mode")
		}
	})
}