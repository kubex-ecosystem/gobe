// Package oauth provides OAuth2/PKCE routes for the application
package oauth

import (
	"net/http"
	"os"

	"github.com/kubex-ecosystem/gobe/internal/app/controllers/sys/oauth"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	sau "github.com/kubex-ecosystem/gobe/factory/security"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"
	oauthsvc "github.com/kubex-ecosystem/gobe/internal/services/oauth"
)

// OAuthRoutes holds the OAuth router
type OAuthRoutes struct {
	ar.IRouter
}

// NewOAuthRoutes creates and returns OAuth2/PKCE routes
func NewOAuthRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for OAuthRoutes")
		return nil
	}
	rtl := *rtr

	// Get database service from router
	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for OAuthRoutes")
		return nil
	}

	// Get GORM DB connection
	dbGorm, err := dbService.GetDB(nil, gdbasez.DefaultDBName)
	if err != nil {
		gl.Log("error", "Failed to get DB from service for OAuthRoutes", err)
		return nil
	}

	// Create OAuth services via bridge (clean abstraction)
	oauthClientService := gdbasez.NewOAuthClientService(dbGorm)
	authCodeService := gdbasez.NewAuthCodeService(dbGorm)

	// Create UserService
	userRepo := gdbasez.NewUserRepo(dbGorm)
	userService := gdbasez.NewUserService(userRepo)

	// Create TokenService (same pattern as user routes)
	certService := crt.NewCertService(
		os.ExpandEnv(cm.DefaultGoBEKeyPath),
		os.ExpandEnv(cm.DefaultGoBECertPath),
	)
	tokenClient := sau.NewTokenClient(certService, dbService)
	tokenService, _, _, err := tokenClient.LoadTokenCfg()
	if err != nil {
		gl.Log("error", "Failed to load token config for OAuthRoutes", err)
		return nil
	}

	// Create OAuth service (business logic)
	oauthService := oauthsvc.NewOAuthService(oauthClientService, authCodeService, userService, tokenService)

	// Create controller
	oauthController := oauth.NewOAuthController(dbGorm, oauthService)

	// Prepare routes map
	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := rtl.GetMiddlewares()

	// Public routes (no authentication required)
	routesMap["OAuthAuthorize"] = proto.NewRoute(
		http.MethodGet,
		"/oauth/authorize",
		"application/json",
		oauthController.Authorize,
		nil, // No middlewares for now - TODO: Add user authentication middleware
		dbService,
		nil,
		nil,
	)

	routesMap["OAuthToken"] = proto.NewRoute(
		http.MethodPost,
		"/oauth/token",
		"application/x-www-form-urlencoded",
		oauthController.Token,
		nil, // Public endpoint
		dbService,
		nil,
		nil,
	)

	// Admin routes (require authentication)
	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = true

	routesMap["OAuthRegisterClient"] = proto.NewRoute(
		http.MethodPost,
		"/oauth/clients",
		"application/json",
		oauthController.RegisterClient,
		middlewaresMap, // Requires authentication
		dbService,
		secureProperties,
		nil,
	)

	gl.Log("info", "OAuth routes registered successfully")
	return routesMap
}
