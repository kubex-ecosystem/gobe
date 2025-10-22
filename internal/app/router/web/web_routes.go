// Package web provides web UI routes with OAuth 2.1 authentication
package web

import (
	"os"

	"github.com/gin-gonic/gin"
	gl "github.com/kubex-ecosystem/logz/logger"

	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	"github.com/kubex-ecosystem/gobe/internal/app/router/proxy"
	gui "github.com/kubex-ecosystem/gobe/internal/app/web"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
)

var (
	webGUIFiles *gui.GUIGoBE = gui.NewGUIGoBE()
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

	//authMiddleware := middlewares.NewAuthenticationMiddleware(tokenService, certService, nil)
	authInstance := &middlewares.AuthenticationMiddleware{
		CertService:  certService,
		TokenService: tokenService,
	}

	// serveIndex := func(c *gin.Context) {
	// 	data, err := webGUIFiles.ReadFile("index.html")
	// 	if err != nil {
	// 		gl.Log("error", "Failed to read index.html", err)
	// 		c.String(http.StatusInternalServerError, "Internal Server Error")
	// 		return
	// 	}
	// 	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	// }

	// serveAsset := func(c *gin.Context, requestedPath string) {
	// 	requestedPath = strings.TrimPrefix(requestedPath, "/")
	// 	if requestedPath == "" {
	// 		c.Status(http.StatusBadRequest)
	// 		return
	// 	}
	// 	if !webGUIFiles.Exists(requestedPath) {
	// 		c.Status(http.StatusNotFound)
	// 		return
	// 	}

	// 	file, err := webGUIFiles.OpenFile(requestedPath)
	// 	if err != nil {
	// 		gl.Log("error", "Failed to open asset file", "file", requestedPath, "error", err)
	// 		c.String(http.StatusInternalServerError, "Internal Server Error")
	// 		return
	// 	}
	// 	defer file.Close()

	// 	content, err := io.ReadAll(file)
	// 	if err != nil {
	// 		gl.Log("error", "Failed to read asset file", "file", requestedPath, "error", err)
	// 		c.String(http.StatusInternalServerError, "Internal Server Error")
	// 		return
	// 	}

	// 	ext := filepath.Ext(requestedPath)
	// 	contentType := mime.TypeByExtension(ext)
	// 	if contentType == "" {
	// 		contentType = http.DetectContentType(content)
	// 	}
	// 	c.Data(http.StatusOK, contentType, content)
	// }

	// assetHandler := func(c *gin.Context) {
	// 	requestedPath := c.Param("filepath")
	// 	serveAsset(c, requestedPath)
	// }

	// router.GET("/assets/*filepath", assetHandler)
	// // router.GET("/", serveIndex)
	// router.GET("/app", serveIndex)
	// // router.GET("/app/", serveIndex)
	// router.GET("/app/*path", func(c *gin.Context) {
	// 	path := strings.TrimPrefix(c.Param("path"), "/")
	// 	if strings.HasPrefix(path, "assets/") {
	// 		serveAsset(c, strings.TrimPrefix(path, ""))
	// 		return
	// 	}
	// 	serveIndex(c)
	// })

	// Authenticated group for proxied services
	webGroup := router.Group("/")
	webGroup.Use(authInstance.ValidateJWT(func(c *gin.Context) {
		c.Next()
	}))

	// Initialize proxy router for ecosystem services
	proxyConfig := getProxyConfig()
	_, err = proxy.NewWebProxyRouter(proxyConfig)
	if err != nil {
		gl.Log("error", "Failed to initialize proxy router", err)
		return err
	}

	// // Register proxy routes (Grompt, Analyzer, GemX)
	// proxyRouter.RegisterRoutes(webGroup)

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
