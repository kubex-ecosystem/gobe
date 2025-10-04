// Package telegram provides models for interacting with the Telegram API.
package telegram

import "time"

// Message represents a Telegram message stored in the database.
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    int64     `json:"chat_id"`
	From      string    `json:"from"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
