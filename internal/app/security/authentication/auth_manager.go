// Package authentication provides methods for generating and validating JWT tokens.
package authentication

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/google/uuid"
	crt "github.com/rafa-mori/gobe/internal/app/security/certificates"
	"github.com/rafa-mori/gobe/internal/module/logger"
)

type AuthManager struct {
	privKey               *rsa.PrivateKey
	pubKey                *rsa.PublicKey
	refreshSecret         string
	idExpirationSecs      int64
	refreshExpirationSecs int64
}

func NewAuthManager(certService crt.CertService) (*AuthManager, error) {
	privKey, err := certService.GetPrivateKey()
	if err != nil {
		logger.Log("error", fmt.Sprintf("Failed to load private key: %v", err))
		return nil, err
	}

	pubKey, err := certService.GetPublicKey()
	if err != nil {
		logger.Log("error", fmt.Sprintf("Failed to load public key: %v", err))
		return nil, err
	}

	return &AuthManager{
		privKey:               privKey,
		pubKey:                pubKey,
		refreshSecret:         "default_refresh_secret", // Replace with a secure secret
		idExpirationSecs:      3600,                     // 1 hour
		refreshExpirationSecs: 604800,                   // 7 days
	}, nil
}

func (am *AuthManager) GenerateIDToken(userID string) (string, error) {

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Duration(am.idExpirationSecs) * time.Second)},
		IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		ID:        uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(am.privKey)
}

func (am *AuthManager) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Duration(am.refreshExpirationSecs) * time.Second)},
		IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		ID:        uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.refreshSecret))
}

func (am *AuthManager) ValidateIDToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return am.pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (am *AuthManager) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(am.refreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
