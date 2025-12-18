package middleware

import (
	"context"

	v2 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
)

// RequireNSFW returns middleware that only allows events from NSFW channels.
// For interactions, it will reply with an ephemeral error message.
// For messages, it will silently skip.
func RequireNSFW(rest port.RestClient) v2.MiddlewareFunc {
	// Guard against nil rest client
	if rest == nil {
		return func(ctx context.Context, e *domain.Event, next func() error) error {
			return next()
		}
	}

	return func(ctx context.Context, e *domain.Event, next func() error) error {
		channelID := e.ChannelID()
		if channelID.IsEmpty() {
			return next() // Can't determine channel, let it through
		}

		channel, err := rest.GetChannel(ctx, channelID)
		if err != nil {
			return next() // Can't fetch channel, let it through
		}

		if !channel.NSFW {
			// Not an NSFW channel
			if e.Interaction != nil {
				resp := &domain.Response{
					Type: domain.ResponseTypeChannelMessage,
					Data: &domain.ResponseData{
						Content: "This command can only be used in NSFW channels.",
						Flags:   domain.ResponseFlagsEphemeral,
					},
				}
				_ = rest.RespondInteraction(ctx, e.Interaction.ID, e.Interaction.Token, resp)
			}
			return nil
		}

		return next()
	}
}
