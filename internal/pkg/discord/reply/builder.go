package reply

import (
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
)

// Builder builds and sends interaction responses.
type Builder struct {
	session         *discordgo.Session
	interaction     *discordgo.Interaction
	deferred        bool
	content         string
	embeds          []*discordgo.MessageEmbed
	files           []*discordgo.File
	allowedMentions *discordgo.MessageAllowedMentions
	ephemeral       bool
	spoiler         bool
	updateMsg       bool // for component interactions: update the original message
	components      []discordgo.MessageComponent
	onSent          func() // callback when response sent
}

// NewBuilder creates a new reply builder.
func NewBuilder(s *discordgo.Session, i *discordgo.Interaction, deferred bool, onSent func()) *Builder {
	return &Builder{
		session:     s,
		interaction: i,
		deferred:    deferred,
		onSent:      onSent,
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

// RawEmbed adds a raw discordgo.MessageEmbed to the reply.
func (b *Builder) RawEmbed(e *discordgo.MessageEmbed) *Builder {
	if e != nil {
		b.embeds = append(b.embeds, e)
	}
	return b
}

// Attachment adds a file attachment to the reply.
func (b *Builder) Attachment(name string, r io.Reader) *Builder {
	b.files = append(b.files, &discordgo.File{
		Name:   name,
		Reader: r,
	})
	return b
}

// AllowedMentions configures which mentions Discord is allowed to parse in this reply.
func (b *Builder) AllowedMentions(am *discordgo.MessageAllowedMentions) *Builder {
	b.allowedMentions = am
	return b
}

// NoMentions disables mention parsing entirely for this reply (prevents pings).
func (b *Builder) NoMentions() *Builder {
	b.allowedMentions = &discordgo.MessageAllowedMentions{
		Parse: []discordgo.AllowedMentionType{},
	}
	return b
}

// Ephemeral makes the reply visible only to the user who triggered the interaction.
func (b *Builder) Ephemeral() *Builder {
	b.ephemeral = true
	return b
}

// Components adds message components (buttons, selects) to the reply.
func (b *Builder) Components(c ...discordgo.MessageComponent) *Builder {
	b.components = append(b.components, c...)
	return b
}

// Spoiler wraps the content in spoiler tags.
// This is useful for NSFW content in non-NSFW channels.
func (b *Builder) Spoiler() *Builder {
	b.spoiler = true
	return b
}

// UpdateMessage makes this response update the original message (for component interactions).
// This is used when responding to button clicks or select menu selections
// to modify the message containing the component.
func (b *Builder) UpdateMessage() *Builder {
	b.updateMsg = true
	return b
}

// Send sends the reply and returns a SentMessage for further operations.
func (b *Builder) Send() (*SentMessage, error) {
	flags := discordgo.MessageFlags(0)
	if b.ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	content := b.content
	if b.spoiler && content != "" {
		content = "||" + content + "||"
	}

	if b.deferred {
		return b.sendDeferred(content)
	}

	// Determine response type
	responseType := discordgo.InteractionResponseChannelMessageWithSource
	if b.updateMsg {
		responseType = discordgo.InteractionResponseUpdateMessage
	}

	err := b.session.InteractionRespond(b.interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Content:         content,
			Embeds:          b.embeds,
			Files:           b.files,
			AllowedMentions: b.allowedMentions,
			Flags:           flags,
			Components:      b.components,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("respond to interaction: %w", err)
	}
	if b.onSent != nil {
		b.onSent()
	}

	msg, err := b.session.InteractionResponse(b.interaction)
	if err != nil {
		return nil, fmt.Errorf("get interaction response: %w", err)
	}

	return NewSentMessage(b.session, b.interaction, msg), nil
}

func (b *Builder) sendDeferred(content string) (*SentMessage, error) {
	_, err := b.session.InteractionResponseEdit(b.interaction, &discordgo.WebhookEdit{
		Content:         &content,
		Embeds:          &b.embeds,
		Files:           b.files,
		Components:      &b.components,
		AllowedMentions: b.allowedMentions,
	})
	if err != nil {
		return nil, fmt.Errorf("edit deferred response: %w", err)
	}

	msg, err := b.session.InteractionResponse(b.interaction)
	if err != nil {
		return nil, fmt.Errorf("get interaction response: %w", err)
	}

	return NewSentMessage(b.session, b.interaction, msg), nil
}

// Modal sends a modal dialog as the response.
// This should only be used as the initial response to an interaction.
func (b *Builder) Modal(m *Modal) error {
	if m == nil {
		return fmt.Errorf("modal is nil")
	}
	err := b.session.InteractionRespond(b.interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: m.Build(),
	})
	if err != nil {
		return fmt.Errorf("respond with modal: %w", err)
	}
	if b.onSent != nil {
		b.onSent()
	}
	return nil
}
