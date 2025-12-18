package domain

// Component represents a message component.
// This is an interface to allow different component types.
type Component interface {
	componentType() ComponentType
}

// ComponentType represents the type of message component.
type ComponentType int

const (
	ComponentTypeActionRow ComponentType = iota + 1
	ComponentTypeButton
	ComponentTypeStringSelect
	ComponentTypeTextInput
	ComponentTypeUserSelect
	ComponentTypeRoleSelect
	ComponentTypeMentionableSelect
	ComponentTypeChannelSelect
)

// ActionRow represents a row of components.
type ActionRow struct {
	Components []Component
}

func (ActionRow) componentType() ComponentType { return ComponentTypeActionRow }

// Button represents a button component.
type Button struct {
	Style    ButtonStyle
	Label    string
	Emoji    *Emoji
	CustomID string
	URL      string
	Disabled bool
}

func (Button) componentType() ComponentType { return ComponentTypeButton }

// ButtonStyle represents a button style.
type ButtonStyle int

const (
	ButtonStylePrimary ButtonStyle = iota + 1
	ButtonStyleSecondary
	ButtonStyleSuccess
	ButtonStyleDanger
	ButtonStyleLink
)

// SelectMenu represents a select menu component.
type SelectMenu struct {
	Type         ComponentType
	CustomID     string
	Placeholder  string
	MinValues    *int
	MaxValues    int
	Options      []*SelectOption
	Disabled     bool
	ChannelTypes []ChannelType // For channel select
}

func (s SelectMenu) componentType() ComponentType { return s.Type }

// SelectOption represents a select menu option.
type SelectOption struct {
	Label       string
	Value       string
	Description string
	Emoji       *Emoji
	Default     bool
}

// TextInput represents a text input component (for modals).
type TextInput struct {
	CustomID    string
	Style       TextInputStyle
	Label       string
	MinLength   int
	MaxLength   int
	Required    bool
	Value       string
	Placeholder string
}

func (TextInput) componentType() ComponentType { return ComponentTypeTextInput }

// TextInputStyle represents a text input style.
type TextInputStyle int

const (
	TextInputStyleShort TextInputStyle = iota + 1
	TextInputStyleParagraph
)

// NewActionRow creates a new action row with the given components.
func NewActionRow(components ...Component) *ActionRow {
	return &ActionRow{Components: components}
}

// NewButton creates a new button.
func NewButton(style ButtonStyle, label, customID string) *Button {
	return &Button{
		Style:    style,
		Label:    label,
		CustomID: customID,
	}
}

// NewLinkButton creates a new link button.
func NewLinkButton(label, url string) *Button {
	return &Button{
		Style: ButtonStyleLink,
		Label: label,
		URL:   url,
	}
}

// WithEmoji adds an emoji to the button.
func (b *Button) WithEmoji(emoji *Emoji) *Button {
	b.Emoji = emoji
	return b
}

// WithDisabled sets the disabled state.
func (b *Button) WithDisabled(disabled bool) *Button {
	b.Disabled = disabled
	return b
}

// NewStringSelectMenu creates a new string select menu.
func NewStringSelectMenu(customID string, options ...*SelectOption) *SelectMenu {
	return &SelectMenu{
		Type:     ComponentTypeStringSelect,
		CustomID: customID,
		Options:  options,
	}
}

// NewSelectOption creates a new select option.
func NewSelectOption(label, value string) *SelectOption {
	return &SelectOption{
		Label: label,
		Value: value,
	}
}

// WithDescription adds a description to the option.
func (o *SelectOption) WithDescription(desc string) *SelectOption {
	o.Description = desc
	return o
}

// WithEmoji adds an emoji to the option.
func (o *SelectOption) WithEmoji(emoji *Emoji) *SelectOption {
	o.Emoji = emoji
	return o
}

// WithDefault sets the option as default.
func (o *SelectOption) WithDefault() *SelectOption {
	o.Default = true
	return o
}

// NewTextInput creates a new text input for modals.
func NewTextInput(customID, label string, style TextInputStyle) *TextInput {
	return &TextInput{
		CustomID: customID,
		Label:    label,
		Style:    style,
	}
}

// WithPlaceholder sets the placeholder text.
func (t *TextInput) WithPlaceholder(placeholder string) *TextInput {
	t.Placeholder = placeholder
	return t
}

// WithValue sets the pre-filled value.
func (t *TextInput) WithValue(value string) *TextInput {
	t.Value = value
	return t
}

// WithMinLength sets the minimum length.
func (t *TextInput) WithMinLength(min int) *TextInput {
	t.MinLength = min
	return t
}

// WithMaxLength sets the maximum length.
func (t *TextInput) WithMaxLength(max int) *TextInput {
	t.MaxLength = max
	return t
}

// WithRequired sets whether the input is required.
func (t *TextInput) WithRequired(required bool) *TextInput {
	t.Required = required
	return t
}
