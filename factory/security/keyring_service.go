package security

import (
	krs "github.com/rafa-mori/gobe/internal/app/security/external"
	sci "github.com/rafa-mori/gobe/internal/app/security/interfaces"
)

type KeyringService interface{ sci.IKeyringService }

func NewKeyringService(service, name string) KeyringService {
	return krs.NewKeyringService(service, name)
}
