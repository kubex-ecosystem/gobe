package whatsapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	wa "github.com/rafa-mori/gobe/internal/whatsapp"
)

// Controller manages WhatsApp webhooks and message sending.
type Controller struct {
	db      *gorm.DB
	service *wa.Service
}

// NewController returns a new WhatsApp controller.
func NewController(db *gorm.DB, service *wa.Service) *Controller {
	return &Controller{db: db, service: service}
}

// HandleWebhook processes incoming WhatsApp webhook events and verification.
func (c *Controller) HandleWebhook(ctx *gin.Context) {
	if ctx.Request.Method == http.MethodGet {
		mode := ctx.Query("hub.mode")
		token := ctx.Query("hub.verify_token")
		challenge := ctx.Query("hub.challenge")
		if mode == "subscribe" && token == c.service.Config().VerifyToken {
			ctx.String(http.StatusOK, challenge)
			return
		}
		ctx.Status(http.StatusForbidden)
		return
	}

	var payload map[string]any
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Persist a simplified message record if possible
	msg := wa.Message{Text: ""}
	if entry, ok := payload["entry"].([]any); ok && len(entry) > 0 {
		if changes, ok := entry[0].(map[string]any)["changes"].([]any); ok && len(changes) > 0 {
			if value, ok := changes[0].(map[string]any)["value"].(map[string]any); ok {
				if msgs, ok := value["messages"].([]any); ok && len(msgs) > 0 {
					if m, ok := msgs[0].(map[string]any); ok {
						msg.From, _ = m["from"].(string)
						if text, ok := m["text"].(map[string]any); ok {
							msg.Text, _ = text["body"].(string)
						}
					}
				}
			}
		}
	}
	c.db.Create(&msg)
	ctx.Status(http.StatusOK)
}

// SendMessage sends a message using the service.
func (c *Controller) SendMessage(ctx *gin.Context) {
	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.service.SendMessage(wa.OutgoingMessage{To: req.To, Text: req.Message}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// Ping verifies service availability.
func (c *Controller) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
