// Package providers provides a factory for creating AI model provider instances.
package providers

import (
	"fmt"

	gateway "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type Config = gateway.ProviderConfig

func New(cfg Config) (gateway.Provider, error) {
	switch cfg.Type {
	case "openai":
		return newOpenAIProvider(cfg)
	case "groq":
		return newGroqProvider(cfg)
	case "anthropic":
		return newAnthropicProvider(cfg)
	case "gemini":
		return newGeminiProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}
