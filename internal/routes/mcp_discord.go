package routes

import (
	"net/http"

	"github.com/rafa-mori/gobe/internal/config"
	discord_controller "github.com/rafa-mori/gobe/internal/controllers/discord"
	"github.com/rafa-mori/gobe/internal/hub"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
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

	routesMap["DiscordWebSocket"] = NewRoute(http.MethodGet, "/api/v1/discord/websocket", "application/json", discordController.HandleWebSocket, middlewaresMap, dbService, secureProperties)
	routesMap["DiscordOAuth2Authorize"] = NewRoute(http.MethodGet, "/api/v1/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, middlewaresMap, dbService, secureProperties)
	routesMap["DiscordOAuth2Token"] = NewRoute(http.MethodGet, "/api/v1/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, middlewaresMap, dbService, secureProperties)

	// Rota principal para aplicações Discord (Activities) - SEM middlewares de segurança para desenvolvimento
	routesMap["DiscordApp"] = NewRoute(http.MethodGet, "/api/v1/discord", "text/html", discordController.HandleDiscordApp, nil, dbService, nil)
	routesMap["OAuth2AuthorizeDiscord"] = NewRoute(http.MethodPost, "/api/v1/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, nil, dbService, nil)
	routesMap["OAuth2TokenDiscord"] = NewRoute(http.MethodPost, "/api/v1/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, nil, dbService, nil)
	routesMap["WebhookDiscord"] = NewRoute(http.MethodPost, "/api/v1/discord/webhook/:webhookId/:webhookToken", "application/json", discordController.HandleDiscordWebhook, nil, dbService, nil)
	routesMap["InteractionsDiscord"] = NewRoute(http.MethodPost, "/api/v1/discord/interactions", "application/json", discordController.HandleDiscordInteractions, nil, dbService, nil)
	routesMap["GetPendingApprovals"] = NewRoute(http.MethodPost, "/api/v1/discord/interactions/pending", "application/json", discordController.GetPendingApprovals, nil, dbService, nil)
	routesMap["GetApprovals"] = NewRoute(http.MethodPost, "/api/v1/discord/approvals", "application/json", discordController.GetPendingApprovals, nil, dbService, nil)
	routesMap["ApproveRequest"] = NewRoute(http.MethodPost, "/api/v1/discord/approve", "application/json", discordController.ApproveRequest, nil, dbService, nil)
	routesMap["RejectRequest"] = NewRoute(http.MethodPost, "/api/v1/discord/reject", "application/json", discordController.RejectRequest, nil, dbService, nil)
	routesMap["HandleTestMessage"] = NewRoute(http.MethodPost, "/api/v1/discord/test", "application/json", discordController.HandleTestMessage, nil, dbService, nil)

	defer discordController.InitiateBotMCP()

	return routesMap
}
