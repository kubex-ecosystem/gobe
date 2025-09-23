package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookController proxies webhook notifications into the GoBE event bus (placeholder).
type WebhookController struct{}

func NewWebhookController() *WebhookController { return &WebhookController{} }

// Handle receives webhook events and acknowledges their processing.
//
// @Summary     Receber webhook
// @Description Aceita eventos externos, valida o JSON e agenda processamento interno. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body map[string]interface{} true "Evento enviado pelo provedor"
// @Success     202 {object} WebhookAckResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /v1/webhooks [post]
func (wc *WebhookController) Handle(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid request body"})
		return
	}

	c.JSON(http.StatusAccepted, WebhookAckResponse{
		Status:    "received",
		Timestamp: time.Now().UTC(),
		Message:   "TODO: persist webhook payload",
		Payload:   payload,
	})
}

// Health validates the readiness of the gateway webhook receiver.
//
// @Summary     Healthcheck webhooks
// @Description Retorna status simplificado do receptor de webhooks. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} WebhookHealthResponse
// @Failure     401 {object} ErrorResponse
// @Router      /v1/webhooks/health [get]
func (wc *WebhookController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, WebhookHealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}
