package providers

import (
	"os"
	"strings"

	gateway "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

func staticAPIKey(cfg Config) string {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" && cfg.KeyEnv != "" {
		key = strings.TrimSpace(os.Getenv(cfg.KeyEnv))
	}
	return key
}

func externalAPIKey(req gateway.ChatRequest) string {
	if req.Meta != nil {
		if raw, ok := req.Meta["external_api_key"].(string); ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw)
		}
	}
	if req.Headers != nil {
		if raw := strings.TrimSpace(req.Headers["x-external-api-key"]); raw != "" {
			return raw
		}
	}
	return ""
}

