// Package discord provides the controller for managing Discord interactions in the application.
package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/rafa-mori/gobe/internal/approval"
	"github.com/rafa-mori/gobe/internal/config"
	"github.com/rafa-mori/gobe/internal/events"

	fscm "github.com/rafa-mori/gdbase/factory/models"
	t "github.com/rafa-mori/gobe/internal/types"
)

type HubInterface interface {
	GetEventStream() *events.Stream
	GetApprovalManager() *approval.Manager
	ProcessMessageWithLLM(ctx context.Context, msg interface{}) error
}

type DiscordController struct {
	discordService fscm.DiscordService
	APIWrapper     *t.APIWrapper[fscm.DiscordModel]
	config         *config.Config
	hub            HubInterface
	upgrader       websocket.Upgrader
}

func NewDiscordController(db *gorm.DB) *DiscordController {
	return &DiscordController{
		discordService: fscm.NewDiscordService(fscm.NewDiscordRepo(db)),
		APIWrapper:     t.NewApiWrapper[fscm.DiscordModel](),
	}
}

// @Summary Discord OAuth2 Authorization
// @Schemes http https
// @Description Initiates the OAuth2 authorization flow for Discord
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {string} Authorization URL
// @Router /discord/authorize [get]
func (dc *DiscordController) HandleDiscordOAuth2Authorize(c *gin.Context) {
	log.Printf("🔐 Discord OAuth2 authorize request received")

	// Log all query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			log.Printf("  %s: %s", key, value)
		}
	}

	// Check for error in query params (Discord sends errors here)
	if errorType := c.Query("error"); errorType != "" {
		errorDesc := c.Query("error_description")
		log.Printf("❌ Discord OAuth2 error: %s - %s", errorType, errorDesc)

		// Return a proper HTML page instead of JSON for browser display
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Discord OAuth2 Error</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 50px; background: #f0f0f0; }
				.container { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
				.error { color: #d32f2f; }
				.suggestion { background: #e3f2fd; padding: 15px; border-radius: 5px; margin-top: 20px; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>🚨 Discord OAuth2 Error</h1>
				<p class="error"><strong>Error:</strong> %s</p>
				<p class="error"><strong>Description:</strong> %s</p>
				
				<div class="suggestion">
					<h3>💡 Para Bots Discord:</h3>
					<p>Se você está tentando adicionar um bot Discord, use esta URL direta:</p>
					<a href="https://discord.com/api/oauth2/authorize?client_id=1344830702780420157&scope=bot&permissions=274877908992" 
					   target="_blank" style="color: #1976d2; text-decoration: none; font-weight: bold;">
						🤖 Clique aqui para convidar o bot
					</a>
					
					<h4>🔧 Ou remova a Redirect URI:</h4>
					<ol>
						<li>Vá para <a href="https://discord.com/developers/applications/1344830702780420157/oauth2/general" target="_blank">Discord Developer Portal</a></li>
						<li>Remova todas as Redirect URIs</li>
						<li>Use apenas URLs de convite diretas para bots</li>
					</ol>
				</div>
			</div>
		</body>
		</html>
		`, errorType, errorDesc)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	// Handle authorization code flow
	code := c.Query("code")
	state := c.Query("state")

	if code != "" {
		log.Printf("✅ Authorization code received: %s", code)
		log.Printf("📦 State: %s", state)

		// In a real app, you'd exchange this code for a token
		// For now, we'll just return success
		c.JSON(http.StatusOK, gin.H{
			"message": "Authorization successful",
			"code":    code,
			"state":   state,
		})
		return
	}

	// If no code and no error, this might be an initial authorization request
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")

	log.Printf("📋 OAuth2 parameters:")
	log.Printf("  client_id: %s", clientID)
	log.Printf("  redirect_uri: %s", redirectURI)
	log.Printf("  response_type: %s", responseType)
	log.Printf("  scope: %s", scope)

	// Return authorization page or redirect to Discord
	c.JSON(http.StatusOK, gin.H{
		"message":      "OAuth2 authorization endpoint",
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
	})
}

// @Summary WebSocket connection
// @Description Upgrades the HTTP connection to a WebSocket connection
// @Tags discord
// @Accept json
// @Produce json
// @Success 101 {string} WebSocket connection established
// @Router /discord/socket [get]
func (dc *DiscordController) HandleWebSocket(c *gin.Context) {
	conn, err := dc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &events.Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan events.Event, 256),
	}

	eventStream := dc.hub.GetEventStream()
	eventStream.RegisterClient(client)

	log.Printf("WebSocket client connected: %s", client.ID)
}

// @Summary Get pending approvals
// @Description Retrieves a list of pending approval requests
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {array} string "Pending approvals"
// @Router /discord/approvals [get]
func (dc *DiscordController) GetPendingApprovals(c *gin.Context) {
	// This would need to be implemented based on your approval manager interface
	c.JSON(http.StatusOK, gin.H{
		"approvals": []interface{}{},
	})
}

// @Summary Approve request
// @Description Approves a pending approval request
// @Tags discord
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {string} Request approved
// @Router /discord/approvals/{id}/approve [post]
func (dc *DiscordController) ApproveRequest(c *gin.Context) {
	requestID := c.Param("id")

	// Mock approval - implement with your approval manager
	log.Printf("Approving request: %s", requestID)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Request approved",
		"request_id": requestID,
	})
}

// @Summary Reject request
// @Description Rejects a pending approval request
// @Tags discord
// @Accept json
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {string} Request rejected
// @Router /discord/approvals/{id}/reject [post]
func (dc *DiscordController) RejectRequest(c *gin.Context) {
	requestID := c.Param("id")

	// Mock rejection - implement with your approval manager
	log.Printf("Rejecting request: %s", requestID)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Request rejected",
		"request_id": requestID,
	})
}

// @Summary Handle test message
// @Description Handles a test message from the user
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {string} Test message processed successfully
// @Router /discord/test [post]
func (dc *DiscordController) HandleTestMessage(c *gin.Context) {
	var testMsg struct {
		Content  string `json:"content"`
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}

	if err := c.ShouldBindJSON(&testMsg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Set defaults
	if testMsg.UserID == "" {
		testMsg.UserID = "test_user_123"
	}
	if testMsg.Username == "" {
		testMsg.Username = "TestUser"
	}

	log.Printf("🧪 Test message received: %s from %s", testMsg.Content, testMsg.Username)

	// Create a mock message object
	mockMessage := map[string]interface{}{
		"content":  testMsg.Content,
		"user_id":  testMsg.UserID,
		"username": testMsg.Username,
		"channel":  "test_channel",
	}

	// Process with the hub
	ctx := context.Background()
	err := dc.hub.ProcessMessageWithLLM(ctx, mockMessage)
	if err != nil {
		log.Printf("❌ Error processing test message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test message processed successfully",
		"content": testMsg.Content,
		"user":    testMsg.Username,
	})
}

// @Summary Handle Discord OAuth2 token
// @Description Handles the OAuth2 token exchange for Discord
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {string} Token exchanged successfully
// @Router /discord/oauth2/token [post]
func (dc *DiscordController) HandleDiscordOAuth2Token(c *gin.Context) {
	log.Printf("🎫 Discord OAuth2 token request received")

	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("❌ Error parsing form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	log.Printf("📋 Token request parameters:")
	log.Printf("  grant_type: %s", grantType)
	log.Printf("  code: %s", code)
	log.Printf("  redirect_uri: %s", redirectURI)
	log.Printf("  client_id: %s", clientID)
	log.Printf("  client_secret: %s", strings.Repeat("*", len(clientSecret)))

	// In a real app, you'd validate these and return a real token
	// For now, return a mock token response
	c.JSON(http.StatusOK, gin.H{
		"access_token":  "mock_access_token",
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": "mock_refresh_token",
		"scope":         "bot identify",
	})
}

// @Summary Handle Discord webhook
// @Description Handles incoming webhook events from Discord
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {string} Webhook processed successfully
// @Router /discord/webhook/{webhookId}/{webhookToken} [post]
func (dc *DiscordController) HandleDiscordWebhook(c *gin.Context) {
	webhookID := c.Param("webhookId")
	webhookToken := c.Param("webhookToken")

	log.Printf("🪝 Discord webhook received:")
	log.Printf("  Webhook ID: %s", webhookID)
	log.Printf("  Webhook Token: %s", webhookToken[:10]+"...")

	// Read the body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	// Parse JSON
	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		log.Printf("❌ Error parsing webhook JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	log.Printf("📦 Webhook data: %+v", webhookData)

	// Process webhook (you can integrate this with your hub)
	// dc.hub.ProcessWebhook(webhookData)

	c.JSON(http.StatusOK, gin.H{"message": "webhook received"})
}

// @Summary Handle Discord interactions
// @Description Handles interactions from Discord
// @Tags discord
// @Accept json
// @Produce json
// @Success 200 {string} Interaction processed successfully
// @Router /discord/interactions [post]
func (dc *DiscordController) HandleDiscordInteractions(c *gin.Context) {
	log.Printf("⚡ Discord interaction received")

	// Verify Discord signature (important for security)
	signature := c.GetHeader("X-Signature-Ed25519")
	timestamp := c.GetHeader("X-Signature-Timestamp")

	log.Printf("📋 Headers:")
	log.Printf("  X-Signature-Ed25519: %s", signature)
	log.Printf("  X-Signature-Timestamp: %s", timestamp)

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Error reading interaction body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	// Parse interaction
	var interaction map[string]interface{}
	if err := json.Unmarshal(body, &interaction); err != nil {
		log.Printf("❌ Error parsing interaction JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	log.Printf("📦 Interaction data: %+v", interaction)

	// Handle ping interactions (Discord requires this)
	if interactionType, ok := interaction["type"].(float64); ok && interactionType == 1 {
		log.Printf("🏓 Ping interaction - responding with pong")
		c.JSON(http.StatusOK, gin.H{"type": 1})
		return
	}

	// Handle other interactions
	c.JSON(http.StatusOK, gin.H{
		"type": 4, // CHANNEL_MESSAGE_WITH_SOURCE
		"data": gin.H{
			"content": "Hello from Discord MCP Hub! 🤖",
		},
	})
}
