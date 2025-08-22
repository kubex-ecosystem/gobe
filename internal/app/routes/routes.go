package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	gdbf "github.com/rafa-mori/gdbase/factory"
	"github.com/rafa-mori/gobe/internal/app/routes/proto"
	ci "github.com/rafa-mori/gobe/internal/proto/interfaces"
	t "github.com/rafa-mori/gobe/internal/proto/types"
	l "github.com/rafa-mori/logz"
)

func GetDefaultRouteMap(rtr ci.IRouter) map[string]map[string]ci.IRoute {
	return map[string]map[string]ci.IRoute{
		"serverManagementRoutes": NewServerRoutes(&rtr),
		"cronRoutes":             NewCronRoutes(&rtr),
		"webhookRoutes":          NewWebhookRoutes(&rtr),
		"contactRoutes":          NewContactRoutes(&rtr),
		"authRoutes":             NewAuthRoutes(&rtr),
		"userRoutes":             NewUserRoutes(&rtr),
		"productRoutes":          NewProductRoutes(&rtr),
		"customerRoutes":         NewCustomerRoutes(&rtr),
		"discordRoutes":          NewDiscordRoutes(&rtr),
		"whatsappRoutes":         NewWhatsAppRoutes(&rtr),
		"telegramRoutes":         NewTelegramRoutes(&rtr),
		"mcpTasksRoutes":         NewMCPTasksRoutes(&rtr),
		"mcpProvidersRoutes":     NewMCPProvidersRoutes(&rtr),
		"mcpLLMRoutes":           NewMCPLLMRoutes(&rtr),
		"mcpPreferencesRoutes":   NewMCPPreferencesRoutes(&rtr),
		"mcpSystemRoutes":        NewMCPSystemRoutes(&rtr),
		"mcpGHbexRoutes":         NewMCPGHbexRoutes(&rtr),
		"swaggerRoutes":          NewSwaggerRoutes(&rtr),
	}
}

func UniqueMiddlewareStack(middlewares []gin.HandlerFunc) []gin.HandlerFunc {
	uniqueMap := make(map[string]gin.HandlerFunc)
	uniqueList := []gin.HandlerFunc{}

	for _, middleware := range middlewares {
		funcPtr := fmt.Sprintf("%p", middleware) // Obtém o endereço da função como string

		if _, exists := uniqueMap[funcPtr]; !exists {
			uniqueMap[funcPtr] = middleware
			uniqueList = append(uniqueList, middleware)
		}
	}

	return uniqueList
}

func NewRoute(method, path, contentType string, handler gin.HandlerFunc, middlewares map[string]gin.HandlerFunc, dbConfig gdbf.DBService, secureProperties map[string]bool, metadata map[string]any) ci.IRoute {
	return proto.NewRoute(method, path, contentType, handler, middlewares, dbConfig, secureProperties, metadata)
}

// NewRouter creates a new Router instance and returns it as an IRouter interface.
func NewRouter(serverConfig *t.GoBEConfig, databaseService gdbf.DBService, logger l.Logger, debug bool) (ci.IRouter, error) {
	return proto.NewRouter(serverConfig, databaseService, logger, debug)
}

// NewRequest is a placeholder function for creating a new request.
func NewRequest(dBConfig gdbf.DBConfig, s string, i1, i2 int) (any, any) {
	panic("unimplemented")
}
