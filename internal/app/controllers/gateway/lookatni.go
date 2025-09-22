package gateway

import (
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

// LookAtniController holds placeholder endpoints for LookAtni automation hooks.
type LookAtniController struct{}

func NewLookAtniController() *LookAtniController { return &LookAtniController{} }

func (lc *LookAtniController) Extract(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    c.JSON(http.StatusAccepted, gin.H{
        "status":    "queued",
        "operation": "extract",
        "payload":   payload,
        "message":   "TODO: wire LookAtni extract pipeline",
        "timestamp": time.Now().UTC(),
    })
}

func (lc *LookAtniController) Archive(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    c.JSON(http.StatusAccepted, gin.H{
        "status":    "queued",
        "operation": "archive",
        "payload":   payload,
        "timestamp": time.Now().UTC(),
        "message":   "TODO: connect LookAtni archive endpoint",
    })
}

func (lc *LookAtniController) Download(c *gin.Context) {
    resourceID := strings.TrimSpace(c.Param("id"))
    if resourceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing resource id"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "download_url": fmt.Sprintf("https://lookatni.local/%s", resourceID),
        "expires_in":   3600,
        "note":         "TODO: proxy real LookAtni artifact",
    })
}

func (lc *LookAtniController) Projects(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "projects": []map[string]interface{}{
            {
                "id":          "demo-project",
                "name":        "Demo Project",
                "description": "Placeholder LookAtni project",
            },
        },
        "version": "gateway-placeholder-1",
    })
}

