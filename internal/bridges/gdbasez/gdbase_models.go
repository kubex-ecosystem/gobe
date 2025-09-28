package gdbasez

import (
	"context"

	fscm "github.com/kubex-ecosystem/gdbase/factory/models"
	"gorm.io/gorm"
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

	CronService = fscm.CronJobService
	CronRepo    = fscm.CronJobRepo
	CronModel   = fscm.CronJobModel

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

	WebhookService = fscm.WebhookService
	WebhookRepo    = fscm.WebhookRepo
	WebhookModel   = fscm.Webhook

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

func NewClientRepo(dbConn *gorm.DB) ClientRepo {
	return fscm.NewClientRepo(dbConn)
}

func NewProductService(db ProductRepo) ProductService {
	return fscm.NewProductService(db)
}

func NewProductModel() *ProductModel {
	return &ProductModel{}
}

func NewProductRepo(dbConn *gorm.DB) ProductRepo {
	return fscm.NewProductRepo(dbConn)
}

func NewCronService(db CronRepo) CronService {
	return fscm.NewCronJobService(db)
}

func NewCronModel() *CronModel {
	return &CronModel{}
}

func NewCronRepo(ctx context.Context, dbConn *gorm.DB) CronRepo {
	return fscm.NewCronJobRepo(ctx, dbConn)
}

func NewDiscordService(db DiscordRepo) DiscordService {
	return fscm.NewDiscordService(db)
}

func NewDiscordModel() *DiscordModel {
	return &DiscordModel{}
}

func NewDiscordRepo(dbConn *gorm.DB) DiscordRepo {
	return fscm.NewDiscordRepo(dbConn)
}
func NewJobQueueService(db JobQueueRepo) JobQueueService {
	return fscm.NewJobQueueService(db)
}

func NewJobQueueRepo(dbConn *gorm.DB) JobQueueRepo {
	return fscm.NewJobQueueRepo(dbConn)
}

func NewJobQueueModel() JobQueueModel {
	return fscm.NewJobQueueModel()
}

func NewAnalysisJobService(db AnalysisJobRepo) AnalysisJobService {
	return fscm.NewAnalysisJobService(db)
}

func NewAnalysisJobRepo(dbConn *gorm.DB) AnalysisJobRepo {
	return fscm.NewAnalysisJobRepo(dbConn)
}

func NewAnalysisJobModel() AnalysisJobModel {
	return fscm.NewAnalysisJobModel()
}
