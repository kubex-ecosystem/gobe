package security

import (
	sau "github.com/rafa-mori/gobe/internal/app/security/authentication"
	sci "github.com/rafa-mori/gobe/internal/app/security/interfaces"
)

func NewTokenService(c *sci.TSConfig) sci.TokenService {
	return sau.NewTokenService(c)
}
