package authentication

import (
	"context"
	"crypto/rsa"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	mdl "github.com/kubex-ecosystem/gdbase/factory/models"

	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	kri "github.com/kubex-ecosystem/gobe/internal/app/security/external"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"

	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type TokenClientImpl struct {
	mapper                ci.IMapper[*sci.TSConfig]
	dbSrv                 *svc.DBServiceImpl
	crtSrv                sci.ICertService
	keyringService        sci.IKeyringService
	TokenService          sci.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
	tokenRepo             mdl.ITokenRepo
}

func (t *TokenClientImpl) LoadPublicKey() *rsa.PublicKey {
	pubKey, err := t.crtSrv.GetPublicKey()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error reading public key file: %v", err))
		return nil
	}
	return pubKey
}

func (t *TokenClientImpl) LoadPrivateKey() (*rsa.PrivateKey, error) {
	return t.crtSrv.GetPrivateKey()
}
func (t *TokenClientImpl) LoadTokenCfg() (sci.TokenService, int64, int64, error) {
	if t == nil {
		gl.Log("error", "TokenClient is nil, trying to create a new one")
		t = &TokenClientImpl{}
	}
	if t.crtSrv == nil {
		gl.Log("error", "crtService is nil, trying to create a new one")
		t.crtSrv = crt.NewCertService(gl.DefaultGoBEKeyPath, gl.DefaultGoBECertPath) // pragma: allowlist secret
		if t.crtSrv == nil {
			gl.Log("fatal", "crtService is nil, unable to create a new one") // pragma: allowlist secret
		}
	}

	// Get RSA keys
	privKey, err := t.crtSrv.GetPrivateKey() // pragma: allowlist secret
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error reading private key file: %v", err))
		return nil, 0, 0, err
	}
	pubKey, pubKeyErr := t.crtSrv.GetPublicKey() // pragma: allowlist secret
	if pubKeyErr != nil {
		gl.Log("error", fmt.Sprintf("Error reading public key file: %v", pubKeyErr))
		return nil, 0, 0, pubKeyErr
	}

	ctx := context.Background()

	// Garantir valores padrão seguros
	if t.IDExpirationSecs == 0 {
		t.IDExpirationSecs = 3600 // 1 hora
	}
	if t.RefreshExpirationSecs == 0 {
		t.RefreshExpirationSecs = 604800 // 7 dias
	}

	// Setup keyring service
	if t.keyringService == nil {
		t.keyringService = kri.NewKeyringService(gl.KeyringService, fmt.Sprintf("gobe-%s", "jwt_secret"))
		if t.keyringService == nil {
			gl.Log("error", fmt.Sprintf("Error creating keyring service: %v", err))
			return nil, 0, 0, err
		}
	}

	// Get or generate JWT secret
	jwtSecret, jwtSecretErr := crt.GetOrGenPasswordKeyringPass("jwt_secret") // pragma: allowlist secret
	if jwtSecretErr != nil {                                                 // pragma: allowlist secret
		gl.Log("fatal", fmt.Sprintf("Error retrieving JWT secret key: %v", jwtSecretErr))
		return nil, 0, 0, jwtSecretErr
	}

	// Create token repository via GDBase factory
	if t.tokenRepo == nil {
		t.tokenRepo = mdl.NewTokenRepo(ctx, t.dbSrv)
		if t.tokenRepo == nil {
			gl.Log("error", "Failed to create token repository")
			return nil, 0, 0, fmt.Errorf("failed to create token repository")
		}
	}

	// Create JWT service using new JWTService
	jwtService := NewJWTService(
		t.tokenRepo,
		privKey,
		pubKey,
		jwtSecret,
		t.IDExpirationSecs,
		t.RefreshExpirationSecs,
	)

	if jwtService == nil {
		gl.Log("error", "Failed to create JWT service")
		return nil, 0, 0, fmt.Errorf("failed to create JWT service")
	}

	// Wrap JWTService to implement TokenService interface
	tokenService := &TokenServiceWrapper{jwtService: jwtService}

	return tokenService, t.IDExpirationSecs, t.RefreshExpirationSecs, nil
}

func NewTokenClient(crtService sci.ICertService, dbService *svc.DBServiceImpl) *TokenClientImpl {
	if crtService == nil {
		gl.Log("error", fmt.Sprintf("error reading private key file: %v", "crtService is nil"))
		return nil
	}
	tokenClient := &TokenClientImpl{
		crtSrv: crtService,
		dbSrv:  dbService,
	}

	return tokenClient
}
