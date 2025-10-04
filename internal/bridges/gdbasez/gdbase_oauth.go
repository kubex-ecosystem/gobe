package gdbasez

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	models "github.com/kubex-ecosystem/gdbase/factory/models"
)

// OAuth Client type aliases - clean abstraction without leaking implementation
type (
	OAuthClient        = models.OAuthClient
	OAuthClientModel   = models.OAuthClientModel
	OAuthClientRepo    = models.OAuthClientRepo
	OAuthClientService = models.OAuthClientService
)

// Auth Code type aliases - clean abstraction without leaking implementation
type (
	AuthCode        = models.AuthCode
	AuthCodeModel   = models.AuthCodeModel
	AuthCodeRepo    = models.AuthCodeRepo
	AuthCodeService = models.AuthCodeService
)

// NewOAuthClientService creates a new OAuth client service
// Note: This function still accepts *gorm.DB but this is the ONLY place where it's needed
// All other code uses only interfaces
func NewOAuthClientService(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) OAuthClientService {
	repo := models.NewOAuthClientRepo(ctx, dbService, dbName)
	return models.NewOAuthClientService(repo)
}

// NewAuthCodeService creates a new authorization code service
// Note: This function still accepts *gorm.DB but this is the ONLY place where it's needed
// All other code uses only interfaces
func NewAuthCodeService(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) AuthCodeService {
	repo := models.NewAuthCodeRepo(ctx, dbService, dbName)
	return models.NewAuthCodeService(repo)
}

// NewOAuthClientModel creates a new OAuth client model
func NewOAuthClientModel(clientID, clientName string, redirectURIs, scopes []string) OAuthClient {
	return models.NewOAuthClientModel(clientID, clientName, redirectURIs, scopes)
}
