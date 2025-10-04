// Package discord implements the Discord adapter for the chatbot service.
package discord

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type Adapter struct {
	session        *discordgo.Session // nil in dev mode
	config         bootstrap.DiscordConfig
	messageHandler atomic.Value // func(interfaces.Message)
}

func NewAdapter(cfg bootstrap.DiscordConfig, purpose string) (interfaces.IAdapter, error) {
	// dev mode: no session
	if cfg.Bot.Token == "dev_token" {
		ad := &Adapter{config: cfg}
		return ad, nil
	}

	prefix := "Bearer"
	switch strings.ToLower(purpose) {
	case "chatbot", "bot":
		prefix = "Bot"
	}

	s, err := discordgo.New(prefix + " " + cfg.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	ad := &Adapter{session: s, config: cfg}
	s.AddHandler(ad.readyHandler)
	s.AddHandler(ad.messageCreateHandler)
	// if using slash/interactions, add handler here
	return ad, nil
}

func (a *Adapter) Connect() error {
	if a.session == nil {
		gl.Log("info", "Discord adapter in dev mode - not connecting")
		return nil
	}
	if err := a.session.Open(); err != nil {
		return fmt.Errorf("open session: %w", err)
	}
	gl.Log("info", "Discord session opened")
	return nil
}

func (a *Adapter) Disconnect() error {
	if a.session == nil {
		return nil
	}
	return a.session.Close()
}

func (a *Adapter) OnMessage(h func(interfaces.Message)) {
	a.messageHandler.Store(h) // thread-safe swap
}

func (a *Adapter) SendMessage(channelID, content string, opts ...interfaces.SendOptions) error {
	if a.session == nil {
		gl.Log("info", fmt.Sprintf("Dev mode - would send to %s: %s", channelID, content))
		return nil
	}
	_, err := a.session.ChannelMessageSend(channelID, content)
	if err != nil {
		gl.Log("error", fmt.Sprintf("send message: %v", err))
		return err
	}
	return nil
}

func (a *Adapter) GetChannels(guildID string) ([]interfaces.Channel, error) {
	if a.session == nil {
		return []interfaces.Channel{
			{ID: "dev_channel_1", Name: "general"},
			{ID: "dev_channel_2", Name: "random"},
		}, nil
	}
	cs, err := a.session.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}
	out := make([]interfaces.Channel, 0, len(cs))
	for _, c := range cs {
		out = append(out, interfaces.Channel{
			ID: c.ID, Name: c.Name, GuildID: guildID, Private: c.Type == discordgo.ChannelTypeDM,
		})
	}
	return out, nil
}

func (a *Adapter) PingAdapter(msg string) error {
	if a.session == nil {
		gl.Log("info", "Discord dev mode - ping skipped")
		return nil
	}
	// ensure we have user, otherwise get it:
	if a.session.State.User == nil {
		if _, err := a.session.User("@me"); err != nil {
			return fmt.Errorf("ping: %w", err)
		}
	}
	gl.Log("info", fmt.Sprintf("discord ping: %s", msg))
	return nil
}

// GetMessageHandler returns the current message handler (for testing)
func (a *Adapter) GetMessageHandler() func(interfaces.Message) {
	hv := a.messageHandler.Load()
	if hv == nil {
		return nil
	}
	return hv.(func(interfaces.Message))
}

/* ---------- Private handlers ---------- */

func (a *Adapter) readyHandler(_ *discordgo.Session, ev *discordgo.Ready) {
	gl.Log("info", fmt.Sprintf("Discord logged as %s#%s, guilds: %d", ev.User.Username, ev.User.Discriminator, len(ev.Guilds)))
	for _, g := range ev.Guilds {
		gl.Log("info", fmt.Sprintf(" - %s (%s)", g.Name, g.ID))
	}
}

func (a *Adapter) messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// dev / no handler
	hv := a.messageHandler.Load()
	if hv == nil || s == nil {
		return
	}
	h := hv.(func(interfaces.Message))

	// ignore own bot
	if s.State != nil && s.State.User != nil && m.Author != nil && m.Author.ID == s.State.User.ID {
		return
	}

	msg := ToNeutralMessage(m)
	h(msg)
}

/* ---------- Centralized conversion ---------- */

// ToNeutralMessage converts Discord message to neutral format (exported for testing)
func ToNeutralMessage(m *discordgo.MessageCreate) interfaces.Message {
	ts, _ := time.Parse(time.RFC3339, m.Timestamp.String())
	att := make([]interfaces.Attachment, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		att = append(att, interfaces.Attachment{
			ID: a.ID, Name: a.Filename, Size: a.Size, URL: a.URL, MimeType: a.ContentType,
		})
	}
	return interfaces.Message{
		ID:          m.ID,
		ChannelID:   m.ChannelID,
		GuildID:     m.GuildID,
		User:        interfaces.User{ID: m.Author.ID, Username: m.Author.Username, Discriminator: m.Author.Discriminator},
		Role:        interfaces.RoleUser,
		Content:     m.Content,
		Timestamp:   ts,
		Attachments: att,
	}
}
