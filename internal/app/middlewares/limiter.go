package middlewares

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ClientLimiter manages rate limiting per client IP
type ClientLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
	cleanup  time.Duration
	lastSeen map[string]time.Time
}

// NewClientLimiter creates a new client-based rate limiter
func NewClientLimiter(limit rate.Limit, burst int) *ClientLimiter {
	cl := &ClientLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    burst,
		cleanup:  time.Hour, // Clean up old entries after 1 hour
		lastSeen: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go cl.cleanupOldEntries()

	return cl
}

// GetLimiter returns the rate limiter for a specific client IP
func (cl *ClientLimiter) GetLimiter(ip string) *rate.Limiter {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	limiter, exists := cl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(cl.limit, cl.burst)
		cl.limiters[ip] = limiter
	}

	cl.lastSeen[ip] = time.Now()
	return limiter
}

// cleanupOldEntries removes old unused limiters to prevent memory leaks
func (cl *ClientLimiter) cleanupOldEntries() {
	ticker := time.NewTicker(time.Minute * 10) // Run cleanup every 10 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cl.mu.Lock()
			now := time.Now()
			for ip, lastSeen := range cl.lastSeen {
				if now.Sub(lastSeen) > cl.cleanup {
					delete(cl.limiters, ip)
					delete(cl.lastSeen, ip)
				}
			}
			cl.mu.Unlock()
		}
	}
}

// Global client limiter instance
var globalClientLimiter *ClientLimiter

// RateLimiter creates a rate limiting middleware with per-client tracking
func RateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
	if globalClientLimiter == nil {
		globalClientLimiter = NewClientLimiter(limit, burst)
	}

	return func(c *gin.Context) {
		// Get client IP (considering proxies)
		clientIP := getClientIP(c)

		// Get the rate limiter for this client
		limiter := globalClientLimiter.GetLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": "60s", // Suggest retry time
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP extracts the real client IP considering proxies
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (most common)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := c.GetHeader("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}
