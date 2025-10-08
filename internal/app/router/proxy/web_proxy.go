// Package proxy provides reverse proxy functionality for Kubex ecosystem web UIs
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// ProxyConfig holds configuration for web proxies
type ProxyConfig struct {
	GromptURL   string // Default: http://localhost:8080
	AnalyzerURL string // Default: http://localhost:8081
	GemXURL     string // Default: http://localhost:8082
}

// DefaultProxyConfig returns default configuration
func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		GromptURL:   "http://localhost:8080",
		AnalyzerURL: "http://localhost:8081",
		GemXURL:     "http://localhost:8082",
	}
}

// WebProxyRouter handles reverse proxy routes for ecosystem UIs
type WebProxyRouter struct {
	config        ProxyConfig
	gromptProxy   *httputil.ReverseProxy
	analyzerProxy *httputil.ReverseProxy
	gemxProxy     *httputil.ReverseProxy
}

// NewWebProxyRouter creates a new web proxy router
func NewWebProxyRouter(config ProxyConfig) (*WebProxyRouter, error) {
	// Parse URLs
	gromptURL, err := url.Parse(config.GromptURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Grompt URL: %w", err)
	}

	analyzerURL, err := url.Parse(config.AnalyzerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Analyzer URL: %w", err)
	}

	gemxURL, err := url.Parse(config.GemXURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GemX URL: %w", err)
	}

	// Create reverse proxies
	gromptProxy := httputil.NewSingleHostReverseProxy(gromptURL)
	analyzerProxy := httputil.NewSingleHostReverseProxy(analyzerURL)
	gemxProxy := httputil.NewSingleHostReverseProxy(gemxURL)

	// Custom error handlers
	gromptProxy.ErrorHandler = createErrorHandler("Grompt")
	analyzerProxy.ErrorHandler = createErrorHandler("Analyzer")
	gemxProxy.ErrorHandler = createErrorHandler("GemX")

	gl.Log("info", "Web proxy router initialized",
		"grompt", config.GromptURL,
		"analyzer", config.AnalyzerURL,
		"gemx", config.GemXURL)

	return &WebProxyRouter{
		config:        config,
		gromptProxy:   gromptProxy,
		analyzerProxy: analyzerProxy,
		gemxProxy:     gemxProxy,
	}, nil
}

// RegisterRoutes registers all proxy routes
func (w *WebProxyRouter) RegisterRoutes(router *gin.RouterGroup) {
	// Grompt UI Proxy
	router.Any("/grompt", w.handleGromptRoot)
	router.Any("/grompt/*path", w.handleGrompt)

	// Analyzer UI Proxy
	router.Any("/analyzer", w.handleAnalyzerRoot)
	router.Any("/analyzer/*path", w.handleAnalyzer)

	// GemX UI Proxy (future)
	router.Any("/gemx", w.handleGemXRoot)
	router.Any("/gemx/*path", w.handleGemX)

	gl.Log("info", "Web proxy routes registered: /web/grompt, /web/analyzer, /web/gemx")
}

// handleGromptRoot handles /web/grompt
func (w *WebProxyRouter) handleGromptRoot(c *gin.Context) {
	gl.Log("debug", "Proxying Grompt root", "path", c.Request.URL.Path)

	// Rewrite path
	c.Request.URL.Path = "/"
	c.Request.Host = w.config.GromptURL

	// Add proxy headers
	addProxyHeaders(c, "grompt")

	// Proxy request
	w.gromptProxy.ServeHTTP(c.Writer, c.Request)
}

// handleGrompt handles /web/grompt/*
func (w *WebProxyRouter) handleGrompt(c *gin.Context) {
	path := c.Param("path")
	gl.Log("debug", "Proxying Grompt", "path", path)

	// Rewrite path: /web/grompt/api/config → /api/config
	c.Request.URL.Path = path
	c.Request.Host = w.config.GromptURL

	// Add proxy headers
	addProxyHeaders(c, "grompt")

	// Proxy request
	w.gromptProxy.ServeHTTP(c.Writer, c.Request)
}

// handleAnalyzerRoot handles /web/analyzer
func (w *WebProxyRouter) handleAnalyzerRoot(c *gin.Context) {
	gl.Log("debug", "Proxying Analyzer root", "path", c.Request.URL.Path)

	c.Request.URL.Path = "/"
	c.Request.Host = w.config.AnalyzerURL

	addProxyHeaders(c, "analyzer")
	w.analyzerProxy.ServeHTTP(c.Writer, c.Request)
}

// handleAnalyzer handles /web/analyzer/*
func (w *WebProxyRouter) handleAnalyzer(c *gin.Context) {
	path := c.Param("path")
	gl.Log("debug", "Proxying Analyzer", "path", path)

	c.Request.URL.Path = path
	c.Request.Host = w.config.AnalyzerURL

	addProxyHeaders(c, "analyzer")
	w.analyzerProxy.ServeHTTP(c.Writer, c.Request)
}

// handleGemXRoot handles /web/gemx
func (w *WebProxyRouter) handleGemXRoot(c *gin.Context) {
	gl.Log("debug", "Proxying GemX root", "path", c.Request.URL.Path)

	c.Request.URL.Path = "/"
	c.Request.Host = w.config.GemXURL

	addProxyHeaders(c, "gemx")
	w.gemxProxy.ServeHTTP(c.Writer, c.Request)
}

// handleGemX handles /web/gemx/*
func (w *WebProxyRouter) handleGemX(c *gin.Context) {
	path := c.Param("path")
	gl.Log("debug", "Proxying GemX", "path", path)

	c.Request.URL.Path = path
	c.Request.Host = w.config.GemXURL

	addProxyHeaders(c, "gemx")
	w.gemxProxy.ServeHTTP(c.Writer, c.Request)
}

// addProxyHeaders adds necessary headers for proxying
func addProxyHeaders(c *gin.Context, service string) {
	// Forward original headers
	c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
	c.Request.Header.Set("X-Forwarded-Proto", "http")
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)
	c.Request.Header.Set("X-Proxied-By", "gobe")
	c.Request.Header.Set("X-Service", service)

	// Pass auth token if present
	if token := c.GetHeader("Authorization"); token != "" {
		c.Request.Header.Set("Authorization", token)
	}

	// Pass session if present
	if session := c.GetHeader("X-Session-ID"); session != "" {
		c.Request.Header.Set("X-Session-ID", session)
	}
}

// createErrorHandler creates a custom error handler for proxies
func createErrorHandler(serviceName string) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		gl.Log("error", "Proxy error", "service", serviceName, "error", err, "path", r.URL.Path)

		// Check if service is down
		if strings.Contains(err.Error(), "connection refused") {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>%s Unavailable</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .container {
            text-align: center;
            background: rgba(255,255,255,0.1);
            padding: 3rem;
            border-radius: 1rem;
            backdrop-filter: blur(10px);
        }
        h1 { font-size: 3rem; margin: 0; }
        p { font-size: 1.2rem; margin: 1rem 0; }
        code { background: rgba(0,0,0,0.3); padding: 0.5rem 1rem; border-radius: 0.5rem; display: inline-block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚠️ %s Unavailable</h1>
        <p>The %s service is currently offline or unreachable.</p>
        <p>Please ensure it's running:</p>
        <code>cd /projects/kubex/%s && ./%s start</code>
    </div>
</body>
</html>`, serviceName, serviceName, serviceName, strings.ToLower(serviceName), strings.ToLower(serviceName))
			return
		}

		// Generic error
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Proxy Error</title>
    <style>
        body {
            font-family: monospace;
            padding: 2rem;
            background: #1a1a1a;
            color: #00ff00;
        }
        pre { background: #000; padding: 1rem; border-radius: 0.5rem; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>🔴 Proxy Error - %s</h1>
    <pre>%s</pre>
    <p>Path: %s</p>
</body>
</html>`, serviceName, err.Error(), r.URL.Path)
	}
}
