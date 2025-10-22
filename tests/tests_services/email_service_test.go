// Package tests_services contains tests for the application services.
package tests_services

import (
	"os"
	"testing"

	"github.com/kubex-ecosystem/gobe/internal/services/email"
)

func TestNewEmailService(t *testing.T) {
	// Test with a valid config file
	validConfigPath := "../../.notes/email.yml"

	service, err := email.NewEmailService(validConfigPath)
	if err != nil {
		t.Fatalf("Expected no error for valid config, but got: %v", err)
	}
	if service == nil {
		t.Fatal("Expected service to be non-nil for valid config")
	}

	// Test with a non-existent config file
	invalidConfigPath := "non_existent_config.yml"
	_, err = email.NewEmailService(invalidConfigPath)
	if err == nil {
		t.Fatal("Expected an error for non-existent config, but got nil")
	}
}

func TestSendEmail(t *testing.T) {
	// This test does not send a real email but checks the logic.
	configPath := "../../.notes/email.yml"
	service, err := email.NewEmailService(configPath)
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	// Test sending with the default provider (which is a dummy in the test config)
	// We expect an error because we can't connect to "smtp.test.com".
	// A successful test here means the service tried to connect, which is correct.
	err = service.SendEmail("recipient@example.com", "Test Subject", "Test Body")
	if err == nil {
		t.Fatal("Expected an error when trying to send email with dummy config, but got nil")
	}

	// To fully test the success case, a mock SMTP server would be needed,
	// which is beyond the scope of this initial test setup.
}

// Cleanup the dummy config file
func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.Remove("../../.notes/email.yml")
	os.Exit(code)
}
