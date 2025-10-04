// Package gdbasez provides a clean bridge to gdbase without leaking infrastructure
package gdbasez

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	models "github.com/kubex-ecosystem/gdbase/factory/models"
	mcpmodels "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"gorm.io/gorm"
)

// Bridge provides a clean interface to gdbase services without exposing *gorm.DB
// This is the ONLY place where *gorm.DB should be visible in gobe
type Bridge struct {
	db  *gorm.DB
	ctx context.Context
}

// NewBridge creates a new clean bridge to gdbase
// This is the ONLY function in gobe that should accept *gorm.DB
func NewBridge(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) *Bridge {
	if dbService == nil {
		gl.Log("error", "Bridge: dbService is nil")
		return nil
	}
	db, err := dbService.GetDB(ctx, dbName)
	if err != nil {
		gl.Log("error", "Bridge: failed to get db: %v", err)
		return nil
	}
	if db == nil {
		gl.Log("error", "Bridge: db is nil")
		return nil
	}
	return &Bridge{
		db:  db,
		ctx: ctx,
	}
}

// WithContext returns a new bridge with the specified context
func (b *Bridge) WithContext(ctx context.Context) *Bridge {
	return &Bridge{
		db:  b.db,
		ctx: ctx,
	}
}

// ========================================
// Users
// ========================================

func (b *Bridge) UserService() UserService {
	repo := models.NewUserRepo(b.db)
	return models.NewUserService(repo)
}

// ========================================
// Clients
// ========================================

func (b *Bridge) ClientService() ClientService {
	repo := models.NewClientRepo(b.db)
	return models.NewClientService(repo)
}

// ========================================
// Products
// ========================================

func (b *Bridge) ProductService() ProductService {
	repo := models.NewProductRepo(b.db)
	return models.NewProductService(repo)
}

// ========================================
// Cron Jobs
// ========================================

func (b *Bridge) NewCronJobService() CronJobService {
	repo := models.NewCronJobRepo(b.ctx, b.db)
	return models.NewCronJobService(repo)
}

// ========================================
// Discord
// ========================================

func (b *Bridge) DiscordService() DiscordService {
	repo := models.NewDiscordRepo(b.db)
	return models.NewDiscordService(repo)
}

// ========================================
// Job Queue
// ========================================

func (b *Bridge) JobQueueService() JobQueueService {
	repo := models.NewJobQueueRepo(b.db)
	return models.NewJobQueueService(repo)
}

// ========================================
// Webhooks
// ========================================

func (b *Bridge) WebhookService() WebhookService {
	repo := models.NewWebhookRepo(b.db)
	return models.NewWebhookService(repo)
}

// ========================================
// Analysis Jobs
// ========================================

func (b *Bridge) AnalysisJobService() AnalysisJobService {
	repo := models.NewAnalysisJobRepo(b.db)
	return models.NewAnalysisJobService(repo)
}

// ========================================
// OAuth (PKCE)
// ========================================

func (b *Bridge) OAuthClientService(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) OAuthClientService {
	repo := models.NewOAuthClientRepo(ctx, dbService, dbName)
	return models.NewOAuthClientService(repo)
}

func (b *Bridge) AuthCodeService(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) AuthCodeService {
	repo := models.NewAuthCodeRepo(ctx, dbService, dbName)
	return models.NewAuthCodeService(repo)
}

// ========================================
// Model Constructors (no DB needed)
// ========================================

func (b *Bridge) NewClientModel() *ClientModel {
	return &ClientModel{}
}

func (b *Bridge) NewProductModel() *ProductModel {
	return &ProductModel{}
}

func (b *Bridge) NewCronJobModel() *CronJobModel {
	return &CronJobModel{}
}

func (b *Bridge) NewDiscordModel() *DiscordModel {
	return &DiscordModel{}
}

func (b *Bridge) NewJobQueueModel() JobQueueModel {
	return models.NewJobQueueModel()
}

func (b *Bridge) NewAnalysisJobModel() AnalysisJobModel {
	return models.NewAnalysisJobModel()
}

func (b *Bridge) NewOAuthClientModel(clientID, clientName string, redirectURIs, scopes []string) OAuthClient {
	return models.NewOAuthClientModel(clientID, clientName, redirectURIs, scopes)
}

// ========================================
// MCP: LLM
// ========================================

func (b *Bridge) LLMService() LLMService {
	repo := mcpmodels.NewLLMRepo(b.db)
	return mcpmodels.NewLLMService(repo)
}

// ========================================
// MCP: Tasks
// ========================================

func (b *Bridge) TasksService() TasksService {
	repo := mcpmodels.NewTasksRepo(b.db)
	return mcpmodels.NewTasksService(repo)
}

// ========================================
// MCP: Preferences
// ========================================

func (b *Bridge) PreferencesService() PreferencesService {
	repo := mcpmodels.NewPreferencesRepo(b.db)
	return mcpmodels.NewPreferencesService(repo)
}

// ========================================
// MCP: Providers
// ========================================

func (b *Bridge) ProvidersService() ProvidersService {
	repo := mcpmodels.NewProvidersRepo(b.db)
	return mcpmodels.NewProvidersService(repo)
}
