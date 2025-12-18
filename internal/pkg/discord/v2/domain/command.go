package domain

// CommandType represents the type of application command.
type CommandType int

const (
	CommandTypeChatInput CommandType = iota + 1
	CommandTypeUser
	CommandTypeMessage
)

// Command represents a Discord application command.
type Command struct {
	ID                       CommandID
	ApplicationID            ApplicationID
	GuildID                  GuildID // Empty for global commands
	Name                     string
	Description              string
	Type                     CommandType
	Options                  []*CommandOptionDefinition
	DefaultPermissions       *int64
	DMPermission             *bool
	NSFW                     bool
	Version                  string
	NameLocalizations        map[string]string
	DescriptionLocalizations map[string]string
}

// CommandOptionDefinition represents a command option definition.
type CommandOptionDefinition struct {
	Type                     CommandOptionType
	Name                     string
	Description              string
	Required                 bool
	Choices                  []*CommandChoice
	Options                  []*CommandOptionDefinition // For subcommands/groups
	ChannelTypes             []ChannelType
	MinValue                 *float64
	MaxValue                 *float64
	MinLength                *int
	MaxLength                *int
	Autocomplete             bool
	NameLocalizations        map[string]string
	DescriptionLocalizations map[string]string
}

// CommandOptionType represents the type of command option.
type CommandOptionType int

const (
	CommandOptionTypeSubCommand CommandOptionType = iota + 1
	CommandOptionTypeSubCommandGroup
	CommandOptionTypeString
	CommandOptionTypeInteger
	CommandOptionTypeBoolean
	CommandOptionTypeUser
	CommandOptionTypeChannel
	CommandOptionTypeRole
	CommandOptionTypeMentionable
	CommandOptionTypeNumber
	CommandOptionTypeAttachment
)

// CommandChoice represents a command option choice.
type CommandChoice struct {
	Name              string
	Value             any
	NameLocalizations map[string]string
}

// NewCommand creates a new chat input command.
func NewCommand(name, description string) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Type:        CommandTypeChatInput,
	}
}

// NewUserCommand creates a new user context menu command.
func NewUserCommand(name string) *Command {
	return &Command{
		Name: name,
		Type: CommandTypeUser,
	}
}

// NewMessageCommand creates a new message context menu command.
func NewMessageCommand(name string) *Command {
	return &Command{
		Name: name,
		Type: CommandTypeMessage,
	}
}

// AddOption adds an option to the command.
func (c *Command) AddOption(opt *CommandOptionDefinition) *Command {
	c.Options = append(c.Options, opt)
	return c
}

// StringOption creates a string option.
func StringOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeString,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// IntOption creates an integer option.
func IntOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeInteger,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// BoolOption creates a boolean option.
func BoolOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeBoolean,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// UserOption creates a user option.
func UserOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeUser,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// ChannelOption creates a channel option.
func ChannelOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeChannel,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// RoleOption creates a role option.
func RoleOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeRole,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// NumberOption creates a number (float) option.
func NumberOption(name, description string, required bool) *CommandOptionDefinition {
	return &CommandOptionDefinition{
		Type:        CommandOptionTypeNumber,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// WithChoices adds choices to an option.
func (o *CommandOptionDefinition) WithChoices(choices ...*CommandChoice) *CommandOptionDefinition {
	o.Choices = append(o.Choices, choices...)
	return o
}

// WithAutocomplete enables autocomplete for the option.
func (o *CommandOptionDefinition) WithAutocomplete() *CommandOptionDefinition {
	o.Autocomplete = true
	return o
}
