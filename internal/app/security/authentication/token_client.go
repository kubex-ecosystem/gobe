package authentication

import (
	"crypto/rsa"
	"fmt"

	s "github.com/kubex-ecosystem/gdbase/factory"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	kri "github.com/kubex-ecosystem/gobe/internal/app/security/external"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	common "github.com/kubex-ecosystem/gobe/internal/commons"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type TokenClientImpl struct {
	mapper                ci.IMapper[sci.TSConfig]
	dbSrv                 s.DBService
	crtSrv                sci.ICertService
	keyringService        sci.IKeyringService
	TokenService          sci.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
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
		t.crtSrv = crt.NewCertService(common.DefaultGoBEKeyPath, common.DefaultGoBECertPath)
		if t.crtSrv == nil {
			gl.Log("fatal", "crtService is nil, unable to create a new one")
		}
	}
	privKey, err := t.crtSrv.GetPrivateKey()
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error reading private key file: %v", err))
		return nil, 0, 0, err
	}
	pubKey, pubKeyErr := t.crtSrv.GetPublicKey()
	if pubKeyErr != nil {
		gl.Log("error", fmt.Sprintf("Error reading public key file: %v", pubKeyErr))
		return nil, 0, 0, pubKeyErr
	}

	dB, dbErr := t.dbSrv.GetDB()
	if dbErr != nil {
		gl.Log("error", fmt.Sprintf("Error getting DB: %v", dbErr))
		return nil, 0, 0, dbErr
	}

	// Garantir valores padrão seguros
	if t.IDExpirationSecs == 0 {
		t.IDExpirationSecs = 3600 // 1 hora
	}
	if t.RefreshExpirationSecs == 0 {
		t.RefreshExpirationSecs = 604800 // 7 dias
	}
	if t.keyringService == nil {
		t.keyringService = kri.NewKeyringService(common.KeyringService, fmt.Sprintf("gobe-%s", "jwt_secret"))
		if t.keyringService == nil {
			gl.Log("error", fmt.Sprintf("Error creating keyring service: %v", err))
			return nil, 0, 0, err
		}
	}

	tokenService := NewTokenService(&sci.TSConfig{
		TokenRepository:  NewTokenRepo(dB),
		IDExpirationSecs: t.IDExpirationSecs,
		PubKey:           pubKey,
		PrivKey:          privKey,
		TokenClient:      t,
		DBService:        &t.dbSrv,
		KeyringService:   t.keyringService,
	})

	return tokenService, t.IDExpirationSecs, t.RefreshExpirationSecs, nil
}

func NewTokenClient(crtService sci.ICertService, dbService s.DBService) sci.TokenClient {
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
