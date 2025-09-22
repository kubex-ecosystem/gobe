package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/transport/sse"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	gatewaysvc "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type ChatController struct {
	service *gatewaysvc.Service
}

func NewChatController(service *gatewaysvc.Service) *ChatController {
	if service == nil {
		gl.Log("warn", "chat controller created without gateway service")
	}
	return &ChatController{service: service}
}

// ChatSSE streams provider responses as Server-Sent Events for conversational use cases.
//
// @Summary     Chat streaming
// @Description Dispara uma conversação streaming (`data: {"delta"}`) com o provedor configurado.
// @Tags        gateway
// @Security    BearerAuth
// @Accept      json
// @Produce     text/event-stream
// @Param       X-External-API-Key header string false "Chave externa do cliente"
// @Param       X-Tenant-ID       header string false "Tenant que originou a requisição"
// @Param       X-User-ID         header string false "Identificador do usuário"
// @Param       payload body ChatRequest true "Dados da sessão de chat"
// @Success     200 {string} string "Fluxo SSE com eventos {\"delta\":string}"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     503 {object} ErrorResponse
// @Router      /chat [post]
func (cc *ChatController) ChatSSE(c *gin.Context) {
	if cc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gateway service unavailable"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gl.Log("warn", fmt.Sprintf("invalid chat payload: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages cannot be empty"})
		return
	}

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.Meta == nil {
		req.Meta = make(map[string]interface{})
	}

	externalKey := strings.TrimSpace(c.GetHeader("x-external-api-key"))
	tenantID := strings.TrimSpace(c.GetHeader("x-tenant-id"))
	userID := strings.TrimSpace(c.GetHeader("x-user-id"))

	if externalKey != "" {
		req.Meta["external_api_key"] = externalKey
	}
	if tenantID != "" {
		req.Meta["tenant_id"] = tenantID
	}
	if userID != "" {
		req.Meta["user_id"] = userID
	}

	svcReq := gatewaysvc.ChatRequest{
		Provider:    req.Provider,
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      true,
		Meta:        req.Meta,
		Headers: map[string]string{
			"x-external-api-key": externalKey,
			"x-tenant-id":        tenantID,
			"x-user-id":          userID,
		},
	}

	ctx := c.Request.Context()
	stream, config, err := cc.service.Chat(ctx, svcReq)
	if err != nil {
		gl.Log("error", fmt.Sprintf("chat service failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, _ := c.Writer.(http.Flusher)
	sendEvent := func(payload interface{}) {
		bytes, err := json.Marshal(payload)
		if err != nil {
			gl.Log("error", fmt.Sprintf("failed to marshal SSE payload: %v", err))
			return
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", bytes); err != nil {
			gl.Log("error", fmt.Sprintf("failed to write SSE payload: %v", err))
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}

	coalescer := sse.NewCoalescer(0, 0, func(chunk string) {
		sendEvent(gin.H{"delta": chunk})
	})

	var lastUsage *gatewaysvc.Usage

streamLoop:
	for {
		select {
		case <-ctx.Done():
			gl.Log("warn", "chat stream cancelled by client")
			sendEvent(gin.H{"done": true, "cancelled": true})
			return
		case chunk, ok := <-stream:
			if !ok {
				break streamLoop
			}

			if chunk.Error != "" {
				coalescer.Close()
				sendEvent(gin.H{"error": chunk.Error, "done": true})
				return
			}

			if chunk.Content != "" {
				if err := coalescer.Add(chunk.Content); err != nil {
					gl.Log("warn", fmt.Sprintf("chat coalescer failed: %v", err))
				}
			}

			if chunk.ToolCall != nil {
				sendEvent(gin.H{"tool_call": chunk.ToolCall})
			}

			if chunk.Usage != nil {
				lastUsage = chunk.Usage
			}

			if chunk.Done {
				break streamLoop
			}
		}
	}

	coalescer.Close()

	response := gin.H{"done": true, "provider": config.Name}
	if lastUsage != nil {
		response["usage"] = lastUsage
		if lastUsage.Model != "" {
			response["model"] = lastUsage.Model
		}
	}
	if _, ok := response["model"]; !ok && config.DefaultModel != "" {
		response["model"] = config.DefaultModel
	}

	sendEvent(response)
}
