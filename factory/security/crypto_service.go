package security

import (
	crp "github.com/kubex-ecosystem/gobe/internal/app/security/crypto"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

type CryptoService interface {
	sci.ICryptoService
}

func NewCryptoService() CryptoService {
	return crp.NewCryptoService()
}
