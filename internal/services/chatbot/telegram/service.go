package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
)

// Service interacts with Telegram Bot API.
type Service struct {
	cfg    bootstrap.TelegramConfig
	client *http.Client
}

// NewService creates a new Telegram service.
func NewService(cfg bootstrap.TelegramConfig) *Service {
	return &Service{cfg: cfg, client: &http.Client{}}
}

// Config returns current configuration.
func (s *Service) Config() bootstrap.TelegramConfig { return s.cfg }

// OutgoingMessage represents a telegram message to send.
type OutgoingMessage struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// SendMessage sends a message using Telegram Bot API.
func (s *Service) SendMessage(msg OutgoingMessage) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("telegram integration disabled")
	}
	body := map[string]any{
		"chat_id": msg.ChatID,
		"text":    msg.Text,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.cfg.BotToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram send failed: %s", resp.Status)
	}
	return nil
}
