package reply

import "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"

// Modal is a fluent builder for Discord modal dialogs.
type Modal struct {
	customID   string
	title      string
	components []domain.Component
}

// NewModal creates a new modal builder.
func NewModal(customID, title string) *Modal {
	return &Modal{
		customID: customID,
		title:    title,
	}
}

// TextInput adds a text input component to the modal.
func (m *Modal) TextInput(customID, label string, style domain.TextInputStyle, opts ...TextInputOption) *Modal {
	ti := domain.NewTextInput(customID, label, style)
	for _, opt := range opts {
		opt(ti)
	}
	// Text inputs must be wrapped in action rows
	m.components = append(m.components, domain.NewActionRow(ti))
	return m
}

// ShortTextInput is a convenience method for adding a single-line text input.
func (m *Modal) ShortTextInput(customID, label string, opts ...TextInputOption) *Modal {
	return m.TextInput(customID, label, domain.TextInputStyleShort, opts...)
}

// ParagraphTextInput is a convenience method for adding a multi-line text input.
func (m *Modal) ParagraphTextInput(customID, label string, opts ...TextInputOption) *Modal {
	return m.TextInput(customID, label, domain.TextInputStyleParagraph, opts...)
}

// Build returns the response data for the modal.
func (m *Modal) Build() *domain.ResponseData {
	return &domain.ResponseData{
		CustomID:   m.customID,
		Title:      m.title,
		Components: m.components,
	}
}

// TextInputOption configures a text input component.
type TextInputOption func(*domain.TextInput)

// WithPlaceholder sets the placeholder text for a text input.
func WithPlaceholder(placeholder string) TextInputOption {
	return func(ti *domain.TextInput) {
		ti.Placeholder = placeholder
	}
}

// WithValue sets the pre-filled value for a text input.
func WithValue(value string) TextInputOption {
	return func(ti *domain.TextInput) {
		ti.Value = value
	}
}

// WithMinLength sets the minimum length for a text input.
func WithMinLength(min int) TextInputOption {
	return func(ti *domain.TextInput) {
		ti.MinLength = min
	}
}

// WithMaxLength sets the maximum length for a text input.
func WithMaxLength(max int) TextInputOption {
	return func(ti *domain.TextInput) {
		ti.MaxLength = max
	}
}

// WithRequired sets whether the text input is required.
func WithRequired(required bool) TextInputOption {
	return func(ti *domain.TextInput) {
		ti.Required = required
	}
}
