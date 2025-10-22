// Package interfaces defines the interfaces for the application services.
package interfaces

// IEmailService defines the interface for sending emails.
type IEmailService interface {
	SendEmail(to, subject, body string) error
}
