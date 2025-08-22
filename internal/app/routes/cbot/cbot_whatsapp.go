package cbot

import (
	"net/http"

	whatsapp_controller "github.com/rafa-mori/gobe/internal/app/controllers/apps/chatbots/whatsapp"
	"github.com/rafa-mori/gobe/internal/config"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
	"github.com/rafa-mori/gobe/internal/services/chatbot/whatsapp"
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
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB for WhatsAppRoutes", err)
		return nil
	}
	cfg, err := config.Load("./")
	if err != nil {
		gl.Log("error", "Failed to load config for WhatsAppRoutes", err)
		return nil
	}
	svc := whatsapp.NewService(cfg.Integrations.WhatsApp)
	controller := whatsapp_controller.NewController(dbGorm, svc)
	routes := make(map[string]ar.IRoute)
	routes["WhatsAppWebhookPost"] = NewRoute(http.MethodPost, "/api/v1/whatsapp/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["WhatsAppWebhookGet"] = NewRoute(http.MethodGet, "/api/v1/whatsapp/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["WhatsAppSend"] = NewRoute(http.MethodPost, "/api/v1/whatsapp/send", "application/json", controller.SendMessage, nil, dbService, nil, nil)
	routes["WhatsAppPing"] = NewRoute(http.MethodGet, "/api/v1/whatsapp/ping", "application/json", controller.Ping, nil, dbService, nil, nil)
	return routes
}
