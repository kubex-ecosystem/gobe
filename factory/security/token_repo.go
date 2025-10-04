package security

import (
	"context"

	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
)

func NewTokenRepo(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) sci.TokenRepo {
	return sau.NewTokenRepo(ctx, dbService, dbName)
}
