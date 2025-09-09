// Package discord provides an adapter for interacting with Discord's API using the discordgo library.
package discord

import (
	"fmt"
	"log"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/config"

	"github.com/bwmarrin/discordgo"
)

type Adapter struct {
	api         *discordgo.Identify
	invite      *discordgo.Invite
	application *discordgo.Application

	// Session is nil in dev mode
	// where we don't connect to Discord
	session        *discordgo.Session
	config         config.DiscordConfig
	messageHandler func(Message)
}

type Message struct {
	ID          string                         `json:"id"`
	ChannelID   string                         `json:"channel_id"`
	GuildID     string                         `json:"guild_id"`
	Author      *discordgo.User                `json:"author"`
	Content     string                         `json:"content"`
	Timestamp   time.Time                      `json:"timestamp"`
	Attachments []*discordgo.MessageAttachment `json:"attachments"`
}

func NewAdapter(config config.DiscordConfig) (*Adapter, error) {
	// Dev mode - don't try to connect to Discord
	if config.Bot.Token == "dev_token" {
		adapter := &Adapter{
			session: nil,
			config:  config,
		}
		return adapter, nil
	}
	var err error
	session, err := discordgo.New("Bot " + config.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	adapter := &Adapter{
		api:         &session.Identify,
		invite:      nil, // invite,
		application: session.State.Application,
		session:     session,
		config:      config,
	}

	// Register event handlers
	session.AddHandler(adapter.messageCreateHandler)
	session.AddHandler(adapter.readyHandler)
	session.Identify.Intents |= discordgo.IntentsGuilds
	session.Identify.Intents |= discordgo.IntentsGuildMembers
	session.Identify.Intents |= discordgo.IntentsGuildPresences
	session.Identify.Intents |= discordgo.IntentsGuildVoiceStates

	return adapter, nil
}

func (a *Adapter) Connect() error {
	if a.session == nil {
		log.Println("Discord adapter in dev mode - not connecting to Discord")
		return nil
	}
	return a.session.Open()
}

func (a *Adapter) Disconnect() error {
	if a.session == nil {
		return nil
	}
	return a.session.Close()
}

func (a *Adapter) OnMessage(handler func(Message)) {
	a.messageHandler = handler
}

func (a *Adapter) SendMessage(channelID, content string) error {
	if a.session == nil {
		log.Printf("Dev mode - would send message to %s: %s", channelID, content)
		return nil
	}
	log.Printf("📤 Enviando mensagem para canal %s: %s", channelID, content)
	_, err := a.session.ChannelMessageSend(channelID, content)
	if err != nil {
		log.Printf("❌ Erro ao enviar mensagem: %v", err)
		return err
	}
	log.Printf("✅ Mensagem enviada com sucesso!")
	return nil
}

func (a *Adapter) GetChannels(guildID string) ([]*discordgo.Channel, error) {
	if a.session == nil {
		// Return mock channels in dev mode
		return []*discordgo.Channel{
			{ID: "dev_channel_1", Name: "general", Type: discordgo.ChannelTypeGuildText},
			{ID: "dev_channel_2", Name: "random", Type: discordgo.ChannelTypeGuildText},
		}, nil
	}
	return a.session.GuildChannels(guildID)
}

func (a *Adapter) readyHandler(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Discord bot logged in as: %v#%v", event.User.Username, event.User.Discriminator)
	log.Printf("Bot is connected to %d guilds", len(event.Guilds))
	for _, guild := range event.Guilds {
		log.Printf("  - Guild: %s (ID: %s)", guild.Name, guild.ID)
	}
}

func (a *Adapter) messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Skip if no message handler or in dev mode
	if a.messageHandler == nil || s == nil {
		return
	}

	// Ignore bot's own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	log.Printf("📨 Nova mensagem recebida:")
	log.Printf("  - Autor: %s#%s", m.Author.Username, m.Author.Discriminator)
	log.Printf("  - Canal: %s", m.ChannelID)
	log.Printf("  - Conteúdo: %s", m.Content)

	// Parse timestamp
	timestamp, _ := time.Parse(time.RFC3339, m.Timestamp.String())

	message := Message{
		ID:          m.ID,
		ChannelID:   m.ChannelID,
		GuildID:     m.GuildID,
		Author:      m.Author,
		Content:     m.Content,
		Timestamp:   timestamp,
		Attachments: m.Attachments,
	}

	if a.messageHandler != nil {
		a.messageHandler(message)
	}
}

func (a *Adapter) PingDiscord(msg string) error {
	if a.session == nil {
		log.Println("Discord adapter in dev mode - not pinging Discord")
		return nil
	}
	if a.session.State.User == nil {
		_, err := a.session.ChannelMessageSend(a.invite.Channel.ID, msg)
		if err != nil {
			log.Printf("❌ Erro ao enviar mensagem: %v", err)
			return err
		}
		log.Printf("✅ Mensagem de ping enviada com sucesso!")
	}
	return nil
}
