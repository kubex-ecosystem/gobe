package gdbasez

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// NewBridgeFromService validates dbService, injects dbName in the context (under gl.ContextDBNameKey)
// and returns the updated context and a *Bridge instance. It centralizes the repeated logic
// used across routes/controllers that need a Bridge constructed from a DBService.
func NewBridgeFromService(ctx context.Context, dbService *svc.DBServiceImpl) (context.Context, *Bridge, error) {
	if dbService == nil {
		gl.Log("error", "Database service is nil in NewBridgeFromService")
		return nil, nil, fmt.Errorf("database service is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	dbCfg := dbService.GetConfig(ctx)
	if dbCfg == nil {
		gl.Log("error", "Database config is nil in NewBridgeFromService")
		return nil, nil, fmt.Errorf("database config is nil")
	}

	dbName := dbCfg.GetDBName()
	if dbName == "" {
		gl.Log("warn", "Database name is empty in DBConfig in NewBridgeFromService")
	}

	ctx = context.WithValue(ctx, gl.ContextDBNameKey, dbName)

	bridge := NewBridge(ctx, dbService, dbName)
	return ctx, bridge, nil
}
