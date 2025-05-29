package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	whk "github.com/rafa-mori/gdbase/factory/models"
	"github.com/rafa-mori/gobe/internal/types"
	"github.com/streadway/amqp"
)

type WebhookController struct {
	Service      whk.WebhookService
	RabbitMQConn *amqp.Connection
	ApiWrapper   *types.ApiWrapper[any]
}

func NewWebhookController(service whk.WebhookService, rabbitMQConn *amqp.Connection) *WebhookController {
	return &WebhookController{
		Service:      service,
		RabbitMQConn: rabbitMQConn,
		ApiWrapper:   types.NewApiWrapper[any](),
	}
}

func (wc *WebhookController) RegisterWebhook(ctx *gin.Context) {
	var request whk.RegisterWebhookRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		wc.ApiWrapper.JSONResponseWithError(ctx, fmt.Errorf("Invalid request: %v", err))
		return
	}

	// if _, err := wc.Service.RegisterWebhook(request); err != nil {
	// 	wc.ApiWrapper.JSONResponseWithError(ctx, http.StatusInternalServerError, err)
	// 	return
	// }

	wc.ApiWrapper.JSONResponseWithSuccess(ctx, "Webhook registered successfully", "", http.StatusCreated)
}
