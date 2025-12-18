package domain

// InteractionType represents the type of Discord interaction.
type InteractionType uint8

const (
	InteractionTypePing InteractionType = iota + 1
	InteractionTypeApplicationCommand
	InteractionTypeComponent
	InteractionTypeAutocomplete
	InteractionTypeModalSubmit
)

// Interaction represents a Discord interaction (unified model with subtypes).
// Based on the interaction type, one of CommandData, ComponentData, or ModalData will be populated.
type Interaction struct {
	ID          InteractionID
	Type        InteractionType
	Token       string
	GuildID     GuildID
	ChannelID   ChannelID
	User        *User
	Member      *Member
	Message     *Message
	AppID       ApplicationID
	Locale      string
	GuildLocale string
	Version     int

	// Subtype data - only one is populated based on Type
	CommandData   *CommandInteractionData
	ComponentData *ComponentInteractionData
	ModalData     *ModalInteractionData
}

// CommandInteractionData contains data for application command interactions.
type CommandInteractionData struct {
	ID       CommandID
	Name     string
	Type     CommandType
	Options  []*CommandOption
	Resolved *ResolvedData
	TargetID string // For user/message context menu commands
}

// CommandOption represents a command option value.
type CommandOption struct {
	Name    string
	Type    CommandOptionType
	Value   any
	Options []*CommandOption // For subcommands/groups
	Focused bool             // For autocomplete
}

// ResolvedData contains resolved objects from command options.
type ResolvedData struct {
	Users    map[UserID]*User
	Members  map[UserID]*Member
	Roles    map[RoleID]*Role
	Channels map[ChannelID]*Channel
	Messages map[MessageID]*Message
}

// ComponentInteractionData contains data for component interactions.
type ComponentInteractionData struct {
	CustomID      string
	ComponentType ComponentType
	Values        []string // For select menus
	Resolved      *ResolvedData
}

// ModalInteractionData contains data for modal submit interactions.
type ModalInteractionData struct {
	CustomID   string
	Components []Component
}

// Response represents an interaction response.
type Response struct {
	Type ResponseType
	Data *ResponseData
}

// ResponseType represents the type of interaction response.
type ResponseType int

const (
	ResponseTypePong ResponseType = iota + 1
	_                             // unused
	_                             // unused
	ResponseTypeChannelMessage
	ResponseTypeDeferredChannelMessage
	ResponseTypeDeferredMessageUpdate
	ResponseTypeUpdateMessage
	ResponseTypeAutocompleteResult
	ResponseTypeModal
	ResponseTypePremiumRequired
)

// ResponseData contains the data for an interaction response.
type ResponseData struct {
	TTS             bool
	Content         string
	Embeds          []*Embed
	Components      []Component
	Files           []*File
	AllowedMentions *AllowedMentions
	Flags           ResponseFlags
	Choices         []*AutocompleteChoice

	// Modal fields
	CustomID string
	Title    string
}

// ResponseFlags represents interaction response flags.
type ResponseFlags int

const (
	ResponseFlagsCrossPosted ResponseFlags = 1 << iota
	ResponseFlagsIsCrossPost
	ResponseFlagsSuppressEmbeds
	ResponseFlagsSourceMessageDeleted
	ResponseFlagsUrgent
	ResponseFlagsHasThread
	ResponseFlagsEphemeral
	ResponseFlagsLoading
	ResponseFlagsFailedToMentionRoles
)

// AutocompleteChoice represents an autocomplete choice.
type AutocompleteChoice struct {
	Name  string
	Value any
}

// ResponseEdit represents an edit to an interaction response.
type ResponseEdit struct {
	Content         *string
	Embeds          *[]*Embed
	Components      *[]Component
	Files           []*File
	AllowedMentions *AllowedMentions
}

// GetUser returns the user who triggered the interaction.
func (i *Interaction) GetUser() *User {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
}

// GetUserID returns the ID of the user who triggered the interaction.
func (i *Interaction) GetUserID() UserID {
	if u := i.GetUser(); u != nil {
		return u.ID
	}
	return ""
}

// IsCommand returns true if this is a command interaction.
func (i *Interaction) IsCommand() bool {
	return i.Type == InteractionTypeApplicationCommand
}

// IsComponent returns true if this is a component interaction.
func (i *Interaction) IsComponent() bool {
	return i.Type == InteractionTypeComponent
}

// IsModal returns true if this is a modal submit interaction.
func (i *Interaction) IsModal() bool {
	return i.Type == InteractionTypeModalSubmit
}

// IsAutocomplete returns true if this is an autocomplete interaction.
func (i *Interaction) IsAutocomplete() bool {
	return i.Type == InteractionTypeAutocomplete
}

// OptionByName returns a command option by name (for command interactions).
func (c *CommandInteractionData) OptionByName(name string) *CommandOption {
	for _, opt := range c.Options {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

// StringValue returns the string value of the option.
func (o *CommandOption) StringValue() string {
	if o == nil || o.Value == nil {
		return ""
	}
	if s, ok := o.Value.(string); ok {
		return s
	}
	return ""
}

// IntValue returns the integer value of the option.
func (o *CommandOption) IntValue() int64 {
	if o == nil || o.Value == nil {
		return 0
	}
	switch v := o.Value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}

// FloatValue returns the float value of the option.
func (o *CommandOption) FloatValue() float64 {
	if o == nil || o.Value == nil {
		return 0
	}
	if f, ok := o.Value.(float64); ok {
		return f
	}
	return 0
}

// BoolValue returns the boolean value of the option.
func (o *CommandOption) BoolValue() bool {
	if o == nil || o.Value == nil {
		return false
	}
	if b, ok := o.Value.(bool); ok {
		return b
	}
	return false
}

// ComponentValue returns the value from a modal text input by custom ID.
func (m *ModalInteractionData) ComponentValue(customID string) string {
	for _, row := range m.Components {
		if ar, ok := row.(*ActionRow); ok {
			for _, comp := range ar.Components {
				if ti, ok := comp.(*TextInput); ok {
					if ti.CustomID == customID {
						return ti.Value
					}
				}
			}
		}
	}
	return ""
}
