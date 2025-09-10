package middlewares

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	srv "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
)

// RequestInfo stores information about a request
type RequestInfo struct {
	IP        string
	Port      string
	Path      string
	Method    string
	UserAgent string
	Timestamp time.Time
}

// ClientRequestTracker tracks requests per client
type ClientRequestTracker struct {
	requests    []RequestInfo
	mutex       sync.RWMutex
	limit       int
	windowSize  time.Duration
	lastCleanup time.Time
}

// RateLimitMiddleware provides advanced rate limiting with database persistence
type RateLimitMiddleware struct {
	dbConfig      *srv.IDBConfig
	LogFile       string
	requestLimit  int
	requestWindow time.Duration
	g             ci.IGoBE
	clients       map[string]*ClientRequestTracker
	globalMutex   sync.RWMutex
}

// NewRateLimitMiddleware creates a new rate limit middleware instance
func NewRateLimitMiddleware(g ci.IGoBE, dbConfig srv.IDBConfig, logDir string, limit int, window time.Duration) (*RateLimitMiddleware, error) {
	rl := &RateLimitMiddleware{
		dbConfig:      &dbConfig,
		LogFile:       logDir,
		requestLimit:  limit,
		requestWindow: window,
		g:             g,
		clients:       make(map[string]*ClientRequestTracker),
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredRequests()

	return rl, nil
}

// RateLimit checks if a request should be allowed based on rate limiting rules
func (rl *RateLimitMiddleware) RateLimit(c *gin.Context) bool {
	ip, port, splitHostPortErr := net.SplitHostPort(c.Request.RemoteAddr)
	if splitHostPortErr != nil {
		// If we can't split, use the full RemoteAddr as IP
		ip = c.Request.RemoteAddr
		port = "unknown"
		log.Printf("WARN: Error splitting host and port: %v", splitHostPortErr.Error())
	}

	// Create request info
	requestInfo := RequestInfo{
		IP:        ip,
		Port:      port,
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		UserAgent: c.Request.UserAgent(),
		Timestamp: time.Now(),
	}

	// Check if this client is within rate limits
	if !rl.isRequestAllowed(requestInfo) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Request limit exceeded",
			"message":     fmt.Sprintf("Too many requests from %s. Limit: %d requests per %v", ip, rl.requestLimit, rl.requestWindow),
			"retry_after": rl.requestWindow.String(),
		})
		c.Abort()

		log.Printf("WARN: Rate limit exceeded for IP %s:%s - %s %s", ip, port, requestInfo.Method, requestInfo.Path)
		return false
	}

	// Add this request to the tracker
	rl.addRequest(requestInfo)

	c.Next()
	return true
}

// isRequestAllowed checks if a request should be allowed
func (rl *RateLimitMiddleware) isRequestAllowed(req RequestInfo) bool {
	rl.globalMutex.Lock()
	defer rl.globalMutex.Unlock()

	clientKey := req.IP

	// Get or create client tracker
	tracker, exists := rl.clients[clientKey]
	if !exists {
		tracker = &ClientRequestTracker{
			requests:    make([]RequestInfo, 0),
			limit:       rl.requestLimit,
			windowSize:  rl.requestWindow,
			lastCleanup: time.Now(),
		}
		rl.clients[clientKey] = tracker
	}

	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	// Clean up old requests for this client
	rl.cleanupClientRequests(tracker)

	// Count requests in the current window
	now := time.Now()
	count := 0
	for _, request := range tracker.requests {
		if now.Sub(request.Timestamp) <= rl.requestWindow {
			count++
		}
	}

	// Check if we're within limits
	return count < rl.requestLimit
}

// addRequest adds a request to the tracking system
func (rl *RateLimitMiddleware) addRequest(req RequestInfo) {
	rl.globalMutex.Lock()
	defer rl.globalMutex.Unlock()

	clientKey := req.IP
	tracker := rl.clients[clientKey]

	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	tracker.requests = append(tracker.requests, req)
}

// cleanupClientRequests removes expired requests for a specific client
func (rl *RateLimitMiddleware) cleanupClientRequests(tracker *ClientRequestTracker) {
	now := time.Now()
	validRequests := make([]RequestInfo, 0)

	for _, request := range tracker.requests {
		if now.Sub(request.Timestamp) <= rl.requestWindow {
			validRequests = append(validRequests, request)
		}
	}

	tracker.requests = validRequests
	tracker.lastCleanup = now
}

// cleanupExpiredRequests periodically cleans up expired requests
func (rl *RateLimitMiddleware) cleanupExpiredRequests() {
	ticker := time.NewTicker(time.Minute * 5) // Run cleanup every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.globalMutex.Lock()
			now := time.Now()

			for clientKey, tracker := range rl.clients {
				tracker.mutex.Lock()

				// Clean up old requests
				rl.cleanupClientRequests(tracker)

				// Remove clients with no recent activity
				if len(tracker.requests) == 0 && now.Sub(tracker.lastCleanup) > rl.requestWindow*2 {
					delete(rl.clients, clientKey)
				}

				tracker.mutex.Unlock()
			}

			rl.globalMutex.Unlock()
			log.Printf("INFO: Rate limit cleanup completed. Active clients: %d", len(rl.clients))
		}
	}
}

// GetRequestLimit returns the current request limit
func (rl *RateLimitMiddleware) GetRequestLimit() int {
	return rl.requestLimit
}

// SetRequestLimit sets the request limit
func (rl *RateLimitMiddleware) SetRequestLimit(limit int) {
	rl.requestLimit = limit

	// Update all client trackers
	rl.globalMutex.Lock()
	defer rl.globalMutex.Unlock()

	for _, tracker := range rl.clients {
		tracker.mutex.Lock()
		tracker.limit = limit
		tracker.mutex.Unlock()
	}
}

// GetRequestWindow returns the current request window
func (rl *RateLimitMiddleware) GetRequestWindow() time.Duration {
	return rl.requestWindow
}

// SetRequestWindow sets the request window
func (rl *RateLimitMiddleware) SetRequestWindow(window time.Duration) {
	rl.requestWindow = window

	// Update all client trackers
	rl.globalMutex.Lock()
	defer rl.globalMutex.Unlock()

	for _, tracker := range rl.clients {
		tracker.mutex.Lock()
		tracker.windowSize = window
		tracker.mutex.Unlock()
	}
}

// GetStats returns statistics about the rate limiter
func (rl *RateLimitMiddleware) GetStats() map[string]interface{} {
	rl.globalMutex.RLock()
	defer rl.globalMutex.RUnlock()

	stats := map[string]interface{}{
		"active_clients": len(rl.clients),
		"request_limit":  rl.requestLimit,
		"request_window": rl.requestWindow.String(),
		"total_requests": 0,
	}

	totalRequests := 0
	for _, tracker := range rl.clients {
		tracker.mutex.RLock()
		totalRequests += len(tracker.requests)
		tracker.mutex.RUnlock()
	}

	stats["total_requests"] = totalRequests
	return stats
}

// Middleware function to be used with Gin
func (rl *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.RateLimit(c)
	}
}
