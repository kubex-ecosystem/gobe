// Package email provides a service for sending emails.
package email

import (
	"fmt"
	"net/smtp"
	"os"

	"gopkg.in/yaml.v3"
)

// EmailConfig defines the structure for the email configuration.
type EmailConfig struct {
	DefaultProvider string `yaml:"default_provider"`
	Providers       map[string]struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"providers"`
}

// EmailService implements the IEmailService interface.
type EmailService struct {
	config *EmailConfig
}

// NewEmailService creates a new instance of the email service.
func NewEmailService(configPath string) (*EmailService, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read email config file: %w", err)
	}

	var config EmailConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal email config: %w", err)
	}

	return &EmailService{config: &config}, nil
}

// SendEmail sends an email using the configured provider.
func (s *EmailService) SendEmail(to, subject, body string) error {
	providerName := s.config.DefaultProvider
	provider, ok := s.config.Providers[providerName]
	if !ok {
		return fmt.Errorf("email provider '%s' not configured", providerName)
	}

	auth := smtp.PlainAuth("", provider.Username, provider.Password, provider.Host)

	addr := fmt.Sprintf("%s:%d", provider.Host, provider.Port)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
		body)

	return smtp.SendMail(addr, auth, provider.Username, []string{to}, msg)
}
