// Package web provides web UI routes with OAuth 2.1 authentication
package web

import (
	"os"

	"github.com/gin-gonic/gin"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	"github.com/kubex-ecosystem/gobe/internal/app/router/proxy"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
)

// SetupWebRoutes configures all web UI routes with OAuth 2.1 authentication
func SetupWebRoutes(router *gin.RouterGroup, dbService *svc.DBServiceImpl) error {
	gl.Log("info", "Setting up web routes with OAuth 2.1 authentication")

	// Initialize OAuth-based authentication middleware
	tokenService, certService, err := middlewares.NewTokenService(dbService)
	if err != nil {
		gl.Log("error", "Failed to initialize token service for web routes", err)
		return err
	}

	authMiddleware := middlewares.NewAuthenticationMiddleware(tokenService, certService, nil)
	authInstance := &middlewares.AuthenticationMiddleware{
		CertService:  certService,
		TokenService: tokenService,
	}

	// Web group - all routes require OAuth authentication
	webGroup := router.Group("/frontend")

	// Apply JWT validation for all web routes
	webGroup.Use(authMiddleware, authInstance.ValidateJWT(func(c *gin.Context) {
		c.Next()
	}))

	// Serve GoBE Dashboard (static files)
	webGroup.Static("/dashboard", "./frontend/dist")
	webGroup.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/frontend/dashboard")
	})

	// Initialize proxy router for ecosystem services
	proxyConfig := getProxyConfig()
	proxyRouter, err := proxy.NewWebProxyRouter(proxyConfig)
	if err != nil {
		gl.Log("error", "Failed to initialize proxy router", err)
		return err
	}

	// Register proxy routes (Grompt, Analyzer, GemX)
	proxyRouter.RegisterRoutes(webGroup)

	gl.Log("info", "Web routes configured successfully",
		"grompt_url", proxyConfig.GromptURL,
		"analyzer_url", proxyConfig.AnalyzerURL,
		"gemx_url", proxyConfig.GemXURL)

	return nil
}

// getProxyConfig loads proxy configuration from environment or defaults
func getProxyConfig() proxy.ProxyConfig {
	config := proxy.DefaultProxyConfig()

	// Override from environment variables if present
	if gromptURL := os.Getenv("GROMPT_URL"); gromptURL != "" {
		config.GromptURL = gromptURL
	}
	if analyzerURL := os.Getenv("ANALYZER_URL"); analyzerURL != "" {
		config.AnalyzerURL = analyzerURL
	}
	if gemxURL := os.Getenv("GEMX_URL"); gemxURL != "" {
		config.GemXURL = gemxURL
	}

	return config
}
