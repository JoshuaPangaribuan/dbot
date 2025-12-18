// Package reply provides fluent builders for Discord interaction responses.
package reply

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
)

// Builder builds and sends interaction responses.
type Builder struct {
	rest            port.RestClient
	interactionID   domain.InteractionID
	token           string
	deferred        bool
	content         string
	embeds          []*domain.Embed
	files           []*domain.File
	allowedMentions *domain.AllowedMentions
	ephemeral       bool
	spoiler         bool
	updateMsg       bool
	components      []domain.Component
	onSent          func()
}

// NewBuilder creates a new reply builder.
func NewBuilder(rest port.RestClient, interactionID domain.InteractionID, token string, deferred bool, onSent func()) *Builder {
	return &Builder{
		rest:          rest,
		interactionID: interactionID,
		token:         token,
		deferred:      deferred,
		onSent:        onSent,
	}
}

// Content sets the text content of the reply.
func (b *Builder) Content(s string) *Builder {
	b.content = s
	return b
}

// Embed adds an embed to the reply using the Embed builder.
func (b *Builder) Embed(e *Embed) *Builder {
	if e != nil {
		b.embeds = append(b.embeds, e.Build())
	}
	return b
}

// RawEmbed adds a raw domain.Embed to the reply.
func (b *Builder) RawEmbed(e *domain.Embed) *Builder {
	if e != nil {
		b.embeds = append(b.embeds, e)
	}
	return b
}

// Attachment adds a file attachment to the reply.
func (b *Builder) Attachment(f *domain.File) *Builder {
	if f != nil {
		b.files = append(b.files, f)
	}
	return b
}

// AllowedMentions configures which mentions Discord is allowed to parse.
func (b *Builder) AllowedMentions(am *domain.AllowedMentions) *Builder {
	b.allowedMentions = am
	return b
}

// NoMentions disables mention parsing entirely (prevents pings).
func (b *Builder) NoMentions() *Builder {
	b.allowedMentions = &domain.AllowedMentions{
		Parse: []domain.AllowedMentionType{},
	}
	return b
}

// Ephemeral makes the reply visible only to the user who triggered the interaction.
func (b *Builder) Ephemeral() *Builder {
	b.ephemeral = true
	return b
}

// Components adds message components (buttons, selects) to the reply.
func (b *Builder) Components(c ...domain.Component) *Builder {
	b.components = append(b.components, c...)
	return b
}

// Spoiler wraps the content in spoiler tags.
func (b *Builder) Spoiler() *Builder {
	b.spoiler = true
	return b
}

// UpdateMessage makes this response update the original message (for component interactions).
func (b *Builder) UpdateMessage() *Builder {
	b.updateMsg = true
	return b
}

// Send sends the reply and returns the sent message.
func (b *Builder) Send(ctx context.Context) (*domain.Message, error) {
	flags := domain.ResponseFlags(0)
	if b.ephemeral {
		flags = domain.ResponseFlagsEphemeral
	}

	content := b.content
	if b.spoiler && content != "" {
		content = "||" + content + "||"
	}

	if b.deferred {
		return b.sendDeferred(ctx, content)
	}

	// Determine response type
	responseType := domain.ResponseTypeChannelMessage
	if b.updateMsg {
		responseType = domain.ResponseTypeUpdateMessage
	}

	resp := &domain.Response{
		Type: responseType,
		Data: &domain.ResponseData{
			Content:         content,
			Embeds:          b.embeds,
			Files:           b.files,
			AllowedMentions: b.allowedMentions,
			Flags:           flags,
			Components:      b.components,
		},
	}

	err := b.rest.RespondInteraction(ctx, b.interactionID, b.token, resp)
	if err != nil {
		return nil, err
	}

	if b.onSent != nil {
		b.onSent()
	}

	// Get the sent message
	return b.rest.GetInteractionResponse(ctx, b.token)
}

func (b *Builder) sendDeferred(ctx context.Context, content string) (*domain.Message, error) {
	edit := &domain.ResponseEdit{
		Content:         &content,
		Embeds:          &b.embeds,
		Components:      &b.components,
		Files:           b.files,
		AllowedMentions: b.allowedMentions,
	}

	return b.rest.EditInteractionResponse(ctx, b.token, edit)
}

// Modal sends a modal dialog as the response.
func (b *Builder) Modal(ctx context.Context, m *Modal) error {
	if m == nil {
		return nil
	}

	resp := &domain.Response{
		Type: domain.ResponseTypeModal,
		Data: m.Build(),
	}

	err := b.rest.RespondInteraction(ctx, b.interactionID, b.token, resp)
	if err != nil {
		return err
	}

	if b.onSent != nil {
		b.onSent()
	}
	return nil
}
