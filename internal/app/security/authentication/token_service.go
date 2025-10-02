package authentication

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type idTokenCustomClaims struct {
	User svc.UserModelType `json:"UserImpl"`
	jwt.RegisteredClaims
}
type TokenServiceImpl struct {
	TokenRepository       sci.TokenRepo
	PrivKey               *rsa.PrivateKey
	PubKey                *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func NewTokenService(c *sci.TSConfig) sci.TokenService {
	if c == nil {
		gl.Log("error", "TokenService config is nil")
		return nil
	}
	var idExpirationSecs, refreshExpirationSecs int64
	if c.IDExpirationSecs == 0 {
		idExpirationSecs = 3600 // Default to 1 hour
	} else {
		idExpirationSecs = c.IDExpirationSecs
	}
	if c.RefreshExpirationSecs == 0 {
		refreshExpirationSecs = 604800 // Default to 7 days
	} else {
		refreshExpirationSecs = c.RefreshExpirationSecs
	}
	tsrv := &TokenServiceImpl{
		TokenRepository:       c.TokenRepository,
		PrivKey:               c.PrivKey,
		PubKey:                c.PubKey,
		RefreshSecret:         c.RefreshSecret,
		IDExpirationSecs:      idExpirationSecs,
		RefreshExpirationSecs: refreshExpirationSecs,
	}
	return tsrv
}

func (s *TokenServiceImpl) NewPairFromUser(ctx context.Context, u svc.UserModel, prevTokenID string) (*sci.TokenPair, error) {
	if prevTokenID != "" {
		if err := s.TokenRepository.DeleteRefreshToken(ctx, u.GetID(), prevTokenID); err != nil {
			return nil, fmt.Errorf("could not delete previous refresh token for uid: %v, tokenID: %v", u.GetID(), prevTokenID)
		}
	}

	idToken, err := generateIDToken(u, s.PrivKey, s.IDExpirationSecs)
	if err != nil {
		return nil, fmt.Errorf("error generating id token for uid: %v: %v", u.GetID(), err)
	}

	if s.RefreshSecret == "" {
		jwtSecret, jwtSecretErr := crt.GetOrGenPasswordKeyringPass("jwt_secret")
		if jwtSecretErr != nil {
			gl.Log("fatal", fmt.Sprintf("Error retrieving JWT secret key: %v", jwtSecretErr))
			return nil, jwtSecretErr
		}
		s.RefreshSecret = jwtSecret
	}

	refreshToken, err := generateRefreshToken(u.GetID(), s.RefreshSecret, s.RefreshExpirationSecs)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token for uid: %v: %v", u.GetID(), err)
	}

	if err := s.TokenRepository.SetRefreshToken(ctx, u.GetID(), refreshToken.ID, refreshToken.ExpiresIn); err != nil {
		return nil, fmt.Errorf("error storing token ID for uid: %v: %v", u.GetID(), err)
	}

	return &sci.TokenPair{
		IDToken:      sci.IDToken{SS: idToken},
		RefreshToken: sci.RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: u.GetID()},
	}, nil
}
func (s *TokenServiceImpl) SignOut(ctx context.Context, uid string) error {
	return s.TokenRepository.DeleteUserRefreshTokens(ctx, uid)
}
func (s *TokenServiceImpl) ValidateIDToken(tokenString string) (svc.UserModel, error) {
	// Garantir que o segredo de atualização esteja configurado
	if s.RefreshSecret == "" || len(s.RefreshSecret) < 32 {
		jwtSecret, jwtSecretErr := crt.GetOrGenPasswordKeyringPass("jwt_secret")
		if jwtSecretErr != nil {
			gl.Log("fatal", fmt.Sprintf("Error retrieving JWT secret key: %v", jwtSecretErr))
			return nil, fmt.Errorf("error retrieving JWT secret key: %v", jwtSecretErr)
		}
		s.RefreshSecret = jwtSecret
	}

	// Validar o token usando a chave pública
	claims, err := validateIDToken(tokenString, s.PubKey)
	if err != nil {
		return nil, fmt.Errorf("unable to validate or parse ID token: %v", err)
	}

	return claims.User, nil
}
func (s *TokenServiceImpl) ValidateRefreshToken(tokenString string) (*sci.RefreshToken, error) {
	claims, claimsErr := validateRefreshToken(tokenString, s.RefreshSecret)
	if claimsErr != nil {
		return nil, fmt.Errorf("unable to validate or parse refresh token for token string %s: %v", tokenString, claimsErr)
	}
	tokenUUID, tokenUUIDErr := uuid.Parse(claims.ID)
	if tokenUUIDErr != nil {
		return nil, fmt.Errorf("claims ID could not be parsed as UUID: %s: %v", claims.UID, tokenUUIDErr)
	}
	return &sci.RefreshToken{
		SS:  tokenString,
		ID:  tokenUUID.String(),
		UID: claims.UID,
	}, nil
}
func (s *TokenServiceImpl) RenewToken(ctx context.Context, refreshToken string) (*sci.TokenPair, error) {
	if len(strings.Split(refreshToken, ".")) != 3 {
		return nil, fmt.Errorf("invalid refresh token format for token string: %s", refreshToken)
	}

	claims, err := validateRefreshToken(refreshToken, s.RefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to validate or parse refresh token for token string %s: %v", refreshToken, err)
	}
	if err := s.TokenRepository.DeleteRefreshToken(ctx, claims.UID, claims.ID); err != nil {
		return nil, fmt.Errorf("error deleting refresh token: %v", err)
	}
	idCClaims, idCClaimsErr := validateIDToken(claims.UID, s.PubKey)
	if idCClaimsErr != nil {
		return nil, fmt.Errorf("error validating id token: %v", idCClaimsErr)
	}
	return s.NewPairFromUser(ctx, idCClaims.User, claims.ID)
}

type refreshTokenData struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}
type refreshTokenCustomClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

func generateIDToken(u svc.UserModel, key *rsa.PrivateKey, exp int64) (string, error) {
	if key == nil {
		gl.Log("error", "Private key is nil")
		return "", fmt.Errorf("private key is nil")
	}
	if u == nil {
		gl.Log("error", "User model is nil")
		return "", fmt.Errorf("user model is nil")
	}
	if exp <= 0 {
		exp = 3600 // Default to 1 hour
	}
	unixTime := time.Now().Unix()
	tokenExp := unixTime + exp
	claims := idTokenCustomClaims{
		User: u.GetUserObj(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Unix(unixTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(tokenExp, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		gl.Log("error", "Error signing ID token: %v", err)
		return "", fmt.Errorf("failed to sign ID token: %w", err)
	}

	gl.Log("info", "ID token generated successfully for user: %s", u.GetID())
	return ss, nil
}
func generateRefreshToken(uid string, key string, exp int64) (*refreshTokenData, error) {
	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(exp) * time.Second)
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token ID")
	}

	claims := refreshTokenCustomClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(tokenExp),
			ID:        tokenID.String(),
		},
	}

	// Create the token using the signing method and claims
	// Note: The signing method is not used in the JWT token, but it's required for signing
	// the token with the secret key.
	// The key is used to sign the token, and the signing method is used to verify it.
	if key == "" {
		return nil, fmt.Errorf("refresh token secret key is empty")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %v", err)
	}

	return &refreshTokenData{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}
func validateIDToken(tokenString string, key *rsa.PublicKey) (*idTokenCustomClaims, error) {
	claims := &idTokenCustomClaims{}

	// Check if the token string is empty
	if tokenString == "" {
		gl.Log("error", "Token string is empty")
		return nil, fmt.Errorf("token string is empty")
	}
	// Check if the key is nil
	if key == nil {
		gl.Log("error", "Public key is nil")
		return nil, fmt.Errorf("public key is nil")
	}

	// Check if the token string is in the correct format
	if len(strings.Split(tokenString, ".")) != 3 {
		gl.Log("error", "Invalid token format")
		return nil, fmt.Errorf("invalid token format")
	}

	// Check if the token string is a valid JWT token
	if !strings.HasPrefix(tokenString, "ey") {
		gl.Log("error", "Invalid JWT token")
		return nil, fmt.Errorf("invalid JWT token")
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			gl.Log("error", fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error parsing token: %v", err))
		return nil, fmt.Errorf("error parsing token: %v", err)
	}
	if !token.Valid {
		gl.Log("error", "Token is invalid")
		return nil, fmt.Errorf("token is invalid")
	}
	claims, ok := token.Claims.(*idTokenCustomClaims)
	if !ok {
		gl.Log("error", "Token valid but couldn't parse claims")
		return nil, fmt.Errorf("token valid but couldn't parse claims")
	}
	if claims.User == nil {
		gl.Log("error", "User claims are nil")
		return nil, fmt.Errorf("user claims are nil")
	}
	if claims.User.GetID() == "" {
		gl.Log("error", "User ID is empty")
		return nil, fmt.Errorf("user ID is empty")
	}
	if claims.User.GetRoleID() == "" {
		gl.Log("error", "User role ID is empty")
		return nil, fmt.Errorf("user role ID is empty")
	}
	if claims.User.GetEmail() == "" {
		gl.Log("error", "User email is empty")
		return nil, fmt.Errorf("user email is empty")
	}
	if claims.User.GetUsername() == "" {
		gl.Log("error", "User username is empty")
		return nil, fmt.Errorf("user username is empty")
	}
	if claims.ExpiresAt.Time.Unix() < time.Now().Unix() || claims.ExpiresAt.Time.Unix() <= 0 {
		return nil, fmt.Errorf("token has expired")
	}
	if claims.IssuedAt.Time.Unix() > claims.ExpiresAt.Time.Unix() || claims.IssuedAt.Time.Unix() <= 0 {
		return nil, fmt.Errorf("token issued at time is greater than expiration time")
	}
	if claims.IssuedAt.Time.Unix() <= 0 {
		return nil, fmt.Errorf("token issued at time is less than or equal to zero")
	}

	return claims, nil

}
func validateRefreshToken(tokenString string, key string) (*refreshTokenCustomClaims, error) {
	claims := &refreshTokenCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("refresh token is invalid")
	}
	claims, ok := token.Claims.(*refreshTokenCustomClaims)
	if !ok {
		return nil, fmt.Errorf("refresh token valid but couldn't parse claims")
	}
	return claims, nil
}
