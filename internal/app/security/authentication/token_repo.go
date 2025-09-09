package authentication

import (
	"context"
	"fmt"
	"time"

	sci "github.com/kubex-ecosystem/gobe/internal/app/security/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/app/security/models"
	"gorm.io/gorm"
)

type TokenRepoImpl struct{ *gorm.DB }

func NewTokenRepo(db *gorm.DB) sci.TokenRepo { return &TokenRepoImpl{db} }

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
	if err := g.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (g *TokenRepoImpl) DeleteRefreshToken(ctx context.Context, userID string, prevTokenID string) error {
	if err := g.WithContext(ctx).Where("user_id = ? AND token_id = ?", userID, prevTokenID).Delete(&models.RefreshTokenModel{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		// Ignore ErrRecordNotFound as it indicates no tokens were found for the user and is not an error condition.
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (g *TokenRepoImpl) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	if err := g.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.RefreshTokenModel{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		// Ignore ErrRecordNotFound as it indicates no tokens were found for the user and is not an error condition.
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}

func (g *TokenRepoImpl) GetRefreshToken(ctx context.Context, tokenID string) (*models.RefreshTokenModel, error) {
	var token models.RefreshTokenModel
	if err := g.WithContext(ctx).Where("token_id = ?", tokenID).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}
	return &token, nil
}
