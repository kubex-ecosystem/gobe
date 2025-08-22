// Package cbot provides the routes for the chatbot functionality.
package cbot

import (
	"net/http"

	discord_controller "github.com/rafa-mori/gobe/internal/app/controllers/app/chatbots/discord"
	"github.com/rafa-mori/gobe/internal/app/router/proto"
	"github.com/rafa-mori/gobe/internal/config"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
	"github.com/rafa-mori/gobe/internal/proxy/hub"
)

type DiscordRoutes struct {
	ar.IRouter
	h *hub.DiscordMCPHub
}

func NewDiscordRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for DiscordRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for DiscordRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}

	routesMap := make(map[string]ar.IRoute)

	middlewaresMap := rtl.GetMiddlewares()
	if len(middlewaresMap) == 0 {
		gl.Log("error", "Middlewares map is empty for DiscordRoute")
		return nil
	}

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	config, configErr := config.Load(
		"./",
	)
	if configErr != nil {
		gl.Log("error", "Failed to load config for DiscordRoute", configErr)
		return nil
	}
	h, err := hub.NewDiscordMCPHub(config)
	if err != nil {
		gl.Log("error", "Failed to create Discord hub", err)
		return nil
	}

	discordController := discord_controller.NewDiscordController(dbGorm, h)

	routesMap["DiscordWebSocket"] = proto.NewRoute(http.MethodGet, "/api/v1/discord/websocket", "application/json", discordController.HandleWebSocket, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DiscordOAuth2Authorize"] = proto.NewRoute(http.MethodGet, "/api/v1/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DiscordOAuth2Token"] = proto.NewRoute(http.MethodGet, "/api/v1/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, middlewaresMap, dbService, secureProperties, nil)

	// Rota principal para aplicações Discord (Activities) - SEM middlewares de segurança para desenvolvimento
	routesMap["DiscordApp"] = proto.NewRoute(http.MethodGet, "/api/v1/discord", "text/html", discordController.HandleDiscordApp, nil, dbService, nil, nil)
	routesMap["OAuth2AuthorizeDiscord"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, nil, dbService, nil, nil)
	routesMap["OAuth2TokenDiscord"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, nil, dbService, nil, nil)
	routesMap["WebhookDiscord"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/webhook/:webhookId/:webhookToken", "application/json", discordController.HandleDiscordWebhook, nil, dbService, nil, nil)
	routesMap["InteractionsDiscord"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/interactions", "application/json", discordController.HandleDiscordInteractions, nil, dbService, nil, nil)
	routesMap["GetPendingApprovals"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/interactions/pending", "application/json", discordController.GetPendingApprovals, nil, dbService, nil, nil)
	routesMap["GetApprovals"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/approvals", "application/json", discordController.GetPendingApprovals, nil, dbService, nil, nil)
	routesMap["ApproveRequest"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/approve", "application/json", discordController.ApproveRequest, nil, dbService, nil, nil)
	routesMap["RejectRequest"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/reject", "application/json", discordController.RejectRequest, nil, dbService, nil, nil)
	routesMap["HandleTestMessage"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/test", "application/json", discordController.HandleTestMessage, nil, dbService, nil, nil)
	routesMap["PingDiscord"] = proto.NewRoute(http.MethodGet, "/api/v1/discord/ping", "application/json", discordController.PingDiscord, nil, dbService, nil, nil)
	routesMap["PingDiscord"] = proto.NewRoute(http.MethodPost, "/api/v1/discord/ping", "application/json", discordController.PingDiscordAdapter, nil, dbService, nil, nil)

	defer discordController.InitiateBotMCP()

	return routesMap
}
