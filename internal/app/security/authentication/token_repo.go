package authentication

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	l "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gobe/internal/app/security/models"
)

type TokenRepoImpl struct{ *sql.Conn }

// NewTokenRepo creates a TokenRepo using the provided DBService. It infers
// dbName from the DBService config via gdbasez.NewBridgeFromService.
func NewTokenRepo(ctx context.Context, dbSrv *svc.DBServiceImpl) sci.TokenRepo {
	db, err := dbSrv.GetConnection(ctx, 60*time.Second)
	if err != nil {
		panic(fmt.Sprintf("failed to get database from DBService: %v", err))
	}
	// Return the TokenRepoImpl instance
	return &TokenRepoImpl{db}
}

func (g *TokenRepoImpl) TableName() string {
	return "refresh_tokens"
}

func (g *TokenRepoImpl) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error {
	expirationTime := time.Now().Add(expiresIn)
	token := &models.RefreshTokenModel{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expirationTime,
	}
	res, err := g.Conn.ExecContext(ctx, "INSERT INTO "+g.TableName()+" (user_id, token_id, expires_at) VALUES (?, ?, ?)", token.UserID, token.TokenID, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after inserting refresh token: %w", err)
	}
	// if rowsAffected == 0 {
	// 	return fmt.Errorf("no rows affected when inserting refresh token")
	// }
	return nil
}

func (g *TokenRepoImpl) DeleteRefreshToken(ctx context.Context, userID string, prevTokenID string) error {
	// if err := g.WithContext(ctx).Where("user_id = ? AND token_id = ?", userID, prevTokenID).Delete(&models.RefreshTokenModel{}).Error; err != nil && err != gorm.ErrRecordNotFound {
	// 	// Ignore ErrRecordNotFound as it indicates no tokens were found for the user and is not an error condition.
	// 	return fmt.Errorf("failed to delete refresh token: %w", err)
	// }
	_, err := g.Conn.ExecContext(ctx, "DELETE FROM "+g.TableName()+" WHERE user_id = ? AND token_id = ?", userID, prevTokenID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (g *TokenRepoImpl) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	// if err := g.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.RefreshTokenModel{}).Error; err != nil && err != gorm.ErrRecordNotFound {
	// 	// Ignore ErrRecordNotFound as it indicates no tokens were found for the user and is not an error condition.
	// 	return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	// }
	_, err := g.Conn.ExecContext(ctx, "DELETE FROM "+g.TableName()+" WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}

func (g *TokenRepoImpl) GetRefreshToken(ctx context.Context, tokenID string) (*models.RefreshTokenModel, error) {
	var token models.RefreshTokenModel
	// if err := g.WithContext(ctx).Where("token_id = ?", tokenID).First(&token).Error; err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	// }
	row := g.Conn.QueryRowContext(ctx, "SELECT user_id, token_id, expires_at FROM "+g.TableName()+" WHERE token_id = ?", tokenID)
	err := row.Scan(&token.UserID, &token.TokenID, &token.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}
	return &token, nil
}

func getDefaultDatabaseConfig(ctx context.Context, dbService *svc.DBServiceImpl) (svc.DBConfig, error) {
	if dbService == nil {
		return nil, fmt.Errorf("dbService cannot be nil")
	}
	cfg := dbService.GetConfig(ctx).(*svc.DBConfigImpl)
	if cfg == nil {
		return nil, fmt.Errorf("dbService config cannot be nil")
	}
	var targetDatabase *svc.DatabaseImpl
	for _, dbSettings := range cfg.Databases {
		if dbSettings.IsDefault {
			targetDatabase = dbSettings
			break
		}
	}
	if targetDatabase == nil {
		return nil, fmt.Errorf("no default database configuration found")
	}
	ref := svc.NewReference("default")

	return &svc.DBConfigImpl{
		Name:     "default",
		FilePath: "",
		Logger:   l.GetLogger("tk_db"),
		//Mutexes:        tp.NewMutexesType(),
		IsConfidential: true,
		Debug:          false,
		AutoMigrate:    false,
		JWT:            &svc.JWT{},
		Reference:      ref.(*svc.ReferenceImpl),
		Enabled:        targetDatabase.Enabled,
		Databases: map[string]*svc.DatabaseImpl{
			targetDatabase.Name: targetDatabase,
		},
		MongoDB:   &svc.MongoDB{},
		Messagery: &svc.Messagery{},
	}, nil
}
