// Package discord provides the controller for managing Discord interactions in the application.
package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/kubex-ecosystem/gobe/internal/config"
	"github.com/kubex-ecosystem/gobe/internal/observers/approval"
	"github.com/kubex-ecosystem/gobe/internal/observers/events"
	"github.com/kubex-ecosystem/gobe/internal/proxy/hub"
	"github.com/kubex-ecosystem/gobe/internal/services/chatbot/discord"

	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"

	l "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"github.com/kubex-ecosystem/gobe/internal/module/logger"
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
	discordService svc.DiscordService
	APIWrapper     *t.APIWrapper[svc.DiscordModel]
	config         *config.Config
	hub            HubInterface
	upgrader       websocket.Upgrader
}

type (
	// ErrorResponse padroniza respostas de erro para endpoints Discord.
	ErrorResponse = t.ErrorResponse
)

// DiscordWebhookEvent descreve o payload básico recebido via webhook.
type DiscordWebhookEvent map[string]any

// DiscordInteractionEvent descreve interações enviadas pela API do Discord.
type DiscordInteractionEvent map[string]any

// DiscordOAuthTokenResponse documenta o retorno mock do fluxo OAuth2.
type DiscordOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// DiscordOAuthAuthorizeResponse registra payloads informativos da autorização.
type DiscordOAuthAuthorizeResponse struct {
	Message     string `json:"message"`
	Code        string `json:"code,omitempty"`
	State       string `json:"state,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// DiscordApprovalList lista solicitações pendentes.
type DiscordApprovalList struct {
	Approvals []interface{} `json:"approvals"`
}

// DiscordActionResponse descreve respostas de aprovação/rejeição.
type DiscordActionResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// DiscordTestMessageRequest representa o payload de teste.
type DiscordTestMessageRequest struct {
	Content  string `json:"content"`
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
}

// DiscordTestResponse agrega dados retornados pelo teste.
type DiscordTestResponse struct {
	Message string `json:"message"`
	Content string `json:"content"`
	User    string `json:"user"`
}

// DiscordWebhookAck confirma recebimento de eventos.
type DiscordWebhookAck struct {
	Message string `json:"message"`
}

// DiscordPingResponse resume o status dos pings.
type DiscordPingResponse struct {
	Message string `json:"message"`
}

// DiscordInteractionResponse representa a resposta padrão do Discord.
type DiscordInteractionResponse struct {
	Type int                    `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

func NewDiscordController(db *gorm.DB, hub *hub.DiscordMCPHub, config *config.Config) *DiscordController {
	return &DiscordController{
		discordService: svc.NewDiscordService(svc.NewDiscordRepo(db)),
		APIWrapper:     t.NewAPIWrapper[svc.DiscordModel](),
		hub:            hub,
		config:         config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Permitir origens do Discord durante desenvolvimento
				origin := r.Header.Get("Origin")
				gl.Log("info", fmt.Sprintf("WebSocket origin: %s", origin))

				// ✅ Para desenvolvimento, ser mais permissivo
				if config.DevMode {
					gl.Log("info", "🔧 Dev mode: allowing all WebSocket origins")
					return true
				}

				// Para desenvolvimento, permita origens do Discord
				allowedOrigins := []string{
					"https://discord.com",
					"https://ptb.discord.com",
					"https://canary.discord.com",
					"null", // Para local development
					"",     // Para requests sem Origin header
				}

				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}

				// Para desenvolvimento local, permita localhost
				if strings.Contains(origin, "localhost") ||
					strings.Contains(origin, "127.0.0.1") ||
					strings.Contains(origin, "192.168.") {
					return true
				}

				gl.Log("warn", fmt.Sprintf("🚫 WebSocket origin rejected: %s", origin))
				return false
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// ✅ Adicionar configurações extras para Discord
			EnableCompression: true,
		},
	}
}

// HandleDiscordApp serve a página HTML da activity.
//
// @Summary     Servir activity Discord
// @Description Entrega o bundle HTML da activity Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     html
// @Success     200 {string} string "HTML da activity"
// @Router      /api/v1/discord [get]
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

	c.Header("Content-Type", "text/html; charset=utf-8")

	// Serve the HTML response (This work like a charm on Discord)
	c.File("./web/index.html")

	// Alternatively, if you want to return the HTML as a string: (Discord does not support this method pretty well)
	// htmlFile, err := os.ReadFile("./web/index.html")
	// if err != nil {
	// 	gl.Log("error", fmt.Sprintf("❌ Failed to read HTML file: %v", err))
	// 	c.String(http.StatusInternalServerError, "Internal Server Error")
	// 	return
	// }
	// // Create HTML response for Discord Application
	// html := fmt.Sprintf(string(htmlFile), channelID, instanceID, launchID, platform, frameID)
	// c.String(http.StatusOK, html)
}

// HandleDiscordOAuth2Authorize gerencia o fluxo de autorização Discord OAuth2.
//
// @Summary     Iniciar OAuth2 Discord
// @Description Controla respostas HTML/JSON para autorizações OAuth2 do Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     html
// @Produce     json
// @Success     200 {object} DiscordOAuthAuthorizeResponse
// @Router      /api/v1/discord/oauth2/authorize [get]
// @Router      /api/v1/discord/oauth2/authorize [post]
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
		c.JSON(http.StatusOK, DiscordOAuthAuthorizeResponse{
			Message: "Authorization successful",
			Code:    code,
			State:   state,
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
	c.JSON(http.StatusOK, DiscordOAuthAuthorizeResponse{
		Message:     "OAuth2 authorization endpoint",
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
	})
	//https://discord.com/api/webhooks/1381317940649132162/KZro3msMCG1h_jl_eW-EGXPIldUpbRf8R0DC04bpFRcSOC4ZeW1HzMAGDvNdiO1jVcKj
}

// HandleWebSocket atualiza a conexão HTTP para WebSocket.
//
// @Summary     Abrir WebSocket Discord
// @Description Estabelece canal WebSocket com o hub MCP para eventos Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Success     101 {string} string "WebSocket connection established"
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/discord/websocket [get]
func (dc *DiscordController) HandleWebSocket(c *gin.Context) {
	gl.Log("info", "🔌 WebSocket upgrade attempt")
	gl.Log("info", fmt.Sprintf("  Origin: %s", c.GetHeader("Origin")))
	gl.Log("info", fmt.Sprintf("  User-Agent: %s", c.GetHeader("User-Agent")))
	gl.Log("info", fmt.Sprintf("  Upgrade: %s", c.GetHeader("Upgrade")))
	gl.Log("info", fmt.Sprintf("  Connection: %s", c.GetHeader("Connection")))

	// ✅ Verificar headers WebSocket
	if c.GetHeader("Upgrade") != "websocket" {
		gl.Log("error", "❌ Missing or invalid Upgrade header")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  "error",
			Message: "invalid websocket upgrade request",
		})
		return
	}

	conn, err := dc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ WebSocket upgrade error: %v", err))
		// ✅ Não retornar JSON após upgrade failure
		return
	}
	defer conn.Close()

	client := &events.Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan events.Event, 256),
	}

	// Verificar se o hub está disponível
	if dc.hub == nil {
		gl.Log("error", "❌ Discord hub is not initialized")
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "hub not initialized"}`))
		return
	}

	eventStream := dc.hub.GetEventStream()
	if eventStream == nil {
		gl.Log("error", "❌ Event stream is not available")
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "event stream not available"}`))
		return
	}

	eventStream.RegisterClient(client)
	gl.Log("info", fmt.Sprintf("✅ WebSocket client connected: %s", client.ID))

	// Enviar mensagem de confirmação
	welcomeMsg := map[string]interface{}{
		"type":      "connection",
		"status":    "connected",
		"client_id": client.ID,
		"message":   "WebSocket connected successfully",
		"timestamp": time.Now().Unix(),
	}

	if msgBytes, err := json.Marshal(welcomeMsg); err == nil {
		conn.WriteMessage(websocket.TextMessage, msgBytes)
	}

	// ✅ Goroutine para enviar mensagens do canal
	go func() {
		defer eventStream.UnregisterClient(client)
		for {
			for event := range client.Send {
				if msgBytes, err := json.Marshal(event); err == nil {
					conn.WriteMessage(websocket.TextMessage, msgBytes)
				}
			}
		}
	}()

	// Loop para manter a conexão viva e processar mensagens
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			gl.Log("info", fmt.Sprintf("WebSocket client %s disconnected: %v", client.ID, err))
			break
		}

		if messageType == websocket.TextMessage {
			gl.Log("info", fmt.Sprintf("📨 WebSocket message from %s: %s", client.ID, string(message)))

			// ✅ Processar mensagem recebida
			var msgData map[string]interface{}
			if err := json.Unmarshal(message, &msgData); err == nil {
				if msgType, ok := msgData["type"].(string); ok {
					switch msgType {
					case "ping":
						response := map[string]interface{}{
							"type":      "pong",
							"timestamp": time.Now().Unix(),
						}
						if respBytes, err := json.Marshal(response); err == nil {
							conn.WriteMessage(websocket.TextMessage, respBytes)
						}
					case "test":
						response := map[string]interface{}{
							"type":    "test_response",
							"message": "WebSocket is working!",
							"echo":    msgData,
						}
						if respBytes, err := json.Marshal(response); err == nil {
							conn.WriteMessage(websocket.TextMessage, respBytes)
						}
					}
				}
			}
		}
	}
}

// GetPendingApprovals retorna solicitações aguardando aprovação manual.
//
// @Summary     Listar aprovações pendentes
// @Description Retorna solicitações aguardando aprovação manual. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Success     200 {object} DiscordApprovalList
// @Router      /api/v1/discord/approvals [post]
// @Router      /api/v1/discord/interactions/pending [post]
func (dc *DiscordController) GetPendingApprovals(c *gin.Context) {
	// This would need to be implemented based on your approval manager interface
	c.JSON(http.StatusOK, DiscordApprovalList{Approvals: []interface{}{}})
}

// ApproveRequest confirma manualmente uma solicitação pendente.
//
// @Summary     Aprovar solicitação Discord
// @Description Confirma solicitações pendentes para liberar ações no hub. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Param       id query string false "ID da solicitação"
// @Success     200 {object} DiscordActionResponse
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/discord/approve [post]
func (dc *DiscordController) ApproveRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		requestID = c.Query("id")
	}
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "missing request id"})
		return
	}

	// Mock approval - implement with your approval manager
	gl.Log("info", fmt.Sprintf("Approving request: %s", requestID))

	c.JSON(http.StatusOK, DiscordActionResponse{Message: "Request approved", RequestID: requestID})
}

// RejectRequest registra a rejeição de uma solicitação pendente.
//
// @Summary     Rejeitar solicitação Discord
// @Description Marca a solicitação como rejeitada no fluxo de aprovação. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Param       id query string false "ID da solicitação"
// @Success     200 {object} DiscordActionResponse
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/discord/reject [post]
func (dc *DiscordController) RejectRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		requestID = c.Query("id")
	}
	if requestID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "missing request id"})
		return
	}

	// Mock rejection - implement with your approval manager
	gl.Log("info", fmt.Sprintf("Rejecting request: %s", requestID))

	c.JSON(http.StatusOK, DiscordActionResponse{Message: "Request rejected", RequestID: requestID})
}

// HandleTestMessage injeta mensagens de teste no hub Discord.
//
// @Summary     Simular mensagem Discord
// @Description Envia uma mensagem de teste para validar o pipeline MCP. [Em desenvolvimento]
// @Tags        discord beta
// @Accept      json
// @Produce     json
// @Param       payload body DiscordTestMessageRequest true "Conteúdo da mensagem"
// @Success     200 {object} DiscordTestResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/discord/test [post]
func (dc *DiscordController) HandleTestMessage(c *gin.Context) {
	var testMsg DiscordTestMessageRequest

	if err := c.ShouldBindJSON(&testMsg); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid JSON"})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "processing failed"})
		return
	}

	c.JSON(http.StatusOK, DiscordTestResponse{
		Message: "Test message processed successfully",
		Content: testMsg.Content,
		User:    testMsg.Username,
	})
}

// HandleDiscordOAuth2Token executa a troca de token OAuth2.
//
// @Summary     Trocar token OAuth2 Discord
// @Description Realiza a troca do código por token em ambiente de testes. [Em desenvolvimento]
// @Tags        discord beta
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Success     200 {object} DiscordOAuthTokenResponse
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/discord/oauth2/token [get]
// @Router      /api/v1/discord/oauth2/token [post]
func (dc *DiscordController) HandleDiscordOAuth2Token(c *gin.Context) {
	gl.Log("info", "🎫 Discord OAuth2 token request received")

	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing form: %v", err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid_request"})
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
	c.JSON(http.StatusOK, DiscordOAuthTokenResponse{
		AccessToken:  "mock_access_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "mock_refresh_token",
		Scope:        "bot identify",
	})
}

// HandleDiscordWebhook recebe eventos do webhook Discord.
//
// @Summary     Receber webhook Discord
// @Description Processa eventos recebidos do webhook oficial do Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Accept      json
// @Produce     json
// @Param       webhookId    path string              true "ID do webhook"
// @Param       webhookToken path string              true "Token do webhook"
// @Param       payload      body DiscordWebhookEvent true "Evento do Discord"
// @Success     200 {object} DiscordWebhookAck
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/discord/webhook/{webhookId}/{webhookToken} [post]
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid_body"})
		return
	}

	// Parse JSON
	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing webhook JSON: %v", err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid_json"})
		return
	}

	gl.Log("info", fmt.Sprintf("📦 Webhook data: %+v", webhookData))

	// Process webhook (you can integrate this with your hub)
	// dc.hub.ProcessWebhook(webhookData)

	c.JSON(http.StatusOK, DiscordWebhookAck{Message: "webhook received"})
}

// HandleDiscordInteractions processa interações recebidas do Discord.
//
// @Summary     Processar interação Discord
// @Description Responde PINGs e interações de componentes enviados pelo Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Accept      json
// @Produce     json
// @Param       payload body DiscordInteractionEvent true "Interação recebida"
// @Success     200 {object} DiscordInteractionResponse
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/discord/interactions [post]
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid_body"})
		return
	}

	// Parse interaction
	var interaction map[string]interface{}
	if err := json.Unmarshal(body, &interaction); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Error parsing interaction JSON: %v", err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid_json"})
		return
	}

	gl.Log("info", fmt.Sprintf("📦 Interaction data: %+v", interaction))

	// Handle ping interactions (Discord requires this)
	if interactionType, ok := interaction["type"].(float64); ok && interactionType == 1 {
		gl.Log("info", "🏓 Ping interaction - responding with pong")

		// ✅ RESPOSTA CORRETA PARA PING
		c.JSON(http.StatusOK, DiscordInteractionResponse{
			Type: 1, // PONG response
		})
		return
	}

	// Handle other interactions
	c.JSON(http.StatusOK, DiscordInteractionResponse{
		Type: 4, // CHANNEL_MESSAGE_WITH_SOURCE
		Data: map[string]interface{}{
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
		// Start the Discord bot connection (if applicable)
		if h == nil {
			gl.Log("error", "Discord hub is nil, cannot start bot")
			return
		}
		// Note: Starting the bot connection is handled inside StartMCPServer now
		// to avoid multiple connections in case of restarts.
		// Uncomment if you want to start the bot separately.
		if err := h.StartDiscordBot(); err != nil {
			gl.Log("error", "Failed to start Discord hub", err)
			return
		}
		h.StartMCPServer()

		h.GetEventStream().Run()
	}()
}

// PingAdapter verifica o estado do hub conectado.
//
// @Summary     Ping hub Discord
// @Description Checa se o hub MCP em execução está respondendo. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Success     200 {object} DiscordPingResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/discord/ping [get]
func (dc *DiscordController) PingAdapter(c *gin.Context) {
	hd := dc.hub
	if hd == nil {
		gl.Log("error", "Failed to ping Discord adapter")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to ping Discord adapter"})
		return
	}
	gl.Log("info", "Discord adapter pinged successfully")
	c.JSON(http.StatusOK, DiscordPingResponse{Message: "Discord adapter pinged successfully"})
}

// PingDiscordAdapter dispara ping ativo direto ao Discord.
//
// @Summary     Ping ativo Discord
// @Description Realiza ping utilizando o adaptador direto do Discord. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Param       msg query string false "Mensagem customizada"
// @Success     200 {object} DiscordPingResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/discord/ping [post]
func (dc *DiscordController) PingDiscordAdapter(c *gin.Context) {
	var cfg *config.Config
	var err error
	if dc.config != nil {
		cfg = dc.config
	} else {
		cfg, err = config.Load[*config.Config](kbx.InitArgs{
			ConfigFile: "./config/discord.json",
		})
		if err != nil {
			gl.Log("error", "Failed to load config for Discord adapter", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to load config"})
			return
		}
	}
	// Create a new Discord adapter instance for oauth2/token or ping
	adapter, adapterErr := discord.NewAdapter(cfg.Discord, "oauth2")
	if adapterErr != nil {
		gl.Log("error", "Failed to create Discord adapter", adapterErr)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to create Discord adapter"})
		return
	}

	msg := c.GetString("msg")
	if msg == "" {
		msg = c.Query("msg")
	}
	if msg == "" {
		msg = "Hello from Discord MCP Hub!"
	}

	err = adapter.PingAdapter(msg)
	if err != nil {
		gl.Log("error", "Failed to ping Discord", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "failed to ping Discord"})
		return
	}
	c.JSON(http.StatusOK, DiscordPingResponse{Message: "Discord is reachable"})

}

// GetHubStatus retorna o status atual do hub Discord.
//
// @Summary     Status do hub Discord
// @Description Retorna informações de debug sobre o estado do hub MCP. [Em desenvolvimento]
// @Tags        discord beta
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /api/v1/discord/hub/status [get]
func (dc *DiscordController) GetHubStatus(c *gin.Context) {
	status := map[string]interface{}{
		"hub_initialized": dc.hub != nil,
		"config_loaded":   dc.config != nil,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	if dc.hub != nil {
		eventStream := dc.hub.GetEventStream()
		status["event_stream_available"] = eventStream != nil

		if eventStream != nil {
			// Adicione mais detalhes se disponível na interface
			status["connected_clients"] = "check event stream"
		}
	}

	if dc.config != nil {
		status["discord_config"] = map[string]interface{}{
			"token_set":  dc.config.Discord.Bot.Token != "",
			"app_id_set": dc.config.Discord.Bot.ApplicationID != "",
		}
	}

	c.JSON(http.StatusOK, status)
}
