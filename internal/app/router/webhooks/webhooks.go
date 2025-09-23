// Package webhooks provides the routes for the Webhook Manager.
package webhooks

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	whk "github.com/kubex-ecosystem/gdbase/factory/models"
	t "github.com/kubex-ecosystem/gdbase/types"
	"github.com/kubex-ecosystem/gobe/internal/app/controllers/webhooks"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
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
	if dbService == nil {
		gl.Log("error", "Database service is nil for WebhookRoutes")
		return nil
	}

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
	url = "amqp://192.168.100.61:5672/" // Remover após testes
	gl.Log("info", fmt.Sprintf("RabbitMQ URL: %s", url))
	var rabbitMQConn *amqp.Connection
	if url != "" {
		gl.Log("info", fmt.Sprintf("Connecting to RabbitMQ at %s", url))
		rabbitMQConn, err = amqp.Dial(url)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Connection failed: %v", err))
			rabbitMQConn = nil // Continue sem RabbitMQ
		}
	}
	// Configuração do RabbitMQ
	if rabbitMQConn == nil {
		gl.Log("error", "Failed to connect to RabbitMQ")
		rabbitMQConn = nil // Continue sem RabbitMQ
	}

	webhookController := webhooks.NewWebhookController(webhookService, rabbitMQConn)

	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	// Mapear as rotas utilizando o WebhookController.
	routesMap := make(map[string]ci.IRoute)
	routesMap["RegisterWebhookRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/webhooks", "application/json", webhookController.RegisterWebhook, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ListWebhooksRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/webhooks", "application/json", webhookController.ListWebhooks, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteWebhookRoute"] = proto.NewRoute(http.MethodDelete, "/api/v1/webhooks/:id", "application/json", webhookController.DeleteWebhook, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}

func getRabbitMQURL(dbConfig *DBConfig) string {
	var host = ""
	var port = ""
	var username = ""
	var password = ""
	if dbConfig.Messagery.RabbitMQ.Host != "" {
		host = dbConfig.Messagery.RabbitMQ.Host
	} else {
		host = "localhost"
	}
	if dbConfig.Messagery.RabbitMQ.Port != "" {
		strPort, ok := dbConfig.Messagery.RabbitMQ.Port.(string)
		if ok {
			port = strPort
		} else {
			gl.Log("error", "RabbitMQ port is not a string")
			port = "5672"
		}
	} else {
		port = "5672"
	}
	if dbConfig.Messagery.RabbitMQ.Username != "" {
		username = dbConfig.Messagery.RabbitMQ.Username
	} else {
		username = "guest"
	}
	if dbConfig.Messagery.RabbitMQ.Password != "" {
		password = dbConfig.Messagery.RabbitMQ.Password
	} else {
		password = "guest"
	}

	if host != "" && port != "" && username != "" && password != "" {
		return fmt.Sprintf("amqp://%s:%se@%s:%s/", username, password, host, port)
	}
	return ""
}
