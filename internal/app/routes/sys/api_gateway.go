package sys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	apia "github.com/rafa-mori/gobe/internal/app/security/authentication"
	"github.com/rafa-mori/gobe/internal/module/logger"
)

type APIGateway struct {
	AuthManager    *apia.AuthManager
	WebhookManager *WebhookManager
}

func NewAPIGateway(authManager *apia.AuthManager) *APIGateway {
	return &APIGateway{
		AuthManager:    authManager,
		WebhookManager: NewWebhookManager(),
	}
}

func (gateway *APIGateway) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	api.POST("/api/v1/smart-plane/register", gateway.HandleRegisterDocument)
	api.POST("/api/v1/smart-plane/approve", gateway.HandleApproveDocument)
	api.POST("/api/v1/smart-plane/sign", gateway.HandleSignDocument)
	api.GET("/api/v1/smart-plane/history", gateway.HandleGetDocumentHistory)
	api.DELETE("/api/v1/smart-plane/delete", gateway.HandleDeleteDocumentState)
}

type WebhookManager struct {
	webhooks map[string]string
}

func NewWebhookManager() *WebhookManager {
	return &WebhookManager{
		webhooks: make(map[string]string),
	}
}

func (wm *WebhookManager) RegisterWebhook(event string, url string) {
	wm.webhooks[event] = url
}

func (wm *WebhookManager) TriggerWebhook(event string, payload interface{}) error {
	url, exists := wm.webhooks[event]
	if !exists {
		return fmt.Errorf("webhook for event %s not found", event)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to trigger webhook: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook responded with status: %d", resp.StatusCode)
	}

	return nil
}

func (gateway *APIGateway) HandleRegisterDocument(c *gin.Context) {
	logger.Log("info", "HandleRegisterDocument called")
	// Simulate document registration logic
	payload := map[string]string{"message": "Document registered successfully"}
	if err := gateway.WebhookManager.TriggerWebhook("register", payload); err != nil {
		logger.Log("error", fmt.Sprintf("Failed to trigger webhook: %v", err))
	}
	c.JSON(http.StatusOK, payload)
}

func (gateway *APIGateway) HandleApproveDocument(c *gin.Context) {
	logger.Log("info", "HandleApproveDocument called")
	// Simulate document approval logic
	payload := map[string]string{"message": "Document approved successfully"}
	if err := gateway.WebhookManager.TriggerWebhook("approve", payload); err != nil {
		logger.Log("error", fmt.Sprintf("Failed to trigger webhook: %v", err))
	}
	c.JSON(http.StatusOK, payload)
}

func (gateway *APIGateway) HandleSignDocument(c *gin.Context) {
	// Implementação da lógica para assinar documentos
	logger.Log("info", "HandleSignDocument called")
	c.JSON(http.StatusOK, gin.H{"message": "Document signed successfully"})
}

func (gateway *APIGateway) HandleGetDocumentHistory(c *gin.Context) {
	// Implementação da lógica para obter histórico de documentos
	logger.Log("info", "HandleGetDocumentHistory called")
	c.JSON(http.StatusOK, gin.H{"message": "Document history retrieved successfully"})
}

func (gateway *APIGateway) HandleDeleteDocumentState(c *gin.Context) {
	// Implementação da lógica para deletar estado de documentos
	logger.Log("info", "HandleDeleteDocumentState called")
	c.JSON(http.StatusOK, gin.H{"message": "Document state deleted successfully"})
}
