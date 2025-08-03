package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	//"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	l "github.com/rafa-mori/logz"

	sau "github.com/rafa-mori/gobe/factory/security"
	cm "github.com/rafa-mori/gobe/internal/common"
	crt "github.com/rafa-mori/gobe/internal/security/certificates"
	sci "github.com/rafa-mori/gobe/internal/security/interfaces"
	srv "github.com/rafa-mori/gobe/internal/services"
	gl "github.com/rafa-mori/gobe/logger"
)

type AuthenticationMiddleware struct {
	contractapi.Contract
	CertService  sci.ICertService
	TokenService sci.TokenService
}

func NewTokenService(config *srv.IDBConfig, logger l.Logger) (sci.TokenService, sci.ICertService, error) {
	if logger == nil {
		logger = l.GetLogger("GoBE")
	}
	var err error
	crtService := crt.NewCertService(os.ExpandEnv(cm.DefaultGoBEKeyPath), os.ExpandEnv(cm.DefaultGoBECertPath))

	dbService, err := srv.NewDBService(config, logger)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DBService: %v", err))
		return nil, nil, fmt.Errorf("❌ Erro ao inicializar DBService: %v", err)
	}

	tkClient := sau.NewTokenClient(crtService, dbService)
	if tkClient == nil {
		gl.Log("error", "❌ Erro ao inicializar TokenClient")
		return nil, nil, fmt.Errorf("❌ Erro ao inicializar TokenClient")
	}
	tkService, _, _, err := tkClient.LoadTokenCfg()
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar TokenService: %v", err))
		return nil, nil, fmt.Errorf("❌ Erro ao inicializar TokenService: %v", err)
	}

	return tkService, crtService, err
}

func NewAuthenticationMiddleware(tokenService sci.TokenService, certService sci.ICertService, err error) gin.HandlerFunc {
	authMiddleware := &AuthenticationMiddleware{
		CertService:  certService,
		TokenService: tokenService,
	}

	if authMiddleware.CertService == nil || authMiddleware.TokenService == nil || err != nil {
		return func(c *gin.Context) {
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate. Please try again later."})
				c.Abort()
				return
			} else {
				gl.Log("error", "❌ Erro ao inicializar AuthenticationMiddleware: CertService or TokenService is nil")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate. Please try again later."})
				c.Next()
			}
		}
	}

	return func(c *gin.Context) {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate. Please try again later."})
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func (a *AuthenticationMiddleware) ValidateJWT(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access Denied"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access Denied"})
			c.Abort()
			return
		}

		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access Denied"})
			c.Abort()
			return
		}

		// Criando um contexto com o usuário autenticado
		ctx := context.WithValue(c.Request.Context(), "user", claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (a *AuthenticationMiddleware) validateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	publicK, err := a.CertService.GetPublicKey()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error getting public key: %v", err))
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			gl.Log("error", fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicK, nil
	})

	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, fmt.Errorf("Access Denied")
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Access Denied")
}
