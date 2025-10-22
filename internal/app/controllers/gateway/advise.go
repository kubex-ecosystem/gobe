// Package gateway implements controllers for gateway-related endpoints.
package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/transport/sse"
	gatewayService "github.com/kubex-ecosystem/gobe/internal/services/gateway"
	gatewaysvc "github.com/kubex-ecosystem/gobe/internal/services/gateway/registry"
	gl "github.com/kubex-ecosystem/logz/logger"
)

type AdviseController struct {
	service *gatewaysvc.Service
}

func NewAdviseController(service *gatewaysvc.Service) *AdviseController {
	if service == nil {
		gl.Log("warn", "advise controller created without gateway service")
	}
	return &AdviseController{service: service}
}

// Advise generates guidance using the configured provider, optionally streaming SSE deltas.
//
// @Summary     Gerar aconselhamento
// @Description Processa o prompt e retorna respostas via JSON ou SSE (`data: {"delta"}`) até concluir. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Produce     text/event-stream
// @Param       X-External-API-Key header string false "Chave externa do cliente"
// @Param       payload body AdviceRequest true "Parâmetros do aconselhamento"
// @Success     200 {object} AdviceResponse "Resposta final consolidada"
// @Failure     400 {object} ErrorResponse "Requisição inválida"
// @Failure     401 {object} ErrorResponse "Não autorizado"
// @Failure     503 {object} ErrorResponse "Gateway indisponível"
// @Router      /api/v1/advise [post]
// @Router      /advise [post]
func (ac *AdviseController) Advise(c *gin.Context) {
	if ac.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "advisor unavailable"})
		return
	}

	var req AdviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	if req.Provider == "" {
		req.Provider = req.Model
	}
	if req.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	if req.Stream || strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
		ac.streamSSE(c, &req)
		return
	}

	ac.respondJSON(c, &req)
}

func (ac *AdviseController) respondJSON(c *gin.Context, req *AdviceRequest) {
	ctx := c.Request.Context()
	svcReq := ac.buildChatRequest(c, req)

	stream, config, err := ac.service.Chat(ctx, svcReq)
	if err != nil {
		gl.Log("error", fmt.Sprintf("advise chat failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var builder strings.Builder
	var usage *gatewayService.Usage

	for chunk := range stream {
		if chunk.Error != "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": chunk.Error})
			return
		}
		if chunk.Content != "" {
			builder.WriteString(chunk.Content)
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
	}

	metadata := map[string]interface{}{
		"provider":  config.Name,
		"model":     config.DefaultModel,
		"timestamp": time.Now().UTC(),
	}
	if usage != nil {
		metadata["usage"] = usage
	}

	c.JSON(http.StatusOK, AdviceResponse{
		Advice:   strings.TrimSpace(builder.String()),
		Metadata: metadata,
	})
}

func (ac *AdviseController) streamSSE(c *gin.Context, req *AdviceRequest) {
	ctx := c.Request.Context()
	svcReq := ac.buildChatRequest(c, req)

	stream, config, err := ac.service.Chat(ctx, svcReq)
	if err != nil {
		gl.Log("error", fmt.Sprintf("advise stream failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, _ := c.Writer.(http.Flusher)
	send := func(payload interface{}) {
		bytes, err := json.Marshal(payload)
		if err != nil {
			gl.Log("error", fmt.Sprintf("advise SSE marshal error: %v", err))
			return
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", bytes); err != nil {
			gl.Log("error", fmt.Sprintf("advise SSE write error: %v", err))
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}

	coalescer := sse.NewCoalescer(0, 0, func(chunk string) {
		send(gin.H{"delta": chunk})
	})

	var usage *gatewayService.Usage

streamLoop:
	for {
		select {
		case <-ctx.Done():
			send(gin.H{"done": true, "cancelled": true})
			return
		case chunk, ok := <-stream:
			if !ok {
				break streamLoop
			}
			if chunk.Error != "" {
				coalescer.Close()
				send(gin.H{"error": chunk.Error, "done": true})
				return
			}
			if chunk.Content != "" {
				if err := coalescer.Add(chunk.Content); err != nil {
					gl.Log("warn", fmt.Sprintf("advise coalescer add error: %v", err))
				}
			}
			if chunk.ToolCall != nil {
				send(gin.H{"tool_call": chunk.ToolCall})
			}
			if chunk.Usage != nil {
				usage = chunk.Usage
			}
			if chunk.Done {
				break streamLoop
			}
		}
	}

	coalescer.Close()

	response := gin.H{
		"done":     true,
		"provider": config.Name,
		"model":    config.DefaultModel,
	}
	if usage != nil {
		response["usage"] = usage
	}

	send(response)
}

func (ac *AdviseController) buildChatRequest(c *gin.Context, req *AdviceRequest) gatewayService.ChatRequest {
	meta := map[string]interface{}{}
	for k, v := range req.Metadata {
		meta[k] = v
	}

	externalKey := strings.TrimSpace(c.GetHeader("x-external-api-key"))

	messages := make([]gatewayService.Message, 0, 4)
	messages = append(messages, gatewayService.Message{
		Role:    "system",
		Content: "You are the Kubex GoBE advisor. Provide concise, actionable insights and highlight next steps.",
	})

	if len(req.Context) > 0 {
		if ctxJSON, err := json.Marshal(req.Context); err == nil {
			messages = append(messages, gatewayService.Message{
				Role:    "system",
				Content: fmt.Sprintf("Context: %s", string(ctxJSON)),
			})
		}
	}

	messages = append(messages, gatewayService.Message{Role: "user", Content: req.Prompt})

	headers := map[string]string{}
	if externalKey != "" {
		meta["external_api_key"] = externalKey
		headers["x-external-api-key"] = externalKey
	}
	if tenantID := strings.TrimSpace(c.GetHeader("x-tenant-id")); tenantID != "" {
		meta["tenant_id"] = tenantID
		headers["x-tenant-id"] = tenantID
	}
	if userID := strings.TrimSpace(c.GetHeader("x-user-id")); userID != "" {
		meta["user_id"] = userID
		headers["x-user-id"] = userID
	}

	return gatewayService.ChatRequest{
		Provider:    req.Provider,
		Model:       req.Model,
		Messages:    messages,
		Temperature: 0.4,
		Stream:      true,
		Meta:        meta,
		Headers:     headers,
	}
}
