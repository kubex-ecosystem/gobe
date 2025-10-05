package gdbasez

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	fscm "github.com/kubex-ecosystem/gdbase/factory/models"
)

type (
	// Client definitions

	ClientService = fscm.ClientService
	ClientModel   = fscm.ClientModel
	ClientRepo    = fscm.ClientRepo

	// Product definitions

	ProductModel   = fscm.ProductModel
	ProductRepo    = fscm.ProductRepo
	ProductService = fscm.ProductService

	// Cron definitions

	CronJobService = fscm.CronJobService
	CronJobRepo    = fscm.CronJobRepo
	CronJobModel   = fscm.CronJobModel

	// Discord definitions

	DiscordService = fscm.DiscordService
	DiscordRepo    = fscm.DiscordRepo
	DiscordModel   = fscm.DiscordModel

	// JobQueue definitions

	JobQueueService = fscm.JobQueueService
	JobQueueRepo    = fscm.JobQueueRepo
	JobQueueImpl    = fscm.JobQueue
	JobQueueModel   = fscm.JobQueueModel

	// Webhook definitions

	WebhookService         = fscm.WebhookService
	WebhookRepo            = fscm.WebhookRepo
	WebhookModel           = fscm.Webhook
	Webhook                = fscm.Webhook
	RegisterWebhookRequest = fscm.RegisterWebhookRequest

	// AnalysisJob definitions

	AnalysisJobService = fscm.AnalysisJobService
	AnalysisJobRepo    = fscm.AnalysisJobRepo
	AnalysisJobImpl    = fscm.AnalysisJob
	AnalysisJobModel   = fscm.AnalysisJobModel

	// OrderModel   = fscm.OrderModel
	// OrderRepo    = fscm.OrderRepo
	// OrderService = fscm.OrderService

	// ContactModel   = fscm.ContactModel
	// ContactRepo    = fscm.ContactRepo
	// ContactService = fscm.ContactService
)

func NewClientService(db ClientRepo) ClientService {
	return fscm.NewClientService(db)
}

func NewClientModel() *ClientModel {
	return &ClientModel{}
}

func NewClientRepo(ctx context.Context, dbService *svc.DBServiceImpl) ClientRepo {
	return fscm.NewClientRepo(ctx, dbService)
}

func NewProductService(db ProductRepo) ProductService {
	return fscm.NewProductService(db)
}

func NewProductModel() *ProductModel {
	return &ProductModel{}
}

func NewProductRepo(ctx context.Context, dbService *svc.DBServiceImpl) ProductRepo {
	return fscm.NewProductRepo(ctx, dbService)
}

func NewCronJobService(db CronJobRepo) CronJobService {
	return fscm.NewCronJobService(db)
}

func NewCronModel() *CronJobModel {
	return &CronJobModel{}
}

func NewCronRepo(ctx context.Context, dbService *svc.DBServiceImpl) CronJobRepo {
	return fscm.NewCronJobRepo(ctx, dbService)
}

func NewDiscordService(db DiscordRepo) DiscordService {
	return fscm.NewDiscordService(db)
}

func NewDiscordModel() *DiscordModel {
	return &DiscordModel{}
}

func NewDiscordRepo(ctx context.Context, dbService *svc.DBServiceImpl) DiscordRepo {
	return fscm.NewDiscordRepo(ctx, dbService)
}
func NewJobQueueService(db JobQueueRepo) JobQueueService {
	return fscm.NewJobQueueService(db)
}

func NewJobQueueRepo(ctx context.Context, dbService *svc.DBServiceImpl) JobQueueRepo {
	return fscm.NewJobQueueRepo(ctx, dbService)
}

func NewJobQueueModel() JobQueueModel {
	return fscm.NewJobQueueModel()
}

func NewAnalysisJobService(db AnalysisJobRepo) AnalysisJobService {
	return fscm.NewAnalysisJobService(db)
}

func NewAnalysisJobRepo(ctx context.Context, dbService *svc.DBServiceImpl) AnalysisJobRepo {
	return fscm.NewAnalysisJobRepo(ctx, dbService)
}

func NewAnalysisJobModel() AnalysisJobModel {
	return fscm.NewAnalysisJobModel()
}

func NewWebhookRepo(ctx context.Context, dbService *svc.DBServiceImpl) WebhookRepo {
	return fscm.NewWebhookRepo(ctx, dbService)
}

func NewWebhookService(repo WebhookRepo) WebhookService {
	return fscm.NewWebhookService(repo)
}
