package security

import (
	krs "github.com/kubex-ecosystem/gobe/internal/app/security/external"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

type KeyringService interface{ sci.IKeyringService }

func NewKeyringService(service, name string) KeyringService {
	return krs.NewKeyringService(service, name)
}
