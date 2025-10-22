// Package gdbasez provides a clean bridge to gdbase without leaking infrastructure
package gdbasez

import (
	"context"
	"time"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	models "github.com/kubex-ecosystem/gdbase/factory/models"
	mcpmodels "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// Bridge provides a clean interface to gdbase services without exposing *gorm.DB
// This is the ONLY place where *gorm.DB should be visible in gobe
type Bridge struct {
	dbService svc.DBService
	ctx       context.Context
}

// NewBridge creates a new clean bridge to gdbase
// This is the ONLY function in gobe that should accept *gorm.DB
func NewBridge(ctx context.Context, dbService svc.DBService, dbName string) *Bridge {
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
func (b *Bridge) DBService() svc.DBService {
	return b.dbService
}

// ========================================
// Job Queue (DBService needed)
// ========================================

func (b *Bridge) JobQueueModel() JobQueueModel {
	return models.NewJobQueueModel()
}
func (b *Bridge) JobQueueRepo(ctx context.Context, dbService svc.DBService) JobQueueRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewJobQueueRepo(ctx, dbSvc)
}
func (b *Bridge) JobQueueService(repo JobQueueRepo) JobQueueService {
	return models.NewJobQueueService(repo)
}

// ========================================
// Webhooks (DBService needed)
// ========================================

func (b *Bridge) WebhookModel(fullURL string, event string, status string) WebhookModel {
	return models.NewWebhookModel(fullURL, event, status)
}
func (b *Bridge) WebhookRepo(ctx context.Context, dbService svc.DBService) WebhookRepo {
	return models.NewWebhookRepo(ctx, dbService)
}
func (b *Bridge) WebhookService(repo WebhookRepo) WebhookService {
	return models.NewWebhookService(repo)
}

// ========================================
// Analysis Jobs (DBService needed)
// ========================================

func (b *Bridge) AnalysisJobModel() AnalysisJobModel {
	return models.NewAnalysisJobModel()
}
func (b *Bridge) AnalysisJobRepo(ctx context.Context, dbService svc.DBService) AnalysisJobRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewAnalysisJobRepo(ctx, dbSvc)
}
func (b *Bridge) AnalysisJobService(repo AnalysisJobRepo) AnalysisJobService {
	return models.NewAnalysisJobService(repo)
}

// ========================================
// OAuth (PKCE) (DBService needed)
// ========================================

func (b *Bridge) OAuthClientModel(clientID, clientName string, redirectURIs, scopes []string) OAuthClient {
	return models.NewOAuthClientModel(clientID, clientName, redirectURIs, scopes)
}
func (b *Bridge) OAuthClientRepo(ctx context.Context, dbService svc.DBService) OAuthClientRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewOAuthClientRepo(ctx, dbSvc)
}
func (b *Bridge) OAuthClientService(repo OAuthClientRepo) OAuthClientService {
	return models.NewOAuthClientService(repo)
}

func (b *Bridge) AuthCodeModel(code string, clientID string, userID string, redirectURI string, codeChallenge string, method string) AuthCode {
	return models.NewAuthCodeModel(code, clientID, redirectURI, codeChallenge, userID, method)
}
func (b *Bridge) AuthCodeRepo(ctx context.Context, dbService svc.DBService) AuthCodeRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewAuthCodeRepo(ctx, dbSvc)
}
func (b *Bridge) AuthCodeService(repo AuthCodeRepo) AuthCodeService {
	return models.NewAuthCodeService(repo)
}

// ========================================
// Model/Repo/Service Constructors (DBService needed)
// ========================================

func (b *Bridge) UserModel(username string, name string, email string) UserModel {
	return models.NewUserModel(username, name, email)
}
func (b *Bridge) UserRepo(ctx context.Context, dbService svc.DBService) UserRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewUserRepo(ctx, dbSvc)
}
func (b *Bridge) UserService(repo UserRepo) UserService {
	return models.NewUserService(repo)
}

func (b *Bridge) ClientModel() *ClientModel {
	return &ClientModel{}
}
func (b *Bridge) ClientRepo(ctx context.Context, dbService svc.DBService) ClientRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewClientRepo(ctx, dbSvc)
}
func (b *Bridge) ClientService(repo ClientRepo) ClientService {
	return models.NewClientService(repo)
}

func (b *Bridge) ProductModel() *ProductModel {
	return &ProductModel{}
}
func (b *Bridge) ProductRepo(ctx context.Context, dbService svc.DBService) ProductRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewProductRepo(ctx, dbSvc)
}
func (b *Bridge) ProductService(repo ProductRepo) ProductService {
	return models.NewProductService(repo)
}

func (b *Bridge) CronJobRepo(ctx context.Context, dbService svc.DBService) CronJobRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewCronJobRepo(ctx, dbSvc)
}
func (b *Bridge) CronJobServiceImpl(repo *CronJobRepoImpl) *CronJobServiceImpl {
	return models.NewCronJobService(repo).(*CronJobServiceImpl)
}
func (b *Bridge) CronJobService(repo *CronJobRepoImpl) models.CronJobService {
	return models.NewCronJobServiceImpl(repo)
}

func (b *Bridge) DiscordModel() DiscordModel {
	return models.NewDiscordModel()
}
func (b *Bridge) DiscordRepo(ctx context.Context, dbService svc.DBService) DiscordRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewDiscordRepo(ctx, dbSvc)
}
func (b *Bridge) DiscordService(repo DiscordRepo) DiscordService {
	return models.NewDiscordService(repo)
}

// func (b *Bridge) AnalysisJobModel() AnalysisJobModel {
// 	return models.NewAnalysisJobModel()
// }
// func (b *Bridge) AnalysisJobRepo(ctx context.Context, dbService svc.DBService) AnalysisJobRepo {
// 	var dbSvc *svc.DBServiceImpl
// 	if dbService != nil {
// 		dbSvc = dbService.(*svc.DBServiceImpl)
// 	}
// 	return models.NewAnalysisJobRepo(ctx, dbSvc)
// }
// func (b *Bridge) AnalysisJobService(repo AnalysisJobRepo) AnalysisJobService {
// 	return models.NewAnalysisJobService(repo)
// }

// ========================================
// MCP: LLM
// ========================================

func (b *Bridge) LLMModel(enabled bool, provider string, model string, temperature float64, maxTokens int, topP float64, frequencyPenalty float64, presencePenalty float64, stopSequences []string) LLMModel {
	return mcpmodels.NewLLMModel(enabled, provider, model, temperature, maxTokens, topP, frequencyPenalty, presencePenalty, stopSequences)
}
func (b *Bridge) LLMRepo(ctx context.Context, dbService svc.DBService) LLMRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return mcpmodels.NewLLMRepo(ctx, dbSvc)
}
func (b *Bridge) LLMService(repo LLMRepo) LLMService {
	return mcpmodels.NewLLMService(repo)
}

// ========================================
// MCP: Tasks
// ========================================

func (b *Bridge) TasksModel(
	provider string,
	target string,
	taskType mcpmodels.TaskType,
	taskSchedule mcpmodels.JobScheduleType,
	taskExpression string,
	taskCommandType string,
	taskMethod mcpmodels.HTTPMethod,
	taskAPIEndpoint string,
	taskPayload svc.JSONBImpl,
	taskHeaders svc.JSONBImpl,
	taskRetries int,
	taskTimeout int,
	taskStatus mcpmodels.TaskStatus,
	taskNextRun *time.Time,
	taskLastRun *time.Time,
	taskLastRunStatus string, taskLastRunMessage string,
	taskCommand string,
	taskActivated bool,
	taskConfig svc.JSONBImpl,
	taskTags []string,
	taskPriority int,
	taskNotes string,
	taskCreatedAt string,
	taskUpdatedAt string,
	taskCreatedBy string,
	taskUpdatedBy string,
	taskLastExecutedBy string,
	taskLastExecutedAt *time.Time,
	config svc.JSONBImpl,
	active bool,
) TasksModel {
	return mcpmodels.NewTasksModel(
		provider,
		target,
		taskType,
		taskSchedule,
		taskExpression,
		taskCommandType,
		taskMethod,
		taskAPIEndpoint,
		taskPayload,
		taskHeaders,
		taskRetries,
		taskTimeout,
		taskStatus,
		taskNextRun,
		taskLastRun,
		taskLastRunStatus,
		taskLastRunMessage,
		taskCommand,
		taskActivated,
		taskConfig,
		taskTags,
		taskPriority,
		taskNotes,
		taskCreatedAt,
		taskUpdatedAt,
		taskCreatedBy,
		taskUpdatedBy,
		taskLastExecutedBy,
		taskLastExecutedAt,
		config,
		active,
	)
}
func (b *Bridge) TasksRepo(ctx context.Context, dbService svc.DBService) TasksRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return mcpmodels.NewTasksRepo(ctx, dbSvc)
}
func (b *Bridge) TasksService(repo TasksRepo) TasksService {
	return mcpmodels.NewTasksService(repo)
}

// ========================================
// MCP: Preferences
// ========================================

func (b *Bridge) PreferencesModel(scope string, config svc.JSONBImpl) PreferencesModel {
	return mcpmodels.NewPreferencesModel(scope, config)
}
func (b *Bridge) PreferencesRepo(ctx context.Context, dbService svc.DBService) PreferencesRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return mcpmodels.NewPreferencesRepo(ctx, dbSvc)
}
func (b *Bridge) PreferencesService(repo PreferencesRepo) PreferencesService {
	return mcpmodels.NewPreferencesService(repo)
}

// ========================================
// MCP: Providers
// ========================================

func (b *Bridge) ProvidersModel(provider string, orgOrGroup string, config svc.JSONBImpl) ProvidersModel {
	return mcpmodels.NewProvidersModel(provider, orgOrGroup, config)
}
func (b *Bridge) ProvidersRepo(ctx context.Context, dbService svc.DBService) ProvidersRepo {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return mcpmodels.NewProvidersRepo(ctx, dbSvc)
}
func (b *Bridge) ProvidersService(repo ProvidersRepo) ProvidersService {
	return mcpmodels.NewProvidersService(repo)
}

func (b *Bridge) RegistrationTokenModel(userID, token string, expiresAt time.Time) RegistrationTokenModel {
	return *models.NewRegistrationToken(userID, token, expiresAt)
}
func (b *Bridge) RegistrationTokenRepo(ctx context.Context, dbService svc.DBService) (RegistrationTokenRepo, error) {
	var dbSvc *svc.DBServiceImpl
	if dbService != nil {
		dbSvc = dbService.(*svc.DBServiceImpl)
	}
	return models.NewRegistrationTokenRepo(ctx, dbSvc)
}
func (b *Bridge) RegistrationTokenService(repo RegistrationTokenRepo) RegistrationTokenService {
	return models.NewRegistrationTokenService(repo)
}
