// Package registration provides a service for user registration.
package registration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
)

// RegistrationService orchestrates the user registration process.
type RegistrationService struct {
	GDBaseBridge    *gdbasez.Bridge
	EmailService    interfaces.IEmailService
	VerificationURL string // Base URL for email verification, e.g., "http://localhost:8080/api/v1/verify-email?token="
}

// NewRegistrationService creates a new instance of the registration service.
func NewRegistrationService(bridge *gdbasez.Bridge, emailService interfaces.IEmailService, verificationURL string) *RegistrationService {
	return &RegistrationService{
		GDBaseBridge:    bridge,
		EmailService:    emailService,
		VerificationURL: verificationURL,
	}
}

// InitiateRegistration starts the registration process for a new user.
func (s *RegistrationService) InitiateRegistration(ctx context.Context, name, email, password string) error {
	// TODO: Get user service from bridge
	// TODO: Check if user already exists

	// Create a new user (inactive)
	newUser := &gdbasez.UserModelImpl{}
	newUser.SetName(name)
	newUser.SetEmail(email)
	newUser.SetPassword(password)
	newUser.SetActive(false)

	// TODO: Save user via user service

	// Generate verification token
	token, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Create and save registration token
	expiresAt := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	regToken := s.GDBaseBridge.RegistrationTokenModel(newUser.GetID(), token, expiresAt)

	// TODO: Get token service from bridge and save token

	// Send verification email
	err = s.sendVerificationEmail(newUser.GetEmail(), regToken.Token)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// CompleteRegistration completes the registration process.
func (s *RegistrationService) CompleteRegistration(ctx context.Context, token string) error {
	// TODO: Get token service from bridge
	// TODO: Find token
	// TODO: Check if token is valid and not expired
	// TODO: Get user service from bridge
	// TODO: Find user by ID from token
	// TODO: Activate user
	// TODO: Delete token

	return nil
}

func (s *RegistrationService) sendVerificationEmail(email, token string) error {
	verificationURL := s.VerificationURL + token

	tmpl, err := template.ParseFiles("internal/services/email/templates/registration.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	data := struct{ VerificationURL string }{VerificationURL: verificationURL}
	err = tmpl.Execute(&body, data)
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return s.EmailService.SendEmail(email, "Complete Your Registration", body.String())
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
