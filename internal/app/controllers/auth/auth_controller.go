// Package auth provides authentication controllers
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"golang.org/x/crypto/bcrypt"
)

type (
	// ErrorResponse padroniza respostas de erro para endpoints Discord.
	ErrorResponse = t.ErrorResponse
)

// AuthController handles authentication routes
type AuthController struct {
	authMiddleware *middlewares.WebAuthMiddleware
	users          map[string]User // In-memory for now - use DB in production
}

// User represents a user account
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Email        string
	DiscordID    string
	CreatedAt    time.Time
}

// NewAuthController creates a new auth controller
func NewAuthController(authMiddleware *middlewares.WebAuthMiddleware) *AuthController {
	controller := &AuthController{
		authMiddleware: authMiddleware,
		users:          make(map[string]User),
	}

	// Create default admin user (for demo/dev)
	controller.createDefaultUser()

	return controller
}

// createDefaultUser creates a default admin user
func (ac *AuthController) createDefaultUser() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	ac.users["admin"] = User{
		ID:           "1",
		Username:     "admin",
		PasswordHash: string(hash),
		Email:        "admin@kubex.io",
		CreatedAt:    time.Now(),
	}

	gl.Log("info", "Default admin user created", "username", "admin", "password", "admin")
}

// Login endpoint handles POST /auth/login
//
// @Summary     Login
// @Description Realiza o login de um usuário.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body  body     map[string]interface{}  true  "Login request"
// @Success     200   {object} map[string]interface{}
// @Failure     400   {object} ErrorResponse
// @Failure     401   {object} ErrorResponse
// @Router      /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Find user
	user, exists := ac.users[req.Username]
	if !exists {
		gl.Log("warn", "Login attempt for non-existent user", "username", req.Username, "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": "Invalid username or password",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		gl.Log("warn", "Failed login attempt", "username", req.Username, "ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := ac.authMiddleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		gl.Log("error", "Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": "Please try again",
		})
		return
	}

	gl.Log("info", "User logged in successfully", "username", user.Username, "ip", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Logout endpoint handles POST /auth/logout
//
// @Summary     Logout
// @Description Realiza o logout de um usuário.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body  body     map[string]interface{}  true  "Logout request"
// @Success     200   {object} map[string]interface{}
// @Failure     400   {object} ErrorResponse
// @Failure     401   {object} ErrorResponse
// @Router      /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// Clear cookie
	c.SetCookie("gobe_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// Me endpoint handles GET /auth/me
//
// @Summary     Me
// @Description Retorna informações do usuário autenticado.
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} ErrorResponse
// @Router      /auth/me [get]
func (ac *AuthController) Me(c *gin.Context) {
	authenticated, _ := c.Get("authenticated")
	if !authenticated.(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         "Not authenticated",
			"authenticated": false,
		})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       userID,
			"username": username,
		},
	})
}

// DiscordOAuth handles GET /auth/discord
func (ac *AuthController) DiscordOAuth(c *gin.Context) {
	// Generate state token for CSRF protection
	stateToken := generateStateToken()

	// Store state in session/cookie
	c.SetCookie("oauth_state", stateToken, 300, "/", "", false, true)

	// Discord OAuth URL
	discordClientID := "YOUR_DISCORD_CLIENT_ID" // TODO: Get from config
	redirectURI := c.Query("redirect")
	if redirectURI == "" {
		redirectURI = "http://localhost:3666/auth/discord/callback"
	}

	oauthURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify%%20email&state=%s",
		discordClientID,
		redirectURI,
		stateToken,
	)

	c.Redirect(http.StatusTemporaryRedirect, oauthURL)
}

// DiscordCallback handles GET /auth/discord/callback
func (ac *AuthController) DiscordCallback(c *gin.Context) {
	// Verify state token
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid state token",
			"message": "CSRF validation failed",
		})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No authorization code",
			"message": "OAuth flow failed",
		})
		return
	}

	// TODO: Exchange code for access token
	// TODO: Fetch user info from Discord API
	// TODO: Create or update user in database
	// TODO: Generate JWT token
	// TODO: Redirect to app with token

	c.JSON(http.StatusOK, gin.H{
		"message": "Discord OAuth not fully implemented yet",
		"code":    code,
	})
}

// Register endpoint handles POST /auth/register
//
// @Summary     Register
// @Description Registra um novo usuário.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body  body     map[string]interface{}  true  "Register request"
// @Success     201   {object} map[string]interface{}
// @Failure     400   {object} ErrorResponse
// @Failure     409   {object} ErrorResponse
// @Router      /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Check if user exists
	if _, exists := ac.users[req.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "User already exists",
			"message": "Username is already taken",
		})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Could not hash password",
		})
		return
	}

	// Create user
	user := User{
		ID:           generateUserID(),
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		CreatedAt:    time.Now(),
	}

	ac.users[req.Username] = user

	gl.Log("info", "New user registered", "username", user.Username, "email", user.Email)

	// Generate token
	token, err := ac.authMiddleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token generation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Helper functions

func generateStateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
