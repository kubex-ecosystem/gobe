package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	llmCtrl "github.com/rafa-mori/gobe/internal/controllers/mcp/llm"
	preferencesCtrl "github.com/rafa-mori/gobe/internal/controllers/mcp/preferences"
	providersCtrl "github.com/rafa-mori/gobe/internal/controllers/mcp/providers"
	tasksCtrl "github.com/rafa-mori/gobe/internal/controllers/mcp/tasks"
)

// MCPRoutes handles the registration of all MCP (Model Context Protocol) related routes
type MCPRoutes struct {
	db *gorm.DB
}

// NewMCPRoutes creates a new instance of MCPRoutes
func NewMCPRoutes(db *gorm.DB) *MCPRoutes {
	return &MCPRoutes{
		db: db,
	}
}

// RegisterMCPRoutes registers all MCP controllers and their routes
func (mcpr *MCPRoutes) RegisterMCPRoutes(router *gin.Engine) {
	// Initialize all MCP controllers
	llmController := llmCtrl.NewLLMController(mcpr.db)
	preferencesController := preferencesCtrl.NewPreferencesController(mcpr.db)
	providersController := providersCtrl.NewProvidersController(mcpr.db)
	tasksController := tasksCtrl.NewTasksController(mcpr.db)

	// Register routes for each controller
	llmController.RegisterRoutes(router)
	preferencesController.RegisterRoutes(router)
	providersController.RegisterRoutes(router)
	tasksController.RegisterRoutes(router)
}

// RegisterLLMRoutes registers only LLM routes
func (mcpr *MCPRoutes) RegisterLLMRoutes(router *gin.Engine) {
	llmController := llmCtrl.NewLLMController(mcpr.db)
	llmController.RegisterRoutes(router)
}

// RegisterPreferencesRoutes registers only Preferences routes
func (mcpr *MCPRoutes) RegisterPreferencesRoutes(router *gin.Engine) {
	preferencesController := preferencesCtrl.NewPreferencesController(mcpr.db)
	preferencesController.RegisterRoutes(router)
}

// RegisterProvidersRoutes registers only Providers routes
func (mcpr *MCPRoutes) RegisterProvidersRoutes(router *gin.Engine) {
	providersController := providersCtrl.NewProvidersController(mcpr.db)
	providersController.RegisterRoutes(router)
}

// RegisterTasksRoutes registers only Tasks routes
func (mcpr *MCPRoutes) RegisterTasksRoutes(router *gin.Engine) {
	tasksController := tasksCtrl.NewTasksController(mcpr.db)
	tasksController.RegisterRoutes(router)
}
