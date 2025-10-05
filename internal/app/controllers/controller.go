// Package controllers provides the controller logic for handling webhooks.
package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gdbase/factory/models"

	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

type WebhookController struct {
	Service      mdl.WebhookService
	RabbitMQConn *amqp.Connection
	APIWrapper   *types.APIWrapper[any]
}

func NewWebhookController(service mdl.WebhookService, rabbitMQConn *amqp.Connection) *WebhookController {
	return &WebhookController{
		Service:      service,
		RabbitMQConn: rabbitMQConn,
		APIWrapper:   types.NewAPIWrapper[any](),
	}
}

func (wc *WebhookController) RegisterWebhook(ctx *gin.Context) {
	var request mdl.RegisterWebhookRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		wc.APIWrapper.JSONResponseWithError(ctx, fmt.Errorf("invalid request: %v", err))
		return
	}

	// if _, err := wc.Service.RegisterWebhook(request); err != nil {
	// 	wc.APIWrapper.JSONResponseWithError(ctx, http.StatusInternalServerError, err)
	// 	return
	// }

	wc.APIWrapper.JSONResponseWithSuccess(ctx, "Webhook registered successfully", "", http.StatusCreated)
}
