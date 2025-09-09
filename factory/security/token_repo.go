package security

import (
	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	"gorm.io/gorm"
)

func NewTokenRepo(db *gorm.DB) sci.TokenRepo { return sau.NewTokenRepo(db) }
