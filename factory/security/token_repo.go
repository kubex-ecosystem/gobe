package security

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	sau "github.com/kubex-ecosystem/gobe/internal/app/security/authentication"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
)

// NewTokenRepo creates a token repository using the provided DB service.
// It no longer requires dbName: the dbName is read from the DBService config and
// injected into context by NewBridgeFromService when needed.
func NewTokenRepo(ctx context.Context, dbService *svc.DBServiceImpl) sci.TokenRepo {
	return sau.NewTokenRepo(ctx, dbService)
}
