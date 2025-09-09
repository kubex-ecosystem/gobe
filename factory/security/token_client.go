package security

import (
	s "github.com/kubex-ecosystem/gdbase/factory"
	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

func NewTokenClient(certService sci.ICertService, db s.DBService) sci.TokenClient {
	return sau.NewTokenClient(certService, db)
}
