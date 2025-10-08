package authentication

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	mdl "github.com/kubex-ecosystem/gdbase/factory/models"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// JWTService handles JWT token generation and validation using GDBase repositories
type JWTService struct {
	tokenRepo             mdl.ITokenRepo
	privKey               *rsa.PrivateKey
	pubKey                *rsa.PublicKey
	refreshSecret         string
	idExpirationSecs      int64
	refreshExpirationSecs int64
}

// jwtIDTokenClaims represents the claims embedded in ID tokens
type jwtIDTokenClaims struct {
	User mdl.UserModel `json:"UserImpl"`
	jwt.RegisteredClaims
}

// jwtRefreshTokenClaims represents the claims embedded in refresh tokens
type jwtRefreshTokenClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// jwtRefreshTokenData holds refresh token metadata
type jwtRefreshTokenData struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}

// NewJWTService creates a new JWT service instance using GDBase token repository
func NewJWTService(tokenRepo mdl.ITokenRepo, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, refreshSecret string, idExpSecs, refreshExpSecs int64) *JWTService {
	if tokenRepo == nil {
		gl.Log("error", "TokenRepo cannot be nil") // pragma: allowlist secret // pragma: allowlist secret
		return nil
	}
	if privKey == nil { // pragma: allowlist secret
		gl.Log("error", "Private key cannot be nil") // pragma: allowlist secret
		return nil
	}
	if pubKey == nil { // pragma: allowlist secret
		gl.Log("error", "Public key cannot be nil") // pragma: allowlist secret
		return nil
	}

	// Set default expiration times
	if idExpSecs == 0 {
		idExpSecs = 3600 // 1 hour
	}
	if refreshExpSecs == 0 {
		refreshExpSecs = 604800 // 7 days
	}

	return &JWTService{
		tokenRepo:             tokenRepo,
		privKey:               privKey,
		pubKey:                pubKey,
		refreshSecret:         refreshSecret,
		idExpirationSecs:      idExpSecs,
		refreshExpirationSecs: refreshExpSecs,
	}
}

// NewPairFromUser generates a new token pair (ID token + Refresh token) for a user
func (s *JWTService) NewPairFromUser(ctx context.Context, u mdl.UserModel, prevTokenID string) (*sci.TokenPair, error) {
	// Delete previous refresh token if provided
	if prevTokenID != "" {
		if err := s.tokenRepo.DeleteRefreshToken(ctx, u.GetID(), prevTokenID); err != nil {
			gl.Log("error", fmt.Sprintf("could not delete previous refresh token for uid: %v, tokenID: %v: %v", u.GetID(), prevTokenID, err))
			return nil, fmt.Errorf("could not delete previous refresh token: %w", err)
		}
	}

	// Generate ID token
	idToken, err := s.generateIDToken(u)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error generating id token for uid: %v: %v", u.GetID(), err))
		return nil, fmt.Errorf("error generating id token: %w", err)
	}

	// Ensure refresh secret is set
	if s.refreshSecret == "" {
		jwtSecret, jwtSecretErr := crt.GetOrGenPasswordKeyringPass("jwt_secret") // pragma: allowlist secret
		if jwtSecretErr != nil {                                                 // pragma: allowlist secret
			gl.Log("fatal", fmt.Sprintf("Error retrieving JWT secret key: %v", jwtSecretErr)) // pragma: allowlist secret
			return nil, jwtSecretErr
		}
		s.refreshSecret = jwtSecret // pragma: allowlist secret
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(u.GetID())
	if err != nil {
		gl.Log("error", fmt.Sprintf("error generating refresh token for uid: %v: %v", u.GetID(), err))
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	// Store refresh token in database via GDBase repo
	if err := s.tokenRepo.SetRefreshToken(ctx, u.GetID(), refreshToken.ID, refreshToken.ExpiresIn); err != nil {
		gl.Log("error", fmt.Sprintf("error storing token ID for uid: %v: %v", u.GetID(), err))
		return nil, fmt.Errorf("error storing token: %w", err)
	}

	return &sci.TokenPair{
		IDToken:      sci.IDToken{SS: idToken},
		RefreshToken: sci.RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: u.GetID()},
	}, nil
}

// SignOut revokes all refresh tokens for a user
func (s *JWTService) SignOut(ctx context.Context, uid string) error {
	return s.tokenRepo.DeleteUserRefreshTokens(ctx, uid)
}

// ValidateIDToken validates an ID token and returns the user claims
func (s *JWTService) ValidateIDToken(tokenString string) (mdl.UserModel, error) {
	claims, err := s.validateIDToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("unable to validate or parse ID token: %w", err)
	}
	return claims.User, nil
}

// ValidateRefreshToken validates a refresh token string
func (s *JWTService) ValidateRefreshToken(tokenString string) (*sci.RefreshToken, error) {
	claims, err := s.validateRefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("unable to validate or parse refresh token: %w", err)
	}

	tokenUUID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("claims ID could not be parsed as UUID: %w", err)
	}

	return &sci.RefreshToken{
		SS:  tokenString,
		ID:  tokenUUID.String(),
		UID: claims.UID,
	}, nil
}

// RenewToken renews an expired ID token using a valid refresh token
func (s *JWTService) RenewToken(ctx context.Context, refreshToken string) (*sci.TokenPair, error) {
	if len(strings.Split(refreshToken, ".")) != 3 {
		return nil, fmt.Errorf("invalid refresh token format")
	}

	claims, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("unable to validate refresh token: %w", err)
	}

	// Delete the old refresh token
	if err := s.tokenRepo.DeleteRefreshToken(ctx, claims.UID, claims.ID); err != nil {
		return nil, fmt.Errorf("error deleting refresh token: %w", err)
	}

	// Get user from ID token claims (assuming UID is the user ID)
	idClaims, err := s.validateIDToken(claims.UID)
	if err != nil {
		return nil, fmt.Errorf("error validating id token: %w", err)
	}

	return s.NewPairFromUser(ctx, idClaims.User, claims.ID)
}

// generateIDToken creates a signed ID token for a user
func (s *JWTService) generateIDToken(u mdl.UserModel) (string, error) {
	if s.privKey == nil { // pragma: allowlist secret
		return "", fmt.Errorf("private key is nil") // pragma: allowlist secret
	}
	if u == nil {
		return "", fmt.Errorf("user model is nil")
	}

	unixTime := time.Now().Unix()
	tokenExp := unixTime + s.idExpirationSecs

	claims := jwtIDTokenClaims{
		User: u.GetUserObj(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Unix(unixTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(tokenExp, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(s.privKey) // pragma: allowlist secret
	if err != nil {
		return "", fmt.Errorf("failed to sign ID token: %w", err)
	}

	return ss, nil
}

// generateRefreshToken creates a signed refresh token
func (s *JWTService) generateRefreshToken(uid string) (*jwtRefreshTokenData, error) {
	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(s.refreshExpirationSecs) * time.Second)
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token ID: %w", err)
	}

	claims := jwtRefreshTokenClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(tokenExp),
			ID:        tokenID.String(),
		},
	}

	if s.refreshSecret == "" {
		return nil, fmt.Errorf("refresh token secret key is empty")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &jwtRefreshTokenData{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}

// validateIDToken validates and parses an ID token
func (s *JWTService) validateIDToken(tokenString string) (*jwtIDTokenClaims, error) {
	claims := &jwtIDTokenClaims{}

	if tokenString == "" {
		return nil, fmt.Errorf("token string is empty")
	}
	if s.pubKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	if len(strings.Split(tokenString, ".")) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	if !strings.HasPrefix(tokenString, "ey") {
		return nil, fmt.Errorf("invalid JWT token")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*jwtIDTokenClaims)
	if !ok {
		return nil, fmt.Errorf("token valid but couldn't parse claims")
	}
	if claims.User == nil {
		return nil, fmt.Errorf("user claims are nil")
	}
	if claims.ExpiresAt.Time.Unix() < time.Now().Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// validateRefreshToken validates and parses a refresh token
func (s *JWTService) validateRefreshToken(tokenString string) (*jwtRefreshTokenClaims, error) {
	claims := &jwtRefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("refresh token is invalid")
	}

	claims, ok := token.Claims.(*jwtRefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("refresh token valid but couldn't parse claims")
	}

	return claims, nil
}
