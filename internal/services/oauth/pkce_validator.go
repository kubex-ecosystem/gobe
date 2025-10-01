// Package oauth provides OAuth2 and PKCE validation services
package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCEValidator handles PKCE (Proof Key for Code Exchange) validation
type PKCEValidator struct{}

// NewPKCEValidator creates a new PKCE validator
func NewPKCEValidator() *PKCEValidator {
	return &PKCEValidator{}
}

// ValidateCodeVerifier validates a code_verifier against a code_challenge
// Supports both S256 and plain methods
func (v *PKCEValidator) ValidateCodeVerifier(codeVerifier, codeChallenge, method string) error {
	if codeVerifier == "" {
		return fmt.Errorf("code_verifier is required")
	}
	if codeChallenge == "" {
		return fmt.Errorf("code_challenge is required")
	}

	// Validate code_verifier length (43-128 characters, RFC 7636)
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return fmt.Errorf("code_verifier must be between 43 and 128 characters")
	}

	var computedChallenge string

	switch method {
	case "S256":
		// SHA256(code_verifier) -> base64url encoding
		hash := sha256.Sum256([]byte(codeVerifier))
		computedChallenge = base64.RawURLEncoding.EncodeToString(hash[:])

	case "plain":
		// Plain method: code_challenge == code_verifier
		computedChallenge = codeVerifier

	default:
		return fmt.Errorf("unsupported code_challenge_method: %s", method)
	}

	if computedChallenge != codeChallenge {
		return fmt.Errorf("code_verifier does not match code_challenge")
	}

	return nil
}

// GenerateCodeChallenge generates a code_challenge from a code_verifier using S256
// This is useful for testing or client-side generation
func (v *PKCEValidator) GenerateCodeChallenge(codeVerifier string) (string, error) {
	if codeVerifier == "" {
		return "", fmt.Errorf("code_verifier is required")
	}

	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return "", fmt.Errorf("code_verifier must be between 43 and 128 characters")
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge, nil
}
