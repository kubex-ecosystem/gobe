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

// AdviseController produces lightweight AI advise responses (placeholder implementation).
type AdviseController struct{}

func NewAdviseController() *AdviseController {
    return &AdviseController{}
}

func (ac *AdviseController) Advise(c *gin.Context) {
    var req AdviceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    req.Stream = req.Stream || strings.Contains(c.GetHeader("Accept"), "text/event-stream")
    if !req.Stream {
        ac.respondJSON(c, &req)
        return
    }

    ac.streamSSE(c, &req)
}

func (ac *AdviseController) respondJSON(c *gin.Context, req *AdviceRequest) {
    advice := AdviceResponse{
        Advice: fmt.Sprintf("Placeholder advice for '%s'", strings.TrimSpace(req.Prompt)),
        Metadata: map[string]interface{}{
            "provider": req.Provider,
            "model":    req.Model,
            "timestamp": time.Now().UTC(),
            "note":      "TODO: replace with real advise engine",
        },
    }
    c.JSON(http.StatusOK, advice)
}

func (ac *AdviseController) streamSSE(c *gin.Context, req *AdviceRequest) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")
    c.Status(http.StatusOK)

    payload := fmt.Sprintf("Advising on: %s", strings.TrimSpace(req.Prompt))
    fragments := []string{"Analyzing context... ", payload, " Done."}

    send := func(data interface{}) {
        b, err := json.Marshal(data)
        if err != nil {
            gl.Log("error", fmt.Sprintf("advise SSE marshal error: %v", err))
            return
        }
        if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", b); err != nil {
            gl.Log("error", fmt.Sprintf("advise SSE write error: %v", err))
            return
        }
        if flusher, ok := c.Writer.(http.Flusher); ok {
            flusher.Flush()
        }
    }

    coalescer := sse.NewCoalescer(0, 0, func(chunk string) {
        send(gin.H{"delta": chunk})
    })

    for _, frag := range fragments {
        if err := coalescer.Add(frag); err != nil {
            gl.Log("warn", fmt.Sprintf("advise coalescer add error: %v", err))
            break
        }
        select {
        case <-c.Request.Context().Done():
            coalescer.Flush()
            send(gin.H{"done": true, "cancelled": true})
            return
        case <-time.After(50 * time.Millisecond):
        }
    }

    coalescer.Close()
    send(gin.H{
        "done": true,
        "metadata": map[string]interface{}{
            "provider": req.Provider,
            "model":    req.Model,
        },
    })
}

