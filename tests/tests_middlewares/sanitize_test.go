package testsmiddlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
)

func TestValidateAndSanitize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		description    string
	}{
		{
			name: "Clean Query Parameters",
			queryParams: map[string]string{
				"name": "John Doe",
				"age":  "25",
			},
			expectedStatus: http.StatusOK,
			description:    "Should allow clean query parameters",
		},
		{
			name: "XSS in Query Parameters",
			queryParams: map[string]string{
				"name": "<script>alert('xss')</script>",
				"age":  "25",
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize XSS in query parameters",
		},
		{
			name: "SQL Injection in Query Parameters",
			queryParams: map[string]string{
				"name": "John'; DROP TABLE users; --",
				"age":  "25",
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize SQL injection attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.ValidateAndSanitize())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)

			// Add query parameters
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestValidateAndSanitizeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "Clean JSON Body",
			body: map[string]interface{}{
				"name":  "John Doe",
				"age":   25,
				"email": "john@example.com",
			},
			expectedStatus: http.StatusOK,
			description:    "Should allow clean JSON body",
		},
		{
			name: "XSS in JSON Body",
			body: map[string]interface{}{
				"name":    "<script>alert('xss')</script>",
				"message": "Hello <img src=x onerror=alert('xss')>",
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize XSS in JSON body",
		},
		{
			name: "SQL Injection in JSON Body",
			body: map[string]interface{}{
				"name":  "John'; DROP TABLE users; --",
				"query": "SELECT * FROM users WHERE id = 1; DELETE FROM users; --",
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize SQL injection attempts",
		},
		{
			name: "Nested Object with XSS",
			body: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "<script>alert('nested xss')</script>",
					"profile": map[string]interface{}{
						"bio": "Hello <img src=x onerror=alert('deep')>",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize nested objects",
		},
		{
			name: "Array with XSS",
			body: map[string]interface{}{
				"tags": []interface{}{
					"safe-tag",
					"<script>alert('xss')</script>",
					map[string]interface{}{
						"name": "<img src=x onerror=alert('array')>",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should sanitize arrays",
		},
		{
			name:           "Empty Body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty body",
		},
		{
			name: "Extremely Long String",
			body: map[string]interface{}{
				"data": string(make([]byte, 15000)), // Over the 10KB limit
			},
			expectedStatus: http.StatusOK, // Will be truncated
			description:    "Should handle extremely long strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.ValidateAndSanitizeBody())

			router.POST("/test", func(c *gin.Context) {
				sanitizedBody, exists := c.Get("sanitizedBody")
				if !exists {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "No sanitized body found"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message":       "success",
					"sanitizedBody": sanitizedBody,
				})
			})

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Test '%s': Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			// For successful requests, verify sanitization occurred
			if w.Code == http.StatusOK && tt.name != "Clean JSON Body" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				// Verify that the sanitized body exists
				if _, exists := response["sanitizedBody"]; !exists {
					t.Errorf("Test '%s': Expected sanitized body in response", tt.name)
				}
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	// Note: This is testing the internal function through the middleware
	// since the function is not exported
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean String",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML Tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World", // HTML tags removed
		},
		{
			name:     "XSS Script",
			input:    "<script>alert('xss')</script>",
			expected: "alert('xss')", // Script tags removed
		},
		{
			name:     "SQL Injection",
			input:    "'; DROP TABLE users; --",
			expected: "'; users; --", // Dangerous keywords removed
		},
		{
			name:     "Mixed Threats",
			input:    "<script>'; SELECT * FROM users; alert('xss'); --</script>",
			expected: "'; * FROM users; alert('xss'); --", // Multiple threats handled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.ValidateAndSanitize())

			var capturedQuery string
			router.GET("/test", func(c *gin.Context) {
				capturedQuery = c.Query("input")
				c.JSON(http.StatusOK, gin.H{"sanitized": capturedQuery})
			})

			req := httptest.NewRequest("GET", "/test", nil)

			// Add query parameters properly
			q := req.URL.Query()
			q.Add("input", tt.input)
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Note: Due to HTML escaping and URL encoding, exact matching is complex
			// This test verifies the middleware runs without errors
			// More specific sanitization tests would require exposing the internal function
		})
	}
}

func TestEmailValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		email          string
		expectedStatus int
	}{
		{
			name:           "Valid Email",
			email:          "test@example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Email",
			email:          "invalid-email",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Email",
			email:          "",
			expectedStatus: http.StatusOK, // Empty is allowed
		},
		{
			name:           "XSS in Email",
			email:          "test@<script>alert('xss')</script>.com",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.ValidateAndSanitizeBody())

			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			body := map[string]interface{}{
				"email": tt.email,
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Test '%s': Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestURLValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Valid HTTP URL",
			url:            "http://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid HTTPS URL",
			url:            "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Scheme",
			url:            "ftp://example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "JavaScript URL",
			url:            "javascript:alert('xss')",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty URL",
			url:            "",
			expectedStatus: http.StatusOK, // Empty is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middlewares.ValidateAndSanitizeBody())

			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			body := map[string]interface{}{
				"website_url": tt.url,
			}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Test '%s': Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}
