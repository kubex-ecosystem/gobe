// Package gateway provides an interface for the authentication gateway.
package gateway

import (
	ii "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	fsi "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
)

type AuthManager = ii.AuthManager

func NewAuthManager(certService fsi.CertService) (*AuthManager, error) {
	return ii.NewAuthManager(certService)
}
