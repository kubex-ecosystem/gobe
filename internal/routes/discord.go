package routes

import (
	"net/http"

	discord_controller "github.com/rafa-mori/gobe/internal/controllers/discord"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type DiscordRoutes struct {
	ar.IRouter
}

func NewDiscordRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for DiscordRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	discordController := discord_controller.NewDiscordController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["OAuth2AuthorizeDiscord"] = NewRoute(http.MethodGet, "/discord/oauth2/authorize", "application/json", discordController.HandleDiscordOAuth2Authorize, middlewaresMap, dbService)
	routesMap["OAuth2TokenDiscord"] = NewRoute(http.MethodGet, "/discord/oauth2/token", "application/json", discordController.HandleDiscordOAuth2Token, middlewaresMap, dbService)
	routesMap["WebhookDiscord"] = NewRoute(http.MethodPost, "/discord/webhook/:webhookId/:webhookToken", "application/json", discordController.HandleDiscordWebhook, middlewaresMap, dbService)
	routesMap["InteractionsDiscord"] = NewRoute(http.MethodPut, "/discord/interactions", "application/json", discordController.HandleDiscordInteractions, middlewaresMap, dbService)
	routesMap["GetPendingApprovals"] = NewRoute(http.MethodGet, "/discord/interactions/pending", "application/json", discordController.GetPendingApprovals, middlewaresMap, dbService)
	routesMap["ApproveRequest"] = NewRoute(http.MethodGet, "/discord/approve", "application/json", discordController.ApproveRequest, middlewaresMap, dbService)
	routesMap["RejectRequest"] = NewRoute(http.MethodGet, "/discord/reject", "application/json", discordController.RejectRequest, middlewaresMap, dbService)
	routesMap["HandleTestMessage"] = NewRoute(http.MethodGet, "/discord/test", "application/json", discordController.HandleTestMessage, middlewaresMap, dbService)

	// HandleWebSocket

	return routesMap
}
