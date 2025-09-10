package testsmiddlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	"golang.org/x/time/rate"
)

func TestRateLimiterBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a rate limiter that allows 2 requests per second with burst of 2
	router := gin.New()
	router.Use(middlewares.RateLimiter(rate.Limit(2), 2))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request should succeed, got status %d", w1.Code)
	}

	// Second request should succeed (within burst)
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request should succeed, got status %d", w2.Code)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got status %d", w3.Code)
	}
}

func TestRateLimiterPerClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Note: The basic RateLimiter creates per-client limiters automatically
	// Create a rate limiter that allows 1 request per second with burst of 1
	router := gin.New()
	router.Use(middlewares.RateLimiter(rate.Limit(1), 1))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request from first client
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Request from first client should succeed, got status %d", w1.Code)
	}

	// Request from second client (different IP) should succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	req2.Header.Set("X-Forwarded-For", "192.168.1.2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Request from second client should succeed, got status %d", w2.Code)
	}

	// Second request from first client should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	req3.Header.Set("X-Forwarded-For", "192.168.1.1")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Second request from first client should be rate limited, got status %d", w3.Code)
	}
}

func TestRateLimiterHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test X-Forwarded-For header handling
	router := gin.New()
	router.Use(middlewares.RateLimiter(rate.Limit(1), 1))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request with X-Forwarded-For header
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request with X-Forwarded-For should succeed, got status %d", w1.Code)
	}

	// Second request with same X-Forwarded-For should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request with same X-Forwarded-For should be rate limited, got status %d", w2.Code)
	}

	// Request with different X-Forwarded-For should succeed
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", "203.0.113.2, 198.51.100.1")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("Request with different X-Forwarded-For should succeed, got status %d", w3.Code)
	}
}

func TestAdvancedRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Skip advanced middleware test due to complex dependencies
	t.Skip("Skipping advanced rate limit middleware test - requires complex setup")
}

func TestRateLimitMiddlewareConfig(t *testing.T) {
	// Skip configuration test due to complex dependencies
	t.Skip("Skipping rate limit middleware configuration test - requires complex setup")
}

func TestRateLimitMiddlewareErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test basic rate limiter with error handling
	router := gin.New()
	router.Use(middlewares.RateLimiter(rate.Limit(1), 1))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with malformed RemoteAddr (should not crash)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "invalid-address-format"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work, using the malformed address as client IP
	if w.Code != http.StatusOK {
		t.Errorf("Request with malformed RemoteAddr should still succeed, got status %d", w.Code)
	}
}
