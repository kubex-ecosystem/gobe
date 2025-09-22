package gateway

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

// WebhookController proxies webhook notifications into the GoBE event bus (placeholder).
type WebhookController struct{}

func NewWebhookController() *WebhookController { return &WebhookController{} }

func (wc *WebhookController) Handle(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    c.JSON(http.StatusAccepted, gin.H{
        "status":    "received",
        "timestamp": time.Now().UTC(),
        "message":   "TODO: persist webhook payload",
        "payload":   payload,
    })
}

func (wc *WebhookController) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "ok",
        "timestamp": time.Now().UTC(),
    })
}

