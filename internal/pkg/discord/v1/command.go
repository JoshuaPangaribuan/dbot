package v1

import "github.com/bwmarrin/discordgo"

// CommandType defines supported command variants.
type CommandType int

const (
	// CommandTypeSlash represents a slash command (e.g., /ping).
	CommandTypeSlash CommandType = iota
	// CommandTypeMessage represents a message context menu command.
	CommandTypeMessage
	// CommandTypeUser represents a user context menu command.
	CommandTypeUser
)

// Command represents a Discord application command.
// Commands are registered with Discord and handled via EventBus subscriptions.
type Command struct {
	Name        string
	Description string
	Type        CommandType
	Options     []*discordgo.ApplicationCommandOption
}

// NewSlashCommand creates a new slash command with the given name and description.
// Use EventBus subscriptions with CommandFilter to handle the command.
func NewSlashCommand(name, description string) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Type:        CommandTypeSlash,
	}
}

// WithOptions adds command options and returns the command for chaining.
func (c *Command) WithOptions(opts ...*discordgo.ApplicationCommandOption) *Command {
	c.Options = append(c.Options, opts...)
	return c
}

// toApplicationCommand converts Command to discordgo.ApplicationCommand.
func (c *Command) toApplicationCommand() *discordgo.ApplicationCommand {
	cmdType := discordgo.ChatApplicationCommand
	switch c.Type {
	case CommandTypeMessage:
		cmdType = discordgo.MessageApplicationCommand
	case CommandTypeUser:
		cmdType = discordgo.UserApplicationCommand
	}

	return &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
		Type:        cmdType,
		Options:     c.Options,
	}
}

