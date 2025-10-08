package security

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	mdl "github.com/kubex-ecosystem/gdbase/factory/models"
)

// NewTokenRepo creates a token repository using GDBase factory.
// This repository is used for persisting refresh tokens in the database.
func NewTokenRepo(ctx context.Context, dbService *svc.DBServiceImpl) mdl.ITokenRepo {
	return mdl.NewTokenRepo(ctx, dbService)
}
