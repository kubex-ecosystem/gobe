package cbot

import (
	"net/http"
	"os"

	whatsapp_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/app/chatbots/whatsapp"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"github.com/kubex-ecosystem/gobe/internal/services/chatbot/whatsapp"
)

// NewWhatsAppRoutes registers WhatsApp related endpoints.
func NewWhatsAppRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for WhatsAppRoutes")
		return nil
	}
	rtl := *rtr
	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for WhatsAppRoutes")
		return nil
	}
	initArgs := rtl.GetInitArgs()
	if !gl.IsObjValid(initArgs) {
		gl.Log("error", "InitArgs is nil for WhatsAppRoutes")
		return nil
	}
	if initArgs.ConfigFile == "" {
		initArgs.ConfigFile = gl.GetEnvOrDefault("WHATSAPP_CONFIG_FILE", os.ExpandEnv("./support/whatsapp_config.yaml"))
	}

	cfg, configErr := bootstrap.Load[*bootstrap.Config](initArgs)
	if configErr != nil {
		gl.Log("error", "Failed to load config for WhatsAppRoutes", configErr)
		return nil
	}
	svc := whatsapp.NewService(cfg.Integrations.WhatsApp)
	controller := whatsapp_controller.NewController(dbService, svc)
	routes := make(map[string]ar.IRoute)
	routes["WhatsAppWebhookPost"] = proto.NewRoute(http.MethodPost, "/api/v1/whatsapp/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["WhatsAppWebhookGet"] = proto.NewRoute(http.MethodGet, "/api/v1/whatsapp/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["WhatsAppSend"] = proto.NewRoute(http.MethodPost, "/api/v1/whatsapp/send", "application/json", controller.SendMessage, nil, dbService, nil, nil)
	routes["WhatsAppPing"] = proto.NewRoute(http.MethodGet, "/api/v1/whatsapp/ping", "application/json", controller.Ping, nil, dbService, nil, nil)
	return routes
}
