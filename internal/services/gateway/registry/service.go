package registry

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	t "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type Service struct {
	registry *Registry
}

func NewService(providerSvc svc.ProvidersService) (*Service, error) {
	reg := New(providerSvc)
	// Try to reload, but don't fail if table doesn't exist yet (during initialization)
	if err := reg.Reload(); err != nil {
		// Log warning but continue - migrations may not have run yet
		// Service will work with env-based providers only until Reload succeeds
		fmt.Printf("Warning: failed to load providers from database (may not be initialized yet): %v\n", err)
	}
	return &Service{registry: reg}, nil
}

func (s *Service) Chat(ctx context.Context, req t.ChatRequest) (<-chan t.ChatChunk, t.ProviderConfig, error) {
	if req.Provider == "" {
		return nil, t.ProviderConfig{}, fmt.Errorf("gateway: chat request missing provider")
	}

	entry, err := s.registry.Resolve(req.Provider)
	if err != nil {
		return nil, t.ProviderConfig{}, err
	}

	config := entry.Config
	if req.Model == "" {
		req.Model = config.DefaultModel
	}

	stream, err := entry.Provider.Chat(ctx, req)
	if err != nil {
		return nil, t.ProviderConfig{}, err
	}

	return stream, config, nil
}

func (s *Service) ProviderSummaries() []t.ProviderSummary {
	return s.registry.Summaries()
}

func (s *Service) ProviderConfig(name string) (t.ProviderConfig, error) {
	return s.registry.ConfigFor(name)
}

func (s *Service) Reload() error {
	return s.registry.Reload()
}
