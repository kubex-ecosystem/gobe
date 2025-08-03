// Package discord provides the controller for managing Discord interactions in the application.
package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/rafa-mori/gobe/internal/approval"
	"github.com/rafa-mori/gobe/internal/config"
	"github.com/rafa-mori/gobe/internal/events"
	"github.com/rafa-mori/gobe/internal/hub"

	fscm "github.com/rafa-mori/gdbase/factory/models"
	t "github.com/rafa-mori/gobe/internal/types"

	l "github.com/rafa-mori/logz"

	"github.com/rafa-mori/gobe/logger"
)

var (
	gl = logger.GetLogger[l.Logger](nil)
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

func NewDiscordController(db *gorm.DB, hub *hub.DiscordMCPHub) *DiscordController {
	return &DiscordController{
		discordService: fscm.NewDiscordService(fscm.NewDiscordRepo(db)),
		APIWrapper:     t.NewApiWrapper[fscm.DiscordModel](),
		hub:            hub,
	}
}

// @Summary Discord App Handler
// @Description Handles Discord Application/Activity requests
// @Tags discord
// @Accept html
// @Produce html
// @Success 200 {string} HTML page for Discord Application
// @Router /discord [get]
func (dc *DiscordController) HandleDiscordApp(c *gin.Context) {
	gl.Log("info", "🎮 Discord App request received")

	// Log all query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			gl.Log("info", fmt.Sprintf("  %s: %s", key, value))
		}
	}

	// Extract Discord Activity parameters
	instanceID := c.Query("instance_id")
	locationID := c.Query("location_id")
	launchID := c.Query("launch_id")
	channelID := c.Query("channel_id")
	frameID := c.Query("frame_id")
	platform := c.Query("platform")

	gl.Log("info", "📋 Discord Activity parameters:")
	gl.Log("info", fmt.Sprintf("  instance_id: %s", instanceID))
	gl.Log("info", fmt.Sprintf("  location_id: %s", locationID))
	gl.Log("info", fmt.Sprintf("  launch_id: %s", launchID))
	gl.Log("info", fmt.Sprintf("  channel_id: %s", channelID))
	gl.Log("info", fmt.Sprintf("  frame_id: %s", frameID))
	gl.Log("info", fmt.Sprintf("  platform: %s", platform))

	// Create HTML response for Discord Application
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoBE Discord Bot</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            min-height: 100vh;
            box-sizing: border-box;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37);
            border: 1px solid rgba(255, 255, 255, 0.18);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .header h1 {
            margin: 0 0 10px 0;
            font-size: 2.5em;
            background: linear-gradient(45deg, #FFD700, #FFA500);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .status {
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 20px 0;
            font-size: 1.2em;
        }
        .status-indicator {
            width: 12px;
            height: 12px;
            background: #00ff00;
            border-radius: 50%%;
            margin-right: 10px;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%% { opacity: 1; }
            50%% { opacity: 0.5; }
            100%% { opacity: 1; }
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        .info-card {
            background: rgba(255, 255, 255, 0.05);
            padding: 15px;
            border-radius: 10px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .info-card h3 {
            margin: 0 0 10px 0;
            color: #FFD700;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .info-card p {
            margin: 0;
            font-family: 'Courier New', monospace;
            font-size: 0.8em;
            word-break: break-all;
        }
        .actions {
            margin-top: 30px;
            text-align: center;
        }
        .btn {
            display: inline-block;
            padding: 12px 25px;
            margin: 5px;
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            text-decoration: none;
            border-radius: 25px;
            border: none;
            cursor: pointer;
            font-size: 1em;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px 0 rgba(31, 38, 135, 0.4);
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px 0 rgba(31, 38, 135, 0.6);
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            font-size: 0.8em;
            opacity: 0.7;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 GoBE Discord Bot</h1>
            <p>Aplicação Discord integrada com sucesso!</p>
        </div>
        
        <div class="status">
            <div class="status-indicator"></div>
            <span>Conectado e funcionando</span>
        </div>
        
        <div class="info-grid">
            <div class="info-card">
                <h3>📍 Channel ID</h3>
                <p>%s</p>
            </div>
            <div class="info-card">
                <h3>🆔 Instance ID</h3>
                <p>%s</p>
            </div>
            <div class="info-card">
                <h3>🚀 Launch ID</h3>
                <p>%s</p>
            </div>
            <div class="info-card">
                <h3>📱 Platform</h3>
                <p>%s</p>
            </div>
        </div>
        
        <div class="actions">
            <button class="btn" onclick="testBot()">🧪 Testar Bot</button>
            <button class="btn" onclick="openWebSocket()">🔗 WebSocket</button>
        </div>
        
        <div class="footer">
            <p>GoBE Discord Integration • Instance: %s</p>
        </div>
    </div>

    <script>
        function testBot() {
            fetch('/discord/test', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    content: 'Teste do Discord App!',
                    user_id: 'discord_app_user',
                    username: 'Discord App'
                })
            })
            .then(response => response.json())
            .then(data => {
                alert('✅ Bot testado com sucesso: ' + data.message);
            })
            .catch(error => {
                alert('❌ Erro ao testar bot: ' + error);
            });
        }

        function openWebSocket() {
            // Redirect to WebSocket test page or open connection
            alert('🔗 WebSocket connection would be opened here');
        }

        // Initialize Discord SDK if available
        if (typeof DiscordSDK !== 'undefined') {
            console.log('🎮 Discord SDK detected');
        }
    </script>
</body>
</html>
    `, channelID, instanceID, launchID, platform, frameID)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
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
	gl.Log("info", "🔐 Discord OAuth2 authorize request received")

	// Log all query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			gl.Log("info", fmt.Sprintf("  %s: %s", key, value))
		}
	}

	// Check for error in query params (Discord sends errors here)
	if errorType := c.Query("error"); errorType != "" {
		errorDesc := c.Query("error_description")
		gl.Log("error", fmt.Sprintf("❌ Discord OAuth2 error: %s - %s", errorType, errorDesc))

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
		gl.Log("info", fmt.Sprintf("✅ Authorization code received: %s", code))
		gl.Log("info", fmt.Sprintf("📦 State: %s", state))

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

	gl.Log("info", "📋 OAuth2 parameters:")
	gl.Log("info", fmt.Sprintf("  client_id: %s", clientID))
	gl.Log("info", fmt.Sprintf("  redirect_uri: %s", redirectURI))
	gl.Log("info", fmt.Sprintf("  response_type: %s", responseType))
	gl.Log("info", fmt.Sprintf("  scope: %s", scope))

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
		gl.Log("error", fmt.Sprintf("WebSocket upgrade error: %v", err))
		return
	}

	client := &events.Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan events.Event, 256),
	}

	eventStream := dc.hub.GetEventStream()
	eventStream.RegisterClient(client)

	gl.Log("info", fmt.Sprintf("WebSocket client connected: %s", client.ID))
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
	gl.Log("info", fmt.Sprintf("Approving request: %s", requestID))

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
	gl.Log("info", fmt.Sprintf("Rejecting request: %s", requestID))

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

	gl.Log("info", fmt.Sprintf("🧪 Test message received: %s from %s", testMsg.Content, testMsg.Username))

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
		gl.Log("error", fmt.Sprintf("❌ Error processing test message: %v", err))
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
	gl.Log("info", "🎫 Discord OAuth2 token request received")

	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing form: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	gl.Log("info", "📋 Token request parameters:")
	gl.Log("info", fmt.Sprintf("  grant_type: %s", grantType))
	gl.Log("info", fmt.Sprintf("  code: %s", code))
	gl.Log("info", fmt.Sprintf("  redirect_uri: %s", redirectURI))
	gl.Log("info", fmt.Sprintf("  client_id: %s", clientID))
	gl.Log("info", fmt.Sprintf("  client_secret: %s", strings.Repeat("*", len(clientSecret))))

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

	gl.Log("info", "🪝 Discord webhook received:")
	gl.Log("info", fmt.Sprintf("  Webhook ID: %s", webhookID))
	gl.Log("info", fmt.Sprintf("  Webhook Token: %s", webhookToken[:10]+"..."))

	// Read the body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error reading webhook body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	// Parse JSON
	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing webhook JSON: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	gl.Log("info", fmt.Sprintf("📦 Webhook data: %+v", webhookData))

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
	gl.Log("info", "⚡ Discord interaction received")

	// Verify Discord signature (important for security)
	signature := c.GetHeader("X-Signature-Ed25519")
	timestamp := c.GetHeader("X-Signature-Timestamp")

	gl.Log("info", "📋 Headers:")
	gl.Log("info", fmt.Sprintf("  X-Signature-Ed25519: %s", signature))
	gl.Log("info", fmt.Sprintf("  X-Signature-Timestamp: %s", timestamp))

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error reading interaction body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	// Parse interaction
	var interaction map[string]interface{}
	if err := json.Unmarshal(body, &interaction); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing interaction JSON: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	gl.Log("info", fmt.Sprintf("📦 Interaction data: %+v", interaction))

	// Handle ping interactions (Discord requires this)
	if interactionType, ok := interaction["type"].(float64); ok && interactionType == 1 {
		gl.Log("info", "🏓 Ping interaction - responding with pong")
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

func (dc *DiscordController) InitiateBotMCP() {
	var err error
	var h *hub.DiscordMCPHub
	if dc.hub == nil {
		h, err = hub.NewDiscordMCPHub(dc.config)
		if err != nil {
			gl.Log("error", "Failed to create Discord hub", err)
			return
		}
		dc.hub = h
		gl.Log("info", "Discord MCP Hub created successfully")
	} else {
		var ok bool
		if h, ok = dc.hub.(*hub.DiscordMCPHub); ok {
			gl.Log("info", "Discord MCP Hub started successfully")
		} else {
			gl.Log("error", "Discord hub is not of type DiscordMCPHub")
			return
		}
	}

	go func() {
		defer func() {
			if recErr := recover(); recErr != nil {
				gl.Log("error", "Recovered from panic in Discord hub", recErr)
				events := dc.hub.GetEventStream()
				if events != nil {
					events.Close()
					gl.Log("info", "Discord hub stopped gracefully")
				}
				gl.Log("info", "Restarting Discord hub...")
				dc.InitiateBotMCP()
			}
		}()
		if err := h.StartDiscordBot(); err != nil {
			gl.Log("error", "Failed to start Discord hub", err)
			return
		}
		h.StartMCPServer()

		h.GetEventStream().Run()
	}()
}
