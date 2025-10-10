package cbot

import (
	"net/http"
	"os"

	telegram_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/app/chatbots/telegram"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
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
	initArgs := rtl.GetInitArgs()
	if !gl.IsObjValid(initArgs) {
		gl.Log("error", "InitArgs is nil for TelegramRoutes")
		return nil
	}
	if initArgs.ConfigFile == "" {
		initArgs.ConfigFile = gl.GetEnvOrDefault("TELEGRAM_CONFIG_FILE", os.ExpandEnv("./config/social_meta.yaml"))
	}

	cfg, configErr := bootstrap.Load[*bootstrap.Config](initArgs)
	if configErr != nil {
		gl.Log("error", "Failed to load config for TelegramRoutes", configErr)
		return nil
	}
	svc := telegram.NewService(cfg.Integrations.Telegram)
	controller := telegram_controller.NewController(dbService.(*gdbasez.DBServiceImpl), svc)
	routes := make(map[string]ar.IRoute)
	routes["TelegramWebhook"] = proto.NewRoute(http.MethodPost, "/api/v1/telegram/webhook", "application/json", controller.HandleWebhook, nil, dbService, nil, nil)
	routes["TelegramSend"] = proto.NewRoute(http.MethodPost, "/api/v1/telegram/send", "application/json", controller.SendMessage, nil, dbService, nil, nil)
	routes["TelegramPing"] = proto.NewRoute(http.MethodGet, "/api/v1/telegram/ping", "application/json", controller.Ping, nil, dbService, nil, nil)
	return routes
}
