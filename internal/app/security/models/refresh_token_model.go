package models

import (
	"time"
)

type RefreshTokenModel struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"index;not null"`
	TokenID   string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}
func (m *RefreshTokenModel) GetID() string {
	return m.TokenID
}
func (m *RefreshTokenModel) GetUserID() string {
	return m.UserID
}
func (m *RefreshTokenModel) GetExpiresAt() time.Time {
	return m.ExpiresAt
}
func (m *RefreshTokenModel) SetExpiresAt(t time.Time) {
	m.ExpiresAt = t
}
func (m *RefreshTokenModel) SetUserID(id string) {
	m.UserID = id
}
func (m *RefreshTokenModel) SetID(id string) {
	m.TokenID = id
}
func (m *RefreshTokenModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}
func (m *RefreshTokenModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}
func (m *RefreshTokenModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}
func (m *RefreshTokenModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}
