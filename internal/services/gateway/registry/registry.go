// Package registry provides a registry for managing AI model provider configurations.
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	gateway "github.com/kubex-ecosystem/gobe/internal/services/gateway"
	"github.com/kubex-ecosystem/gobe/internal/services/gateway/providers"
)

type Registry struct {
	mu          sync.RWMutex
	entries     map[string]*gateway.ProviderEntry
	providerSvc svc.ProvidersService
}

var ErrProviderNotFound = errors.New("gateway: provider not found")

type providerPayload struct {
	Type         string                 `json:"type"`
	BaseURL      string                 `json:"base_url"`
	DefaultModel string                 `json:"default_model"`
	APIKey       string                 `json:"api_key"`
	KeyEnv       string                 `json:"key_env"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func New(providerSvc svc.ProvidersService) *Registry {
	return &Registry{
		entries:     make(map[string]*gateway.ProviderEntry),
		providerSvc: providerSvc,
	}
}

func (r *Registry) Reload() error {
	if r.providerSvc == nil {
		return errors.New("gateway: provider service is nil")
	}

	records, err := r.providerSvc.ListProviders()
	if err != nil {
		return fmt.Errorf("gateway: failed to list providers: %w", err)
	}

	entries := make(map[string]*gateway.ProviderEntry)

	for _, record := range records {
		configMap := record.GetConfig()

		payload := providerPayload{Metadata: map[string]interface{}{}}
		if len(configMap) > 0 {
			if raw, err := json.Marshal(configMap); err == nil {
				if err := json.Unmarshal(raw, &payload); err != nil {
					gl.Log("warn", "gateway registry unable to parse provider config", record.GetProvider(), err)
				}
			}
		}

		cfg := gateway.ProviderConfig{
			Name:         record.GetProvider(),
			Type:         payload.Type,
			BaseURL:      payload.BaseURL,
			DefaultModel: payload.DefaultModel,
			APIKey:       payload.APIKey,
			KeyEnv:       payload.KeyEnv,
			Org:          record.GetOrgOrGroup(),
			Metadata:     payload.Metadata,
		}

		if cfg.Type == "" {
			gl.Log("warn", "gateway registry skipping provider without type", cfg.Name)
			continue
		}

		prov, err := providers.New(cfg)
		if err != nil {
			gl.Log("error", "gateway registry failed to instantiate provider", cfg.Name, err)
			continue
		}

		entries[cfg.Name] = &gateway.ProviderEntry{Config: cfg, Provider: prov}
	}

	r.mu.Lock()
	r.entries = entries
	r.mu.Unlock()

	return nil
}

func (r *Registry) Resolve(name string) (*gateway.ProviderEntry, error) {
	r.mu.RLock()
	entry, ok := r.entries[name]
	r.mu.RUnlock()
	if ok {
		return entry, nil
	}

	if err := r.Reload(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	entry, ok = r.entries[name]
	r.mu.RUnlock()
	if !ok {
		return nil, ErrProviderNotFound
	}

	return entry, nil
}

func (r *Registry) Summaries() []gateway.ProviderSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summaries := make([]gateway.ProviderSummary, 0, len(r.entries))
	for name, entry := range r.entries {
		summary := gateway.ProviderSummary{
			Name:         name,
			Type:         entry.Config.Type,
			Org:          entry.Config.Org,
			DefaultModel: entry.Config.DefaultModel,
			Metadata:     entry.Config.Metadata,
		}

		if entry.Provider != nil {
			if err := entry.Provider.Available(); err != nil {
				summary.Available = false
				summary.LastError = err.Error()
			} else {
				summary.Available = true
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

func (r *Registry) ConfigFor(name string) (gateway.ProviderConfig, error) {
	r.mu.RLock()
	entry, ok := r.entries[name]
	r.mu.RUnlock()
	if ok {
		return entry.Config, nil
	}

	if err := r.Reload(); err != nil {
		return gateway.ProviderConfig{}, err
	}

	r.mu.RLock()
	entry, ok = r.entries[name]
	r.mu.RUnlock()
	if !ok {
		return gateway.ProviderConfig{}, ErrProviderNotFound
	}

	return entry.Config, nil
}
