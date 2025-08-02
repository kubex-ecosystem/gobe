package security

import (
	sau "github.com/rafa-mori/gobe/internal/security/authentication"
	sci "github.com/rafa-mori/gobe/internal/security/interfaces"
	"gorm.io/gorm"
)

func NewTokenRepo(db *gorm.DB) sci.TokenRepo { return sau.NewTokenRepo(db) }
