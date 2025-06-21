package routes

import (
	"net/http"

	"github.com/streadway/amqp"

	whk "github.corafa-moriori/gdbase/factory/models"
	"github.com/rafa-mori/gobe/internal/controllers/webhooks"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

// WebhookRoutes utiliza o padrão Route para registrar endpoints do Webhook Manager.
type WebhookRoutes struct {
	ci.IRouter
}

func NewWebhookRoutes(rtr *ci.IRouter) map[string]ci.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for WebhookRoutes", nil)
		return nil
	}
	rtl := *rtr

	// Obtenha o dbService já configurado no router
	dbService := rtl.GetDatabaseService()

	db, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from dbService")
		return nil
	}

	// Inicialize o repositório e o serviço de webhooks.
	webhookRepo := whk.NewWebhookRepo(db)
	webhookService := whk.NewWebhookService(webhookRepo)

	// Configuração do RabbitMQ
	rabbitMQConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		gl.Log("error", "Failed to connect to RabbitMQ")
		rabbitMQConn = nil // Continue sem RabbitMQ
	}

	webhookController := webhooks.NewWebhookController(webhookService, rabbitMQConn)

	// Mapear as rotas utilizando o WebhookController.
	routesMap := make(map[string]ci.IRoute)
	routesMap["RegisterWebhookRoute"] = NewRoute(http.MethodPost, "/webhooks", "application/json", webhookController.RegisterWebhook, nil, dbService)
	routesMap["ListWebhooksRoute"] = NewRoute(http.MethodGet, "/webhooks", "application/json", webhookController.ListWebhooks, nil, dbService)
	routesMap["DeleteWebhookRoute"] = NewRoute(http.MethodDelete, "/webhooks/:id", "application/json", webhookController.DeleteWebhook, nil, dbService)

	return routesMap
}
