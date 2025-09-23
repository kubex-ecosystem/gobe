package whatsapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	wa "github.com/kubex-ecosystem/gobe/internal/services/chatbot/whatsapp"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

// Controller manages WhatsApp webhooks and message sending.
type Controller struct {
	db      *gorm.DB
	service *wa.Service
}

type (
	// ErrorResponse padroniza mensagens de erro para os endpoints do WhatsApp.
	ErrorResponse = t.ErrorResponse
)

// SendMessageRequest descreve o payload para disparo manual de mensagens.
type SendMessageRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// NewController returns a new WhatsApp controller.
func NewController(db *gorm.DB, service *wa.Service) *Controller {
	return &Controller{db: db, service: service}
}

// HandleWebhook processes incoming WhatsApp webhook events and verification.
//
// @Summary     Webhook WhatsApp
// @Description Recebe chamadas de verificação (GET) e eventos (POST) vindo do WhatsApp Business. [Em desenvolvimento]
// @Tags        whatsapp beta
// @Accept      json
// @Produce     json
// @Param       hub.mode         query string false "Modo do webhook"       example(subscribe)
// @Param       hub.verify_token query string false "Token de verificação"
// @Param       hub.challenge    query string false "Token de desafio"
// @Param       payload body map[string]any true "Evento enviado pelo WhatsApp"
// @Success     200 {object} map[string]string
// @Failure     400 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /api/v1/whatsapp/webhook [get]
// @Router      /api/v1/whatsapp/webhook [post]
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
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: err.Error()})
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
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SendMessage sends a message using the service.
//
// @Summary     Enviar mensagem WhatsApp
// @Description Dispara mensagem via canal de integração configurado. [Em desenvolvimento]
// @Tags        whatsapp beta
// @Accept      json
// @Produce     json
// @Param       payload body SendMessageRequest true "Dados da mensagem"
// @Success     200 {object} map[string]string "status"
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/whatsapp/send [post]
func (c *Controller) SendMessage(ctx *gin.Context) {
	var req SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: err.Error()})
		return
	}
	if err := c.service.SendMessage(wa.OutgoingMessage{To: req.To, Text: req.Message}); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// Ping verifies service availability.
//
// @Summary     Ping WhatsApp
// @Description Verifica se o adaptador de WhatsApp está ativo. [Em desenvolvimento]
// @Tags        whatsapp beta
// @Produce     json
// @Success     200 {object} map[string]string "status"
// @Router      /api/v1/whatsapp/ping [get]
func (c *Controller) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
