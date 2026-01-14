package middleware

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/reply"
	"github.com/bwmarrin/discordgo"
)

// RequireNSFW returns middleware that only allows events from NSFW channels.
// For interactions, it will reply with an ephemeral error message.
// For messages, it will silently skip.
func RequireNSFW() discord.MiddlewareFunc {
	return func(ctx context.Context, e *discord.Event, next func() error) error {
		channelID := discord.ExtractChannelID(e)
		if channelID == "" {
			return next() // Can't determine channel, let it through
		}

		channel, err := e.Session().Channel(channelID)
		if err != nil {
			return next() // Can't fetch channel, let it through
		}

		if !channel.NSFW {
			// Not an NSFW channel
			if ic, ok := e.Data().(*discordgo.InteractionCreate); ok {
				_ = reply.SendEphemeralError(e.Session(), ic.Interaction, "This command can only be used in NSFW channels.")
			}
			return nil
		}

		return next()
	}
}

