package gateway

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	gui "github.com/kubex-ecosystem/gobe/internal/app/web"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// WebUIController serves the embedded web UI, if bundled.
type WebUIController struct{}

var (
	webGUIFiles *gui.GUIGoBE = gui.NewGUIGoBE()
)

func NewWebUIController() *WebUIController {
	gl.Log("info", "WebUIController initialized")
	return &WebUIController{}
}

func (wc *WebUIController) ServeRoot(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		path = "index.html"
	}
	urlPath, err := url.JoinPath("./", path)
	if err != nil {
		gl.Log("error", "Failed to join URL path for web UI", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if wc.serveFile(c, urlPath) {
		return
	}
	urlPath, err = url.JoinPath("./assets", path)
	if err != nil {
		gl.Log("error", "Failed to join URL path for web UI assets", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if wc.serveFile(c, urlPath) {
		return
	}
	// Fallback to SPA entry point.
	if wc.serveFile(c, "index.html") {
		return
	}
	c.Status(http.StatusNotFound)
}

func (wc *WebUIController) ServeFavicon(c *gin.Context) {
	if wc.serveFile(c, "favicon.ico") {
		return
	}
	c.Status(http.StatusNotFound)
}

func (wc *WebUIController) ServeAssets(c *gin.Context) {
	// Assets path is relative to /assets/
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	urlPath, err := url.JoinPath("assets", path)
	if err != nil {
		gl.Log("error", "Failed to join URL path for web UI assets", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if wc.serveFile(c, urlPath) {
		return
	}
	// Fallback to SPA entry point.
	if wc.serveFile(c, "index.html") {
		return
	}
	c.Status(http.StatusNotFound)
}

func (wc *WebUIController) ServeApp(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		path = "index.html"
	}
	if wc.serveFile(c, path) {
		return
	}
	urlPath, err := url.JoinPath("assets", path)
	if err != nil {
		gl.Log("error", "Failed to join URL path for web UI assets", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if wc.serveFile(c, urlPath) {
		return
	}
	// Fallback to SPA entry point.
	if wc.serveFile(c, "index.html") {
		return
	}
	c.Status(http.StatusNotFound)
}

func (wc *WebUIController) serveFile(c *gin.Context, relative string) bool {
	clean := filepath.Clean(relative)
	// Prevent directory traversal
	if strings.Contains(clean, "..") {
		gl.Log("warn", "Attempted directory traversal in web UI path:", relative)
		c.Status(http.StatusBadRequest)
		return true
	}
	// Serve embedded file if available
	data, err := webGUIFiles.ReadFile(clean)
	if err != nil {
		gl.Log("debug", "Web UI file not found:", clean)
		return false
	}
	contentType := mimeTypeByExtension(filepath.Ext(clean))
	c.Data(http.StatusOK, contentType, data)
	return true
}

func mimeTypeByExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".ico":
		return "image/x-icon"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}
