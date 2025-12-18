package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// MiddlewareFunc processes an event before passing to the next handler.
// Call next to continue the chain; return an error to halt execution.
type MiddlewareFunc func(ctx context.Context, e *Event, next func() error) error

// Chain combines multiple middleware functions into one.
func Chain(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(ctx context.Context, e *Event, next func() error) error {
		// Build chain from end to start
		handler := next
		for i := len(middlewares) - 1; i >= 0; i-- {
			mw := middlewares[i]
			h := handler
			handler = func() error {
				return mw(ctx, e, h)
			}
		}
		return handler()
	}
}

// IgnoreBot returns middleware that skips events from bot users.
func IgnoreBot() MiddlewareFunc {
	return func(ctx context.Context, e *Event, next func() error) error {
		var user *discordgo.User
		switch data := e.Data().(type) {
		case *discordgo.MessageCreate:
			user = data.Author
		case *discordgo.InteractionCreate:
			if data.Member != nil {
				user = data.Member.User
			} else {
				user = data.User
			}
		}
		if user != nil && user.Bot {
			return nil
		}
		return next()
	}
}

// RequireGuild returns middleware that only processes events from guilds (not DMs).
func RequireGuild() MiddlewareFunc {
	return func(ctx context.Context, e *Event, next func() error) error {
		var guildID string
		switch data := e.Data().(type) {
		case *discordgo.MessageCreate:
			guildID = data.GuildID
		case *discordgo.InteractionCreate:
			guildID = data.GuildID
		case *discordgo.VoiceStateUpdate:
			guildID = data.GuildID
		case *discordgo.MessageReactionAdd:
			guildID = data.GuildID
		case *discordgo.MessageReactionRemove:
			guildID = data.GuildID
		}
		if guildID == "" {
			return nil // Skip DMs
		}
		return next()
	}
}
