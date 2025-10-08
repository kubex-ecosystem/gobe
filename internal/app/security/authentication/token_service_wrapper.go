package authentication

import (
	"context"

	mdl "github.com/kubex-ecosystem/gdbase/factory/models"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

// TokenServiceWrapper wraps JWTService to implement the TokenService interface
type TokenServiceWrapper struct {
	jwtService *JWTService
}

// NewPairFromUser generates a new token pair for the given user
func (w *TokenServiceWrapper) NewPairFromUser(ctx context.Context, u mdl.UserModel, prevTokenID string) (*sci.TokenPair, error) {
	return w.jwtService.NewPairFromUser(ctx, u, prevTokenID)
}

// SignOut signs out a user by revoking all their tokens
func (w *TokenServiceWrapper) SignOut(ctx context.Context, uid string) error {
	return w.jwtService.SignOut(ctx, uid)
}

// ValidateIDToken validates an ID token and returns the user
func (w *TokenServiceWrapper) ValidateIDToken(tokenString string) (mdl.UserModel, error) {
	return w.jwtService.ValidateIDToken(tokenString)
}

// ValidateRefreshToken validates a refresh token
func (w *TokenServiceWrapper) ValidateRefreshToken(refreshTokenString string) (*sci.RefreshToken, error) {
	return w.jwtService.ValidateRefreshToken(refreshTokenString)
}

// RenewToken renews a token pair using a valid refresh token
func (w *TokenServiceWrapper) RenewToken(ctx context.Context, refreshToken string) (*sci.TokenPair, error) {
	return w.jwtService.RenewToken(ctx, refreshToken)
}
