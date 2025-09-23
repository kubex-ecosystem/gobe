package router

import (
	gdbf "github.com/kubex-ecosystem/gdbase/factory"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/router/app"
	"github.com/kubex-ecosystem/gobe/internal/app/router/cbot"
	"github.com/kubex-ecosystem/gobe/internal/app/router/gateway"
	"github.com/kubex-ecosystem/gobe/internal/app/router/mcp"
	"github.com/kubex-ecosystem/gobe/internal/app/router/sys"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/app/router/user"
	"github.com/kubex-ecosystem/gobe/internal/app/router/webhooks"
)

func NewRoute(method, path, contentType string, handler gin.HandlerFunc, middlewares map[string]gin.HandlerFunc, dbConfig gdbf.DBService, secureProperties map[string]bool, metadata map[string]any) ci.IRoute {
	return proto.NewRoute(method, path, contentType, handler, middlewares, dbConfig, secureProperties, metadata)
}

func GetDefaultRouteMap(rtr ci.IRouter) map[string]map[string]ci.IRoute {
	return map[string]map[string]ci.IRoute{
		"serverManagementRoutes": sys.NewServerRoutes(&rtr),
		"cronRoutes":             sys.NewCronRoutes(&rtr),
		"swaggerRoutes":          sys.NewSwaggerRoutes(&rtr),

		"webhookRoutes": webhooks.NewWebhookRoutes(&rtr),
		"gatewayRoutes": gateway.NewGatewayRoutes(&rtr),

		"contactRoutes":  app.NewContactRoutes(&rtr),
		"productRoutes":  app.NewProductRoutes(&rtr),
		"customerRoutes": app.NewCustomerRoutes(&rtr),

		"authRoutes": user.NewAuthRoutes(&rtr),
		"userRoutes": user.NewUserRoutes(&rtr),

		"discordRoutes":  cbot.NewDiscordRoutes(&rtr),
		"whatsappRoutes": cbot.NewWhatsAppRoutes(&rtr),
		"telegramRoutes": cbot.NewTelegramRoutes(&rtr),

		"mcpTasksRoutes":       mcp.NewMCPTasksRoutes(&rtr),
		"mcpProvidersRoutes":   mcp.NewMCPProvidersRoutes(&rtr),
		"mcpLLMRoutes":         mcp.NewMCPLLMRoutes(&rtr),
		"mcpPreferencesRoutes": mcp.NewMCPPreferencesRoutes(&rtr),
		"mcpSystemRoutes":      mcp.NewMCPSystemRoutes(&rtr),
		// "mcpGHbexRoutes":       mcp.NewMCPGHbexRoutes(&rtr),
		"mcpGDBaseRoutes": mcp.NewMCPGDBaseRoutes(&rtr),
	}
}
