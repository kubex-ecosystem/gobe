package telegram

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	tg "github.com/rafa-mori/gobe/internal/telegram"
)

// Controller handles Telegram webhook events and messaging.
type Controller struct {
	db      *gorm.DB
	service *tg.Service
}

// NewController creates a new Telegram controller.
func NewController(db *gorm.DB, service *tg.Service) *Controller {
	return &Controller{db: db, service: service}
}

// HandleWebhook processes incoming Telegram updates.
func (c *Controller) HandleWebhook(ctx *gin.Context) {
	var update map[string]any
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg := tg.Message{}
	if m, ok := update["message"].(map[string]any); ok {
		if from, ok := m["from"].(map[string]any); ok {
			msg.From, _ = from["username"].(string)
		}
		msg.ChatID, _ = getInt64(m["chat"].(map[string]any)["id"])
		msg.Text, _ = m["text"].(string)
	}
	c.db.Create(&msg)
	ctx.Status(http.StatusOK)
}

func getInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

// SendMessage sends a message via Telegram service.
func (c *Controller) SendMessage(ctx *gin.Context) {
	var req struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.service.SendMessage(tg.OutgoingMessage{ChatID: req.ChatID, Text: req.Text}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// Ping endpoint for health checks.
func (c *Controller) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
