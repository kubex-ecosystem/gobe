package security

import (
	s "github.com/rafa-mori/gdbase/factory"
	sau "github.com/rafa-mori/gobe/internal/security/authentication"
	sci "github.com/rafa-mori/gobe/internal/security/interfaces"
)

func NewTokenClient(certService sci.ICertService, db s.DBService) sci.TokenClient {
	return sau.NewTokenClient(certService, db)
}
