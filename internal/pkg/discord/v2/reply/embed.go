package reply

import "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"

// Embed is a fluent builder for Discord message embeds.
type Embed struct {
	embed *domain.Embed
}

// NewEmbed creates a new embed builder.
func NewEmbed() *Embed {
	return &Embed{
		embed: &domain.Embed{},
	}
}

// Title sets the embed title.
func (e *Embed) Title(title string) *Embed {
	e.embed.Title = title
	return e
}

// Description sets the embed description.
func (e *Embed) Description(desc string) *Embed {
	e.embed.Description = desc
	return e
}

// URL sets the embed title URL.
func (e *Embed) URL(url string) *Embed {
	e.embed.URL = url
	return e
}

// Color sets the embed color (as integer).
func (e *Embed) Color(color int) *Embed {
	e.embed.Color = color
	return e
}

// Field adds a field to the embed.
func (e *Embed) Field(name, value string, inline bool) *Embed {
	e.embed.Fields = append(e.embed.Fields, &domain.EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

// Thumbnail sets the embed thumbnail.
func (e *Embed) Thumbnail(url string) *Embed {
	e.embed.Thumbnail = &domain.EmbedThumbnail{
		URL: url,
	}
	return e
}

// Image sets the embed image.
func (e *Embed) Image(url string) *Embed {
	e.embed.Image = &domain.EmbedImage{
		URL: url,
	}
	return e
}

// Author sets the embed author.
func (e *Embed) Author(name, url, iconURL string) *Embed {
	e.embed.Author = &domain.EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return e
}

// Footer sets the embed footer.
func (e *Embed) Footer(text, iconURL string) *Embed {
	e.embed.Footer = &domain.EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return e
}

// Timestamp sets the embed timestamp (ISO8601 format).
func (e *Embed) Timestamp(ts string) *Embed {
	e.embed.Timestamp = ts
	return e
}

// Build returns the underlying domain.Embed.
func (e *Embed) Build() *domain.Embed {
	return e.embed
}
