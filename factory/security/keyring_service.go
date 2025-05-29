package security

import (
	krs "github.com/rafa-mori/gobe/internal/security/external"
	sci "github.com/rafa-mori/gobe/internal/security/interfaces"
)

type KeyringService interface{ sci.IKeyringService }

func NewKeyringService(service, name string) KeyringService {
	return krs.NewKeyringService(service, name)
}
