package security

import (
	crp "github.com/rafa-mori/gobe/internal/app/security/crypto"
	sci "github.com/rafa-mori/gobe/internal/app/security/interfaces"
)

type CryptoService interface {
	sci.ICryptoService
}

func NewCryptoService() CryptoService {
	return crp.NewCryptoService()
}
