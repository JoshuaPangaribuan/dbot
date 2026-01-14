package reply

import "github.com/bwmarrin/discordgo"

// Modal is a fluent builder for Discord modal dialogs.
type Modal struct {
	customID   string
	title      string
	components []discordgo.MessageComponent
}

// NewModal creates a new modal builder.
func NewModal(customID, title string) *Modal {
	return &Modal{
		customID: customID,
		title:    title,
	}
}

// TextInput adds a text input component to the modal.
func (m *Modal) TextInput(customID, label string, style discordgo.TextInputStyle, opts ...TextInputOption) *Modal {
	ti := &discordgo.TextInput{
		CustomID: customID,
		Label:    label,
		Style:    style,
	}
	for _, opt := range opts {
		opt(ti)
	}
	// Text inputs must be wrapped in action rows
	m.components = append(m.components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{ti},
	})
	return m
}

// ShortTextInput is a convenience method for adding a single-line text input.
func (m *Modal) ShortTextInput(customID, label string, opts ...TextInputOption) *Modal {
	return m.TextInput(customID, label, discordgo.TextInputShort, opts...)
}

// ParagraphTextInput is a convenience method for adding a multi-line text input.
func (m *Modal) ParagraphTextInput(customID, label string, opts ...TextInputOption) *Modal {
	return m.TextInput(customID, label, discordgo.TextInputParagraph, opts...)
}

// Build returns the interaction response data for the modal.
func (m *Modal) Build() *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		CustomID:   m.customID,
		Title:      m.title,
		Components: m.components,
	}
}

// TextInputOption configures a text input component.
type TextInputOption func(*discordgo.TextInput)

// WithPlaceholder sets the placeholder text for a text input.
func WithPlaceholder(placeholder string) TextInputOption {
	return func(ti *discordgo.TextInput) {
		ti.Placeholder = placeholder
	}
}

// WithValue sets the pre-filled value for a text input.
func WithValue(value string) TextInputOption {
	return func(ti *discordgo.TextInput) {
		ti.Value = value
	}
}

// WithMinLength sets the minimum length for a text input.
func WithMinLength(min int) TextInputOption {
	return func(ti *discordgo.TextInput) {
		ti.MinLength = min
	}
}

// WithMaxLength sets the maximum length for a text input.
func WithMaxLength(max int) TextInputOption {
	return func(ti *discordgo.TextInput) {
		ti.MaxLength = max
	}
}

// WithRequired sets whether the text input is required.
func WithRequired(required bool) TextInputOption {
	return func(ti *discordgo.TextInput) {
		ti.Required = required
	}
}
