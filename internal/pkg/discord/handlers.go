package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// CommandHandler handles slash commands with a typed context.
type CommandHandler func(ctx context.Context, cmd *CommandContext) error

// MessageHandler handles messages with a typed context.
type MessageHandler func(ctx context.Context, msg *MessageContext) error

// ReactionHandler handles reactions with a typed context.
type ReactionHandler func(ctx context.Context, r *ReactionContext) error

// CommandOption configures a slash command option.
type CommandOption func(*discordgo.ApplicationCommandOption)

// option creates a command option with the specified type, name, description, and required flag.
func option(optType discordgo.ApplicationCommandOptionType, name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = optType
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}

// StringOption creates a required or optional string option.
func StringOption(name, description string, required bool) CommandOption {
	return option(discordgo.ApplicationCommandOptionString, name, description, required)
}

// IntOption creates a required or optional integer option.
func IntOption(name, description string, required bool) CommandOption {
	return option(discordgo.ApplicationCommandOptionInteger, name, description, required)
}

// UserOption creates a required or optional user option.
func UserOption(name, description string, required bool) CommandOption {
	return option(discordgo.ApplicationCommandOptionUser, name, description, required)
}

// BoolOption creates a required or optional boolean option.
func BoolOption(name, description string, required bool) CommandOption {
	return option(discordgo.ApplicationCommandOptionBoolean, name, description, required)
}

// ChannelOption creates a required or optional channel option.
func ChannelOption(name, description string, required bool) CommandOption {
	return option(discordgo.ApplicationCommandOptionChannel, name, description, required)
}
