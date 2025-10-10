// Package gdbasez provides a clean bridge to gdbase without leaking infrastructure
package gdbasez

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	models "github.com/kubex-ecosystem/gdbase/factory/models"
	mcpmodels "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// Bridge provides a clean interface to gdbase services without exposing *gorm.DB
// This is the ONLY place where *gorm.DB should be visible in gobe
type Bridge struct {
	dbService *svc.DBServiceImpl
	ctx       context.Context
}

// NewBridge creates a new clean bridge to gdbase
// This is the ONLY function in gobe that should accept *gorm.DB
func NewBridge(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) *Bridge {
	if dbService == nil {
		gl.Log("error", "Bridge: dbService is nil")
		return nil
	}

	return &Bridge{
		dbService: dbService,
		ctx:       ctx,
	}
}

// WithContext returns a new bridge with the specified context
func (b *Bridge) WithContext(ctx context.Context) *Bridge {
	return &Bridge{
		dbService: b.dbService,
		ctx:       ctx,
	}
}

// DBService returns the underlying *gorm.DB used by the bridge.
// This is provided for callers that need direct DB access for repo constructors.
func (b *Bridge) DBService() *svc.DBServiceImpl {
	return b.dbService
}

// ========================================
// Users
// ========================================

func (b *Bridge) UserService(ctx context.Context) UserService {
	repo := models.NewUserRepo(ctx, b.dbService)
	return models.NewUserService(repo)
}

// ========================================
// Clients
// ========================================

func (b *Bridge) ClientService(ctx context.Context) ClientService {
	repo := models.NewClientRepo(ctx, b.dbService)
	return models.NewClientService(repo)
}

// ========================================
// Products
// ========================================

func (b *Bridge) ProductService(ctx context.Context) ProductService {
	repo := models.NewProductRepo(ctx, b.dbService)
	return models.NewProductService(repo)
}

// ========================================
// Cron Jobs
// ========================================

func (b *Bridge) NewCronJobRepoImpl(ctx context.Context, db *svc.DBServiceImpl) *CronJobRepoImpl {
	return models.NewCronJobRepo(ctx, db).(*CronJobRepoImpl)
}
func (b *Bridge) NewCronJobRepo(ctx context.Context, db *svc.DBServiceImpl) CronJobRepo {
	return models.NewCronJobRepo(ctx, db)
}
func (b *Bridge) NewCronJobServiceImpl(ctx context.Context, repo *CronJobRepoImpl) *CronJobServiceImpl {
	return models.NewCronJobService(repo).(*CronJobServiceImpl)
}
func (b *Bridge) NewCronJobService(ctx context.Context, repo *CronJobRepoImpl) models.CronJobService {
	return models.NewCronJobServiceImpl(repo)
}

// ========================================
// Discord
// ========================================

func (b *Bridge) DiscordService(ctx context.Context) DiscordService {
	repo := models.NewDiscordRepo(ctx, b.dbService)
	return models.NewDiscordService(repo)
}

// ========================================
// Job Queue
// ========================================

func (b *Bridge) JobQueueService(ctx context.Context) JobQueueService {
	repo := models.NewJobQueueRepo(ctx, b.dbService)
	return models.NewJobQueueService(repo)
}

// ========================================
// Webhooks
// ========================================

func (b *Bridge) WebhookService(ctx context.Context) WebhookService {
	repo := models.NewWebhookRepo(ctx, b.dbService)
	return models.NewWebhookService(repo)
}

// ========================================
// Analysis Jobs
// ========================================

func (b *Bridge) AnalysisJobService(ctx context.Context) AnalysisJobService {
	repo := models.NewAnalysisJobRepo(ctx, b.dbService)
	return models.NewAnalysisJobService(repo)
}

// ========================================
// OAuth (PKCE)
// ========================================

func (b *Bridge) OAuthClientService(ctx context.Context) OAuthClientService {
	repo := models.NewOAuthClientRepo(ctx, b.dbService)
	return models.NewOAuthClientService(repo)
}

func (b *Bridge) AuthCodeService(ctx context.Context) AuthCodeService {
	repo := models.NewAuthCodeRepo(ctx, b.dbService)
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

func (b *Bridge) LLMService(ctx context.Context) LLMService {
	repo := mcpmodels.NewLLMRepo(ctx, b.dbService)
	return mcpmodels.NewLLMService(repo)
}

// ========================================
// MCP: Tasks
// ========================================

func (b *Bridge) TasksService(ctx context.Context) TasksService {
	repo := mcpmodels.NewTasksRepo(ctx, b.dbService)
	return mcpmodels.NewTasksService(repo)
}

// ========================================
// MCP: Preferences
// ========================================

func (b *Bridge) PreferencesService(ctx context.Context) PreferencesService {
	repo := mcpmodels.NewPreferencesRepo(ctx, b.dbService)
	return mcpmodels.NewPreferencesService(repo)
}

// ========================================
// MCP: Providers
// ========================================

func (b *Bridge) ProvidersService(ctx context.Context) ProvidersService {
	repo := mcpmodels.NewProvidersRepo(ctx, b.dbService)
	return mcpmodels.NewProvidersService(repo)
}
