package routes

import (
	"fmt"
	"net/http"

	whk "github.com/rafa-mori/gdbase/factory/models"
	t "github.com/rafa-mori/gdbase/types"
	"github.com/rafa-mori/gobe/internal/controllers/webhooks"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
	"github.com/streadway/amqp"
)

type DBConfig = t.DBConfig

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

	dbConfig := dbService.GetConfig()
	if dbConfig == nil {
		gl.Log("error", "Failed to get DBConfig from dbService")
		return nil
	}
	url := getRabbitMQURL(dbConfig)

	var rabbitMQConn *amqp.Connection
	if url != "" {
		rabbitMQConn, err = amqp.Dial(url)
		if err != nil {
			gl.Log("error", "Failed to connect to RabbitMQ")
			rabbitMQConn = nil // Continue sem RabbitMQ
		}
	}
	// Configuração do RabbitMQ
	if rabbitMQConn == nil {
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

func getRabbitMQURL(dbConfig *DBConfig) string {
	if dbConfig != nil {
		if dbConfig.Messagery != nil {
			if dbConfig.Messagery.RabbitMQ != nil {
				return fmt.Sprintf("amqp://%s:%s@%s:%d/",
					dbConfig.Messagery.RabbitMQ.Username,
					dbConfig.Messagery.RabbitMQ.Password,
					dbConfig.Messagery.RabbitMQ.Host,
					dbConfig.Messagery.RabbitMQ.Port,
				)
			}
		}
	}
	return ""
}
