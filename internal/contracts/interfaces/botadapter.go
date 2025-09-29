package interfaces

import "time"

type Role string

const (
	RoleUser Role = "user"
	RoleBot  Role = "assistant"
)

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

type Attachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}

type Message struct {
	ID          string       `json:"id"`
	ChannelID   string       `json:"channel_id"`
	GuildID     string       `json:"guild_id"`
	User        User         `json:"user"`
	Role        Role         `json:"role"`
	Content     string       `json:"content"`
	Timestamp   time.Time    `json:"timestamp"`
	Attachments []Attachment `json:"attachments"`
}

type SendOptions struct {
	ReplyToID string `json:"reply_to_id"`
	Ephemeral bool   `json:"ephemeral"` // ignored if provider doesn't support
}

type Channel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Private bool   `json:"private"`
	GuildID string `json:"guild_id"`
}

type IAdapter interface {
	Connect() error
	Disconnect() error
	OnMessage(func(Message))        // neutral callback
	SendMessage(channelID, content string, opts ...SendOptions) error
	GetChannels(guildID string) ([]Channel, error)
	PingAdapter(msg string) error
}
