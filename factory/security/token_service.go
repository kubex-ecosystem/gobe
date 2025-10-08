package security

import (
	mdl "github.com/kubex-ecosystem/gdbase/factory/models"
)

// NewTokenService creates a token service using GDBase factory.
// This service handles token CRUD operations at the persistence layer.
func NewTokenService(tokenRepo mdl.ITokenRepo) mdl.ITokenService {
	return mdl.NewTokenService(tokenRepo)
}
