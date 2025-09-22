package gateway

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/kubex-ecosystem/gobe/internal/app/transport/sse"
    gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

// ChatController handles /chat SSE traffic.
type ChatController struct{}

func NewChatController() *ChatController {
    return &ChatController{}
}

func (cc *ChatController) ChatSSE(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        gl.Log("warn", fmt.Sprintf("invalid chat payload: %v", err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if req.Temperature == 0 {
        req.Temperature = 0.7
    }
    req.Stream = true
    if req.Meta == nil {
        req.Meta = make(map[string]interface{})
    }

    // Propagate BYOK and tenant headers into metadata for downstream providers.
    req.Meta["external_api_key"] = strings.TrimSpace(c.GetHeader("x-external-api-key"))
    req.Meta["tenant_id"] = strings.TrimSpace(c.GetHeader("x-tenant-id"))
    req.Meta["user_id"] = strings.TrimSpace(c.GetHeader("x-user-id"))

    // Prepare SSE response.
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")
    c.Status(http.StatusOK)

    ctx := c.Request.Context()
    usage := struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    }{
        PromptTokens: len(req.Messages),
    }

    sendEvent := func(payload interface{}) {
        data, err := json.Marshal(payload)
        if err != nil {
            gl.Log("error", fmt.Sprintf("failed to marshal SSE payload: %v", err))
            return
        }
        if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
            gl.Log("error", fmt.Sprintf("failed to write SSE payload: %v", err))
            return
        }
        if flusher, ok := c.Writer.(http.Flusher); ok {
            flusher.Flush()
        }
    }

    coalescer := sse.NewCoalescer(0, 0, func(chunk string) {
        sendEvent(gin.H{"delta": chunk})
        usage.CompletionTokens += len(chunk)
        usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
    })

    stream := cc.simulateStream(&req)
    for _, chunk := range stream {
        if err := coalescer.Add(chunk); err != nil {
            gl.Log("warn", fmt.Sprintf("coalescer add failed: %v", err))
            break
        }
        select {
        case <-ctx.Done():
            gl.Log("warn", "chat stream cancelled by client")
            coalescer.Flush()
            coalescer.Close()
            sendEvent(gin.H{"done": true, "cancelled": true})
            return
        case <-time.After(25 * time.Millisecond):
        }
    }

    coalescer.Close()
    sendEvent(gin.H{"done": true, "usage": usage})
}

func (cc *ChatController) simulateStream(req *ChatRequest) []string {
    if len(req.Messages) == 0 {
        return []string{"Hello!", " I'm the GoBE gateway placeholder response."}
    }

    last := req.Messages[len(req.Messages)-1]
    reply := fmt.Sprintf("Echoing (%s/%s): %s", req.Provider, req.Model, last.Content)
    if strings.TrimSpace(reply) == "" {
        reply = "Hi there!"
    }

    // Basic chunking to emulate tokenized output.
    words := strings.Fields(reply)
    if len(words) == 0 {
        return []string{"(empty response)"}
    }

    chunks := make([]string, 0, len(words))
    current := words[0]
    for i := 1; i < len(words); i++ {
        next := words[i]
        if len(current)+len(next)+1 > 20 {
            chunks = append(chunks, current+" ")
            current = next
        } else {
            current += " " + next
        }
    }
    chunks = append(chunks, current+".")

    return chunks
}

