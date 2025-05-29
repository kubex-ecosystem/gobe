package tests

// import (
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"testing"
// 	"time"

// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/rafa-mori/gobe/internal/security/authentication"
// 	"github.com/stretchr/testify/assert"
// )

// func TestValidateIDToken(t *testing.T) {
// 	// Generate RSA keys for testing
// 	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
// 	assert.NoError(t, err)
// 	pubKey := &privKey.PublicKey

// 	// Create a valid token
// 	expiresAt := time.Now().Add(time.Hour).Unix()
// 	claims := &authentication.idTokenCustomClaims{
// 		User: &authentication.UserModelType{
// 			ID:       "123",
// 			Username: "testuser",
// 			Email:    "test@example.com",
// 		},
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: expiresAt,
// 			IssuedAt:  time.Now().Unix(),
// 		},
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
// 	tokenString, err := token.SignedString(privKey)
// 	assert.NoError(t, err)

// 	// Validate the token
// 	validatedClaims, err := authentication.validateIDToken(tokenString, pubKey)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, validatedClaims)
// 	assert.Equal(t, "123", validatedClaims.User.GetID())
// 	assert.Equal(t, "testuser", validatedClaims.User.GetUsername())
// 	assert.Equal(t, "test@example.com", validatedClaims.User.GetEmail())

// 	// Test with an expired token
// 	expiredClaims := &authentication.idTokenCustomClaims{
// 		User: &authentication.UserModelType{
// 			ID:       "123",
// 			Username: "testuser",
// 			Email:    "test@example.com",
// 		},
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
// 			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
// 		},
// 	}
// 	expiredToken := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
// 	expiredTokenString, err := expiredToken.SignedString(privKey)
// 	assert.NoError(t, err)

// 	_, err = authentication.validateIDToken(expiredTokenString, pubKey)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "token has expired")

// 	// Test with an invalid token
// 	invalidTokenString := "invalid.token.string"
// 	_, err = authentication.validateIDToken(invalidTokenString, pubKey)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invalid token format")
// }
