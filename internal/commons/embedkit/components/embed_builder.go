package components

import (
	"strings"
	"time"
)

// EmbedBuilder helps to create Discord-friendly embeds without juggling maps manually.
type EmbedBuilder struct {
	title       string
	url         string
	description string
	color       string
	fields      []map[string]interface{}
	footer      map[string]interface{}
	author      map[string]interface{}
	thumbnail   map[string]interface{}
	timestamp   string
}

// NewEmbedBuilder creates a new builder instance with the provided title.
func NewEmbedBuilder(title string) *EmbedBuilder {
	return &EmbedBuilder{
		title:  strings.TrimSpace(title),
		fields: make([]map[string]interface{}, 0, 6),
	}
}

// WithURL sets the embed URL when provided.
func (b *EmbedBuilder) WithURL(url string) *EmbedBuilder {
	if trimmed := strings.TrimSpace(url); trimmed != "" {
		b.url = trimmed
	}
	return b
}

// WithDescription sets the embed description.
func (b *EmbedBuilder) WithDescription(description string) *EmbedBuilder {
	if trimmed := strings.TrimSpace(description); trimmed != "" {
		b.description = trimmed
	}
	return b
}

// WithColor sets the embed color (Discord expects an int encoded as string).
func (b *EmbedBuilder) WithColor(color string) *EmbedBuilder {
	if trimmed := strings.TrimSpace(color); trimmed != "" {
		b.color = trimmed
	}
	return b
}

// WithFooter sets the footer text and optional icon URL.
func (b *EmbedBuilder) WithFooter(text string, iconURL ...string) *EmbedBuilder {
	if trimmed := strings.TrimSpace(text); trimmed != "" {
		footer := map[string]interface{}{"text": trimmed}
		if len(iconURL) > 0 {
			if icon := strings.TrimSpace(iconURL[0]); icon != "" {
				footer["icon_url"] = icon
			}
		}
		b.footer = footer
	}
	return b
}

// WithTimestamp sets the embed timestamp (UTC RFC3339).
func (b *EmbedBuilder) WithTimestamp(ts time.Time) *EmbedBuilder {
	if ts.IsZero() {
		ts = time.Now()
	}
	b.timestamp = ts.UTC().Format(time.RFC3339)
	return b
}

// WithAuthor sets the embed author block.
func (b *EmbedBuilder) WithAuthor(name, url, iconURL string) *EmbedBuilder {
	name = strings.TrimSpace(name)
	if name == "" {
		return b
	}

	author := map[string]interface{}{"name": name}
	if url = strings.TrimSpace(url); url != "" {
		author["url"] = url
	}
	if iconURL = strings.TrimSpace(iconURL); iconURL != "" {
		author["icon_url"] = iconURL
	}
	b.author = author
	return b
}

// WithThumbnail sets the embed thumbnail URL.
func (b *EmbedBuilder) WithThumbnail(url string) *EmbedBuilder {
	if trimmed := strings.TrimSpace(url); trimmed != "" {
		b.thumbnail = map[string]interface{}{"url": trimmed}
	}
	return b
}

// AddField appends a new field to the embed, skipping empty values.
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" || value == "" {
		return b
	}

	b.fields = append(b.fields, map[string]interface{}{
		"name":   name,
		"value":  value,
		"inline": inline,
	})
	return b
}

// AddInlineField is a convenience wrapper around AddField with inline=true.
func (b *EmbedBuilder) AddInlineField(name, value string) *EmbedBuilder {
	return b.AddField(name, value, true)
}

// Build returns the underlying map expected by Discord/MCP.
func (b *EmbedBuilder) Build() map[string]interface{} {
	embed := map[string]interface{}{}

	if b.title != "" {
		embed["title"] = b.title
	}
	if b.url != "" {
		embed["url"] = b.url
	}
	if b.description != "" {
		embed["description"] = b.description
	}
	if b.color != "" {
		embed["color"] = b.color
	}
	if len(b.fields) > 0 {
		embed["fields"] = b.fields
	}
	if b.footer != nil {
		embed["footer"] = b.footer
	}
	if b.author != nil {
		embed["author"] = b.author
	}
	if b.thumbnail != nil {
		embed["thumbnail"] = b.thumbnail
	}
	if b.timestamp != "" {
		embed["timestamp"] = b.timestamp
	}

	return embed
}
