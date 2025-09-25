package gateway

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	webhooks "github.com/kubex-ecosystem/gobe/internal/services/webhooks"
)

// WebhookController proxies webhook notifications into the GoBE event bus.
type WebhookController struct {
	webhookService *webhooks.WebhookService
}

func NewWebhookController(webhookService *webhooks.WebhookService) *WebhookController {
	return &WebhookController{
		webhookService: webhookService,
	}
}

// Handle receives webhook events and acknowledges their processing.
//
// @Summary     Receber webhook
// @Description Aceita eventos externos, valida o JSON e agenda processamento interno
// @Tags        gateway
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

	// Extract source and event type from headers or payload
	source := c.GetHeader("X-Webhook-Source")
	if source == "" {
		source = "unknown"
	}

	eventType := c.GetHeader("X-Event-Type")
	if eventType == "" {
		if et, exists := payload["event_type"].(string); exists {
			eventType = et
		} else {
			eventType = "generic"
		}
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Process webhook using service
	if wc.webhookService != nil {
		event, err := wc.webhookService.ReceiveWebhook(source, eventType, payload, headers)
		if err != nil {
			gl.Log("error", "Failed to process webhook", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to process webhook"})
			return
		}

		c.JSON(http.StatusAccepted, WebhookAckResponse{
			Status:    "received",
			Timestamp: time.Now().UTC(),
			Message:   "webhook processed successfully",
			Payload: map[string]interface{}{
				"event_id": event.ID,
				"source":   event.Source,
				"type":     event.EventType,
			},
		})
	} else {
		// Fallback if service is not available
		c.JSON(http.StatusAccepted, WebhookAckResponse{
			Status:    "received",
			Timestamp: time.Now().UTC(),
			Message:   "webhook received (service unavailable)",
			Payload:   payload,
		})
	}
}

// Health validates the readiness of the gateway webhook receiver.
//
// @Summary     Healthcheck webhooks
// @Description Retorna status do receptor de webhooks com estatísticas detalhadas
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} WebhookHealthResponse
// @Failure     401 {object} ErrorResponse
// @Router      /v1/webhooks/health [get]
func (wc *WebhookController) Health(c *gin.Context) {
	response := WebhookHealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}

	// Add detailed stats if service is available
	if wc.webhookService != nil {
		stats := wc.webhookService.GetStats()
		response.Stats = stats
	}

	c.JSON(http.StatusOK, response)
}

// ListEvents returns a paginated list of webhook events.
//
// @Summary     Listar webhook events
// @Description Retorna lista paginada de eventos de webhook recebidos
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       limit query int false "Número máximo de eventos (default: 50)"
// @Param       offset query int false "Número de eventos a pular (default: 0)"
// @Param       source query string false "Filtrar por fonte do webhook"
// @Success     200 {object} WebhookEventsResponse
// @Failure     401 {object} ErrorResponse
// @Router      /v1/webhooks/events [get]
func (wc *WebhookController) ListEvents(c *gin.Context) {
	if wc.webhookService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Status: "error", Message: "webhook service unavailable"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	source := c.Query("source")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	events, total, err := wc.webhookService.ListWebhookEvents(limit, offset, source)
	if err != nil {
		gl.Log("error", "Failed to list webhook events", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to list events"})
		return
	}

	// Convert events to interface{} slice
	eventInterfaces := make([]interface{}, len(events))
	for i, event := range events {
		eventInterfaces[i] = event
	}

	c.JSON(http.StatusOK, WebhookEventsResponse{
		Events: eventInterfaces,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// GetEvent returns details of a specific webhook event.
//
// @Summary     Obter webhook event
// @Description Retorna detalhes de um evento específico de webhook
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do evento"
// @Success     200 {object} WebhookEvent
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /v1/webhooks/events/{id} [get]
func (wc *WebhookController) GetEvent(c *gin.Context) {
	if wc.webhookService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Status: "error", Message: "webhook service unavailable"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid event ID"})
		return
	}

	event, err := wc.webhookService.GetWebhookEvent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Status: "error", Message: "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// RetryFailedEvents retries all failed webhook events.
//
// @Summary     Retry webhook events
// @Description Reagenda todos os eventos de webhook que falharam para reprocessamento
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} WebhookRetryResponse
// @Failure     401 {object} ErrorResponse
// @Router      /v1/webhooks/retry [post]
func (wc *WebhookController) RetryFailedEvents(c *gin.Context) {
	if wc.webhookService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Status: "error", Message: "webhook service unavailable"})
		return
	}

	retried, err := wc.webhookService.RetryFailedWebhooks()
	if err != nil {
		gl.Log("error", "Failed to retry webhook events", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to retry events"})
		return
	}

	c.JSON(http.StatusOK, WebhookRetryResponse{
		Status:         "success",
		RetriedCount:   retried,
		Timestamp:      time.Now().UTC(),
		Message:        "failed events queued for retry",
	})
}
