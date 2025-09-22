package gateway

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/services/gateway/registry"
)

type Service struct {
	registry *registry.Registry
}

func NewService(providerSvc svc.ProvidersService) (*Service, error) {
	reg := registry.New(providerSvc)
	if err := reg.Reload(); err != nil {
		return nil, err
	}
	return &Service{registry: reg}, nil
}

func (s *Service) Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, ProviderConfig, error) {
	if req.Provider == "" {
		return nil, ProviderConfig{}, fmt.Errorf("gateway: chat request missing provider")
	}

	entry, err := s.registry.Resolve(req.Provider)
	if err != nil {
		return nil, ProviderConfig{}, err
	}

	config := entry.Config
	if req.Model == "" {
		req.Model = config.DefaultModel
	}

	stream, err := entry.Provider.Chat(ctx, req)
	if err != nil {
		return nil, ProviderConfig{}, err
	}

	return stream, config, nil
}

func (s *Service) ProviderSummaries() []ProviderSummary {
	return s.registry.Summaries()
}

func (s *Service) ProviderConfig(name string) (ProviderConfig, error) {
	return s.registry.ConfigFor(name)
}

func (s *Service) Reload() error {
	return s.registry.Reload()
}

