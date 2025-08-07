// Package whatsapp provides models for WhatsApp messages.
package whatsapp

import "time"

// Message represents a WhatsApp message stored in the database.
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
