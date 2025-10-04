package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
)

// Service provides methods to interact with the WhatsApp Business API.
type Service struct {
	cfg    bootstrap.WhatsAppConfig
	client *http.Client
}

// NewService creates a new WhatsApp service with the provided configuration.
func NewService(cfg bootstrap.WhatsAppConfig) *Service {
	return &Service{cfg: cfg, client: &http.Client{}}
}

// Config returns the underlying WhatsApp configuration.
func (s *Service) Config() bootstrap.WhatsAppConfig { return s.cfg }

// OutgoingMessage represents a message to be sent via WhatsApp.
type OutgoingMessage struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

// SendMessage sends a text message using the WhatsApp Business API.
func (s *Service) SendMessage(msg OutgoingMessage) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("whatsapp integration disabled")
	}

	body := map[string]any{
		"messaging_product": "whatsapp",
		"to":                msg.To,
		"type":              "text",
		"text": map[string]string{
			"body": msg.Text,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/messages", s.cfg.PhoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("whatsapp send failed: %s", resp.Status)
	}
	return nil
}
