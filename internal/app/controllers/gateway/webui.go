package gateway

import (
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
    gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

// WebUIController serves the embedded web UI, if bundled.
type WebUIController struct {
    root      string
    indexPath string
}

func NewWebUIController(root string) *WebUIController {
    if root == "" {
        root = "web"
    }
    absRoot, err := filepath.Abs(root)
    if err != nil {
        gl.Log("error", "failed to resolve web root")
        absRoot = root
    }
    index := filepath.Join(absRoot, "index.html")
    return &WebUIController{root: absRoot, indexPath: index}
}

func (wc *WebUIController) ServeRoot(c *gin.Context) {
    if wc.serveFile(c, "index.html") {
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Web UI bundle not present", "root": wc.root})
}

func (wc *WebUIController) ServeApp(c *gin.Context) {
    path := strings.TrimPrefix(c.Param("path"), "/")
    if path == "" {
        path = "index.html"
    }
    if wc.serveFile(c, path) {
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
    full := filepath.Join(wc.root, clean)
    if !strings.HasPrefix(full, wc.root) {
        c.Status(http.StatusForbidden)
        return true
    }
    if _, err := os.Stat(full); err != nil {
        return false
    }
    c.File(full)
    return true
}

