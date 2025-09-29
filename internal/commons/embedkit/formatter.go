package embedkit

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ------- helpers -------

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	type unit struct {
		dur   time.Duration
		label string
	}
	units := []unit{
		{time.Hour * 24, "d"},
		{time.Hour, "h"},
		{time.Minute, "m"},
		{time.Second, "s"},
	}
	out := ""
	rem := d
	for _, u := range units {
		if rem >= u.dur {
			val := rem / u.dur
			rem = rem % u.dur
			out += fmt.Sprintf("%d%s", val, u.label)
			if len(out) > 12 {
				break
			}
		}
	}
	if out == "" {
		out = "0s"
	}
	return out
}

func statusColor(status string) int {
	switch status {
	case "ok", "healthy", "green":
		return 0x57F287 // green
	case "warn", "degraded", "yellow":
		return 0xFEE75C // yellow
	default:
		return 0xED4245 // red
	}
}

func statusEmoji(status string) string {
	switch status {
	case "ok", "healthy", "green":
		return "🟢"
	case "warn", "degraded", "yellow":
		return "🟡"
	default:
		return "🔴"
	}
}

func mkField(name, value string, inline bool) *discordgo.MessageEmbedField {
	if value == "" {
		value = "`n/a`"
	}
	return &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	}
}

// ------- builder -------

type SystemHealth struct {
	Status  string // ok / warn / error
	Version string
	Uptime  time.Duration
	Host    string

	Mem        string // ok/warn/error ou "512MB (68%)"
	Goroutines string // "ok (123)"

	GoBE     string // markdown curto por módulo
	MCP      string
	Analyzer string
}

func BuildStatusEmbed(user, env string, h SystemHealth, swaggerURL, kortexURL, logsURL string) (*discordgo.MessageSend, error) {
	color := statusColor(h.Status)
	desc := fmt.Sprintf("**Comando executado por:** `%s`\n**Ambiente:** `%s`", user, env)

	embed := &discordgo.MessageEmbed{
		Title:       "🖥️ Status do Sistema",
		Description: desc,
		Color:       color,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			mkField("✅ Status", fmt.Sprintf("**%s** %s", h.Status, statusEmoji(h.Status)), true),
			mkField("🏷️ Versão", fmt.Sprintf("`%s`", h.Version), true),
			mkField("⏱️ Uptime", fmt.Sprintf("`%s`", formatDuration(h.Uptime)), true),

			mkField("🌐 Host", fmt.Sprintf("`%s`", h.Host), true),
			mkField("🧠 Memory", fmt.Sprintf("`%s`", h.Mem), true),
			mkField("🧵 Goroutines", fmt.Sprintf("`%s`", h.Goroutines), true),

			mkField("🧩 GoBE", h.GoBE, false),
			mkField("🪄 MCP", h.MCP, false),
			mkField("🔎 Analyzer", h.Analyzer, false),
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Kubex • Control Tower"},
	}

	// botões (link buttons)
	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{Style: discordgo.LinkButton, Label: "Swagger", URL: swaggerURL},
			&discordgo.Button{Style: discordgo.LinkButton, Label: "Kortex", URL: kortexURL},
			&discordgo.Button{Style: discordgo.LinkButton, Label: "Logs", URL: logsURL},
		},
	}

	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row},
	}, nil
}
