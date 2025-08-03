package routes

import (
	"net/http"

	discord_controller "github.com/rafa-mori/gobe/internal/controllers/discord"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
)

type DiscordRoutes struct {
	ar.IRouter
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
	discordController := discord_controller.NewDiscordController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	middlewaresMap := rtl.GetMiddlewares()
	if len(middlewaresMap) == 0 {
		gl.Log("error", "Middlewares map is empty for DiscordRoute")
		return nil
	}

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["OAuth2AuthorizeDiscord"] = NewRoute(http.MethodGet, "/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, middlewaresMap, dbService, secureProperties)
	routesMap["OAuth2TokenDiscord"] = NewRoute(http.MethodGet, "/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, middlewaresMap, dbService, secureProperties)
	routesMap["WebhookDiscord"] = NewRoute(http.MethodPost, "/discord/webhook/:webhookId/:webhookToken", "application/json", discordController.HandleDiscordWebhook, middlewaresMap, dbService, secureProperties)
	routesMap["InteractionsDiscord"] = NewRoute(http.MethodPut, "/discord/interactions", "application/json", discordController.HandleDiscordInteractions, middlewaresMap, dbService, secureProperties)
	routesMap["GetPendingApprovals"] = NewRoute(http.MethodGet, "/discord/interactions/pending", "application/json", discordController.GetPendingApprovals, middlewaresMap, dbService, secureProperties)
	routesMap["GetApprovals"] = NewRoute(http.MethodGet, "/discord/approvals", "application/json", discordController.GetPendingApprovals, middlewaresMap, dbService, secureProperties)
	routesMap["ApproveRequest"] = NewRoute(http.MethodGet, "/discord/approve", "application/json", discordController.ApproveRequest, middlewaresMap, dbService, secureProperties)
	routesMap["RejectRequest"] = NewRoute(http.MethodGet, "/discord/reject", "application/json", discordController.RejectRequest, middlewaresMap, dbService, secureProperties)
	routesMap["HandleTestMessage"] = NewRoute(http.MethodGet, "/discord/test", "application/json", discordController.HandleTestMessage, middlewaresMap, dbService, secureProperties)

	// HandleWebSocket

	return routesMap
}
