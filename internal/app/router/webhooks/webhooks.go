// Package webhooks provides the routes for the Webhook Manager.
package webhooks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/logz"
)

// WebhookRoutes utiliza o padrão Route para registrar endpoints do Webhook Manager.
type WebhookRoutes struct {
	ci.IRouter
}

func NewWebhookRoutes(rtr *ci.IRouter) map[string]ci.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for WebhookRoutes")
		return nil
	}
	// rtl := *rtr

	// // Obtenha o dbService já configurado no router
	// dbService := rtl.GetDatabaseService()
	// if dbService == nil {
	// 	gl.Log("error", "Database service is nil for WebhookRoutes")
	// 	return nil
	// }

	// db, err := dbService.GetDB()
	// if err != nil {
	// 	gl.Log("error", "Failed to get DB from dbService")
	// 	return nil
	// }

	// // Inicialize o repositório e o serviço de webhooks.
	// webhookRepo := whk.NewWebhookRepo(db)
	// webhookService := whk.NewWebhookService(webhookRepo)

	// dbConfig := dbService.GetConfig()
	// if dbConfig == nil {
	// 	gl.Log("error", "Failed to get DBConfig from dbService")
	// 	return nil
	// }
	// url := msg.GetRabbitMQURL(dbConfig)
	// gl.Log("debug", fmt.Sprintf("RabbitMQ URL: %s", url))
	// var rabbitMQConn *amqp.Connection
	// if url != "" {
	// 	gl.Log("debug", fmt.Sprintf("Connecting to RabbitMQ at %s", url))
	// 	rabbitMQConn, err = amqp.Dial(url)
	// 	if err != nil {
	// 		gl.Log("error", fmt.Sprintf("Connection failed: %v", err))
	// 		rabbitMQConn = nil // Continue sem RabbitMQ
	// 	}
	// }
	// // Configuração do RabbitMQ
	// if rabbitMQConn == nil {
	// 	gl.Log("error", "Failed to connect to RabbitMQ")
	// 	rabbitMQConn = nil // Continue sem RabbitMQ
	// }

	// webhookController := webhooks.NewWebhookController(webhookService, rabbitMQConn)

	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	// Mapear as rotas utilizando o WebhookController.
	routesMap := make(map[string]ci.IRoute)
	routesMap["RegisterWebhookRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/webhooks", "application/json", nil, middlewaresMap, nil, secureProperties, nil)
	routesMap["ListWebhooksRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/webhooks", "application/json", nil, middlewaresMap, nil, secureProperties, nil)
	routesMap["DeleteWebhookRoute"] = proto.NewRoute(http.MethodDelete, "/api/v1/webhooks/:id", "application/json", nil, middlewaresMap, nil, secureProperties, nil)

	return routesMap
}
