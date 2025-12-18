package port

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
)

// RestClient defines the interface for Discord REST API operations.
// Adapters implement this to make REST API calls to Discord.
type RestClient interface {
	// Commands

	// BulkOverwriteCommands overwrites all commands for the application.
	// If guildID is empty, global commands are overwritten.
	BulkOverwriteCommands(ctx context.Context, guildID domain.GuildID, cmds []*domain.Command) ([]*domain.Command, error)

	// DeleteCommand deletes a command.
	// If guildID is empty, deletes a global command.
	DeleteCommand(ctx context.Context, guildID domain.GuildID, cmdID domain.CommandID) error

	// Interactions

	// RespondInteraction sends an initial response to an interaction.
	RespondInteraction(ctx context.Context, interactionID domain.InteractionID, token string, resp *domain.Response) error

	// DeferInteraction sends a deferred response to an interaction.
	DeferInteraction(ctx context.Context, interactionID domain.InteractionID, token string, flags domain.ResponseFlags) error

	// EditInteractionResponse edits the initial response to an interaction.
	EditInteractionResponse(ctx context.Context, token string, edit *domain.ResponseEdit) (*domain.Message, error)

	// GetInteractionResponse gets the initial response to an interaction.
	GetInteractionResponse(ctx context.Context, token string) (*domain.Message, error)

	// DeleteInteractionResponse deletes the initial response to an interaction.
	DeleteInteractionResponse(ctx context.Context, token string) error

	// FollowupMessage sends a followup message to an interaction.
	FollowupMessage(ctx context.Context, token string, msg *domain.ResponseData) (*domain.Message, error)

	// Messages

	// SendMessage sends a message to a channel.
	SendMessage(ctx context.Context, channelID domain.ChannelID, msg *domain.MessageSend) (*domain.Message, error)

	// EditMessage edits a message.
	EditMessage(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, edit *domain.MessageEdit) (*domain.Message, error)

	// DeleteMessage deletes a message.
	DeleteMessage(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID) error

	// Reactions

	// AddReaction adds a reaction to a message.
	AddReaction(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, emoji string) error

	// RemoveReaction removes the bot's reaction from a message.
	RemoveReaction(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, emoji string) error

	// Lookups

	// GetChannel gets a channel by ID.
	GetChannel(ctx context.Context, channelID domain.ChannelID) (*domain.Channel, error)

	// GetGuildRoles gets all roles in a guild.
	GetGuildRoles(ctx context.Context, guildID domain.GuildID) ([]*domain.Role, error)

	// GetUser gets a user by ID.
	GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error)

	// GetMember gets a guild member.
	GetMember(ctx context.Context, guildID domain.GuildID, userID domain.UserID) (*domain.Member, error)
}
