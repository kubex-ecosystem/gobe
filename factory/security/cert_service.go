package security

import (
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

type CertService interface{ sci.ICertService }

func NewCertService(keyPath, certPath string) CertService {
	return crt.NewCertService(keyPath, certPath)
}
