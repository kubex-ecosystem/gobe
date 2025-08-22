package gateway

import (
	ii "github.com/rafa-mori/gobe/internal/app/security/authentication"
	fsi "github.com/rafa-mori/gobe/internal/app/security/certificates"
)

type AuthManager = ii.AuthManager

func NewAuthManager(certService fsi.CertService) (*AuthManager, error) {
	return ii.NewAuthManager(certService)
}
