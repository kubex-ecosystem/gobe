package cbot

import (
	"net/http"

	telegram_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/app/chatbots/telegram"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/config"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/chatbot/telegram"
)

// NewTelegramRoutes registers Telegram related endpoints.
func NewTelegramRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for TelegramRoutes")
		return nil
	}
	rtl := *rtr
	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for TelegramRoutes")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB for TelegramRoutes", err)
		return nil
	}
	cfg, err := config.Load("./")
	if err != nil {
		gl.Log("error", "Failed to load config for TelegramRoutes", err)
		return nil
	}
	svc := telegram.NewService(cfg.Integrations.Telegram)
	controller := telegram_controller.NewController(dbGorm, svc)
	routes := make(map[string]ar.IRoute)
	routes["TelegramWebhook"] = proto.NewRoute(http.MethodPost, "/api/v1/telegram/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["TelegramSend"] = proto.NewRoute(http.MethodPost, "/api/v1/telegram/send", "application/json", controller.SendMessage, nil, dbService, nil, nil)
	routes["TelegramPing"] = proto.NewRoute(http.MethodGet, "/api/v1/telegram/ping", "application/json", controller.Ping, nil, dbService, nil, nil)
	return routes
}
