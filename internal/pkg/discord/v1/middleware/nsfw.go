package middleware

import (
	"context"

	v1 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v1"
	"github.com/bwmarrin/discordgo"
)

// RequireNSFW returns middleware that only allows events from NSFW channels.
// For interactions, it will reply with an ephemeral error message.
// For messages, it will silently skip.
func RequireNSFW() v1.MiddlewareFunc {
	return func(ctx context.Context, e *v1.Event, next func() error) error {
		channelID := getChannelID(e)
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
				// For interactions, send ephemeral error
				_ = e.Session().InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This command can only be used in NSFW channels.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return nil
		}

		return next()
	}
}

func getChannelID(e *v1.Event) string {
	switch data := e.Data().(type) {
	case *discordgo.MessageCreate:
		return data.ChannelID
	case *discordgo.InteractionCreate:
		return data.ChannelID
	case *discordgo.MessageReactionAdd:
		return data.ChannelID
	case *discordgo.MessageReactionRemove:
		return data.ChannelID
	}
	return ""
}

