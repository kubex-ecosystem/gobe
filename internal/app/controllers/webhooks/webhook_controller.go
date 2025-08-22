// Package webhooks provides the WebhookController for managing webhook-related operations.
package webhooks

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	whk "github.com/rafa-mori/gdbase/factory/models"
	t "github.com/rafa-mori/gobe/internal/proto/types"
	"github.com/streadway/amqp"
)

type WebhookController struct {
	Service      whk.WebhookService
	RabbitMQConn *amqp.Connection
	APIWrapper   *t.APIWrapper[any]
}

func NewWebhookController(service whk.WebhookService, rabbitMQConn *amqp.Connection) *WebhookController {
	return &WebhookController{
		Service:      service,
		RabbitMQConn: rabbitMQConn,
		APIWrapper:   t.NewAPIWrapper[any](),
	}
}

func (wc *WebhookController) RegisterWebhook(ctx *gin.Context) {
	var webhook whk.Webhook
	if err := ctx.ShouldBindJSON(&webhook); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	created, err := wc.Service.RegisterWebhook(webhook)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register webhook"})
		return
	}

	// Publish event to RabbitMQ
	if wc.RabbitMQConn != nil {
		channel, err := wc.RabbitMQConn.Channel()
		if err == nil {
			defer channel.Close()
			channel.Publish(
				"webhook_events",  // exchange
				"webhook.created", // routing key
				false,             // mandatory
				false,             // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(created.GetID().String()),
				},
			)
		}
	}

	ctx.JSON(http.StatusCreated, created)
}

func (wc *WebhookController) ListWebhooks(ctx *gin.Context) {
	webhooks, err := wc.Service.ListWebhooks()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}
	ctx.JSON(http.StatusOK, webhooks)
}

func (wc *WebhookController) DeleteWebhook(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		// Invalid ID
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err = wc.Service.RemoveWebhook(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Webhook deleted"})
}
