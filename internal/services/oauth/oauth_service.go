package oauth

import (
	"context"
	"fmt"

	"github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// IOAuthService defines the OAuth2/PKCE service interface
type IOAuthService interface {
	// Authorization flow
	GenerateAuthorizationCode(ctx context.Context, userID, clientID, redirectURI, codeChallenge, method, scope string) (string, error)

	// Token exchange
	ExchangeCodeForTokens(ctx context.Context, code, codeVerifier, clientID string) (*interfaces.TokenPair, error)

	// Client validation
	ValidateClient(clientID, redirectURI string) error
}

// OAuthService implements IOAuthService
type OAuthService struct {
	oauthClientService gdbasez.OAuthClientService
	authCodeService    gdbasez.AuthCodeService
	userService        gdbasez.UserService
	tokenService       interfaces.TokenService
	pkceValidator      *PKCEValidator
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	oauthClientService gdbasez.OAuthClientService,
	authCodeService gdbasez.AuthCodeService,
	userService gdbasez.UserService,
	tokenService interfaces.TokenService,
) IOAuthService {
	return &OAuthService{
		oauthClientService: oauthClientService,
		authCodeService:    authCodeService,
		userService:        userService,
		tokenService:       tokenService,
		pkceValidator:      NewPKCEValidator(),
	}
}

// GenerateAuthorizationCode creates a new authorization code after validating the client
func (s *OAuthService) GenerateAuthorizationCode(
	ctx context.Context,
	userID, clientID, redirectURI, codeChallenge, method, scope string,
) (string, error) {
	// Validate client exists and is active
	if err := s.ValidateClient(clientID, redirectURI); err != nil {
		return "", fmt.Errorf("client validation failed: %w", err)
	}

	// Generate authorization code (10 minutes expiration)
	authCode, err := s.authCodeService.GenerateCode(
		userID,
		clientID,
		redirectURI,
		codeChallenge,
		method,
		scope,
		10, // 10 minutes expiration
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuth: generated authorization code for user %s, client %s", userID, clientID))
	return authCode.GetCode(), nil
}

// ExchangeCodeForTokens exchanges an authorization code for access and refresh tokens
func (s *OAuthService) ExchangeCodeForTokens(
	ctx context.Context,
	code, codeVerifier, clientID string,
) (*interfaces.TokenPair, error) {
	// Validate and consume the authorization code
	authCode, err := s.authCodeService.ValidateAndConsume(code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}

	// Verify client_id matches
	if authCode.GetClientID() != clientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	// Validate PKCE code_verifier
	if err := s.pkceValidator.ValidateCodeVerifier(
		codeVerifier,
		authCode.GetCodeChallenge(),
		authCode.GetCodeChallengeMethod(),
	); err != nil {
		return nil, fmt.Errorf("PKCE validation failed: %w", err)
	}

	// Get user from injected service
	user, err := s.userService.GetUserByID(authCode.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate token pair using existing TokenService
	tokenPair, err := s.tokenService.NewPairFromUser(ctx, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	gl.Log("info", fmt.Sprintf("OAuth: exchanged code for tokens for user %s", authCode.GetUserID()))
	return tokenPair, nil
}

// ValidateClient validates that a client exists, is active, and the redirect_uri is allowed
func (s *OAuthService) ValidateClient(clientID, redirectURI string) error {
	if clientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if redirectURI == "" {
		return fmt.Errorf("redirect_uri is required")
	}

	// Use the service method that checks both existence and active status
	if err := s.oauthClientService.ValidateRedirectURI(clientID, redirectURI); err != nil {
		return fmt.Errorf("client validation failed: %w", err)
	}

	return nil
}
