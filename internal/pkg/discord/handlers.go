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

// StringOption creates a required or optional string option.
func StringOption(name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = discordgo.ApplicationCommandOptionString
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}

// IntOption creates a required or optional integer option.
func IntOption(name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = discordgo.ApplicationCommandOptionInteger
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}

// UserOption creates a required or optional user option.
func UserOption(name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = discordgo.ApplicationCommandOptionUser
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}

// BoolOption creates a required or optional boolean option.
func BoolOption(name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = discordgo.ApplicationCommandOptionBoolean
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}

// ChannelOption creates a required or optional channel option.
func ChannelOption(name, description string, required bool) CommandOption {
	return func(opt *discordgo.ApplicationCommandOption) {
		opt.Type = discordgo.ApplicationCommandOptionChannel
		opt.Name = name
		opt.Description = description
		opt.Required = required
	}
}
