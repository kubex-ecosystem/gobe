package gdbasez

import (
	svc "github.com/kubex-ecosystem/gdbase/factory"
	fscm "github.com/kubex-ecosystem/gdbase/factory/models"
)

type (
	JSONB      = svc.JSONB
	JSONBData  = svc.JSONBData
	IJSONBData = svc.JSONBData
	JSONBImpl  = svc.JSONBData

	// Client definitions

	ClientService = fscm.ClientService
	ClientModel   = fscm.ClientModel
	ClientRepo    = fscm.ClientRepo

	// Product definitions

	ProductModel   = fscm.ProductModel
	ProductRepo    = fscm.ProductRepo
	ProductService = fscm.ProductService

	// Cron definitions

	CronJobServiceImpl = fscm.CronJobServiceImpl
	CronJobService     = fscm.CronJobService
	CronJobRepoImpl    = fscm.CronJobRepoImpl
	CronJobRepo        = fscm.CronJobRepo
	CronJobModel       = fscm.CronJobModel

	// Discord definitions

	DiscordService   = fscm.DiscordService
	DiscordRepo      = fscm.DiscordRepo
	DiscordModelImpl = fscm.DiscordModel
	DiscordModel     = fscm.DiscordModelInterface

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

	OAuthClient        = fscm.OAuthClient
	OAuthClientRepo    = fscm.OAuthClientRepo
	OAuthClientService = fscm.OAuthClientService

	AuthCode        = fscm.AuthCode
	AuthCodeRepo    = fscm.AuthCodeRepo
	AuthCodeService = fscm.AuthCodeService

	// OrderModel   = fscm.OrderModel
	// OrderRepo    = fscm.OrderRepo
	// OrderService = fscm.OrderService

	// ContactModel   = fscm.ContactModel
	// ContactRepo    = fscm.ContactRepo
	// ContactService = fscm.ContactService
)
