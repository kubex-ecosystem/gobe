// Package oauth provides OAuth2/PKCE HTTP controllers
package oauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"github.com/kubex-ecosystem/gobe/internal/services/oauth"
)

// OAuthController handles OAuth2/PKCE endpoints
type OAuthController struct {
	dbService    *svc.DBServiceImpl
	oauthService oauth.IOAuthService
}

// NewOAuthController creates a new OAuth controller
func NewOAuthController(dbService *svc.DBServiceImpl, oauthService oauth.IOAuthService) *OAuthController {
	return &OAuthController{
		dbService:    dbService,
		oauthService: oauthService,
	}
}

// Authorize handles GET /oauth/authorize
// Generates an authorization code after user authentication
//
// @Summary OAuth2 Authorization Endpoint
// @Description Initiates OAuth2 PKCE flow and generates authorization code
// @Tags oauth
// @Accept json
// @Produce json
// @Param client_id query string true "OAuth Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param code_challenge query string true "PKCE Code Challenge"
// @Param code_challenge_method query string false "PKCE Method (S256 or plain)" default(S256)
// @Param scope query string false "Requested scope"
// @Param state query string false "State parameter"
// @Success 302 {string} string "Redirect with authorization code"
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /oauth/authorize [get]
func (c *OAuthController) Authorize(ctx *gin.Context) {
	clientID := ctx.Query("client_id")
	redirectURI := ctx.Query("redirect_uri")
	codeChallenge := ctx.Query("code_challenge")
	codeChallengeMethod := ctx.DefaultQuery("code_challenge_method", "S256")
	scope := ctx.Query("scope")
	state := ctx.Query("state")

	// Validate required parameters
	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "client_id, redirect_uri, and code_challenge are required",
		})
		return
	}

	// Validate code_challenge_method
	if codeChallengeMethod != "S256" && codeChallengeMethod != "plain" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "code_challenge_method must be 'S256' or 'plain'",
		})
		return
	}

	// TODO: Get authenticated user from session/JWT
	// For now, we'll use a placeholder - in production, this should come from authentication middleware
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":             "unauthorized",
			"error_description": "User must be authenticated",
		})
		return
	}

	// Generate authorization code
	code, err := c.oauthService.GenerateAuthorizationCode(
		ctx,
		userID,
		clientID,
		redirectURI,
		codeChallenge,
		codeChallengeMethod,
		scope,
	)
	if err != nil {
		gl.Log("error", "OAuth Authorize error: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": err.Error(),
		})
		return
	}

	// Build redirect URL with code and state
	redirectURL := redirectURI + "?code=" + code
	if state != "" {
		redirectURL += "&state=" + state
	}

	// Redirect to client with authorization code
	ctx.Redirect(http.StatusFound, redirectURL)
}

// Token handles POST /oauth/token
// Exchanges authorization code for access and refresh tokens
//
// @Summary OAuth2 Token Endpoint
// @Description Exchanges authorization code for tokens using PKCE
// @Tags oauth
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type (authorization_code)"
// @Param code formData string true "Authorization code"
// @Param code_verifier formData string true "PKCE code verifier"
// @Param client_id formData string true "OAuth Client ID"
// @Param redirect_uri formData string true "Redirect URI"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /oauth/token [post]
func (c *OAuthController) Token(ctx *gin.Context) {
	grantType := ctx.PostForm("grant_type")
	code := ctx.PostForm("code")
	codeVerifier := ctx.PostForm("code_verifier")
	clientID := ctx.PostForm("client_id")
	_ = ctx.PostForm("redirect_uri") // Optional: for validation if needed

	// Validate grant_type
	if grantType != "authorization_code" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_grant_type",
			"error_description": "Only 'authorization_code' grant type is supported",
		})
		return
	}

	// Validate required parameters
	if code == "" || codeVerifier == "" || clientID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "code, code_verifier, and client_id are required",
		})
		return
	}

	// Exchange code for tokens
	tokenPair, err := c.oauthService.ExchangeCodeForTokens(ctx, code, codeVerifier, clientID)
	if err != nil {
		gl.Log("error", "OAuth Token error: "+err.Error())

		// Determine error type
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_grant",
				"error_description": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": err.Error(),
		})
		return
	}

	// Return tokens in OAuth2 standard format
	ctx.JSON(http.StatusOK, TokenResponse{
		AccessToken:  tokenPair.IDToken.SS,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // TODO: Get from config
		RefreshToken: tokenPair.RefreshToken.SS,
		Scope:        "", // TODO: Return actual scope
	})
}

// RegisterClient handles POST /oauth/clients
// Registers a new OAuth2 client
//
// @Summary Register OAuth2 Client
// @Description Registers a new OAuth2 client application
// @Tags oauth
// @Accept json
// @Produce json
// @Param payload body RegisterClientRequest true "Client registration data"
// @Success 201 {object} ClientResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /oauth/clients [post]
func (c *OAuthController) RegisterClient(ctx *gin.Context) {
	var req RegisterClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if req.ClientName == "" || len(req.RedirectURIs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "client_name and redirect_uris are required"})
		return
	}

	dbService := c.dbService
	if dbService == nil {
		gl.Log("error", "Database service is nil for OAuthRoutes")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	dbCfg := dbService.GetConfig(ctx)
	if dbCfg == nil {
		gl.Log("error", "Database config is nil for OAuthRoutes")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	dbName := dbCfg.GetDBName()
	ctxB := context.WithValue(ctx, gl.ContextDBNameKey, dbName)

	// Create OAuth client service
	clientService := svc.NewOAuthClientService(ctxB, dbService, dbName)

	// Generate client_id
	clientID := generateClientID()

	// Create client model
	client := svc.NewOAuthClientModel(clientID, req.ClientName, req.RedirectURIs, req.Scopes)

	// Save to database
	created, err := clientService.CreateClient(client)
	if err != nil {
		gl.Log("error", "Failed to register client: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, ClientResponse{
		ClientID:     created.GetClientID(),
		ClientName:   created.GetClientName(),
		RedirectURIs: created.GetRedirectURIs(),
		Scopes:       created.GetScopes(),
		Active:       created.GetActive(),
	})
}

// Helper function to generate client_id
func generateClientID() string {
	// TODO: Implement proper client_id generation
	// For now, use a simple UUID-based approach
	return "client_" + strings.ReplaceAll(svc.NewUserModel("", "", "").GetID(), "-", "")[:16]
}

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// RegisterClientRequest represents the client registration request
type RegisterClientRequest struct {
	ClientName   string   `json:"client_name" binding:"required"`
	RedirectURIs []string `json:"redirect_uris" binding:"required"`
	Scopes       []string `json:"scopes"`
}

// ClientResponse represents the client registration response
type ClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Active       bool     `json:"active"`
}
