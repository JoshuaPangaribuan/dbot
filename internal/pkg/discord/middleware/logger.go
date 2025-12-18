package middleware

import (
	"context"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// EventLogger returns middleware that logs events for the EventBus system.
func EventLogger(log logger.Logger) discord.MiddlewareFunc {
	return func(ctx context.Context, e *discord.Event, next func() error) error {
		start := time.Now()

		// Extract common fields based on event type
		fields := logger.Fields{
			"event_type": e.Type().String(),
		}

		switch data := e.Data().(type) {
		case *discordgo.MessageCreate:
			if data.Author != nil {
				fields["user"] = data.Author.Username
				fields["user_id"] = data.Author.ID
			}
			fields["channel_id"] = data.ChannelID
			fields["guild_id"] = data.GuildID
		case *discordgo.InteractionCreate:
			if data.Member != nil && data.Member.User != nil {
				fields["user"] = data.Member.User.Username
				fields["user_id"] = data.Member.User.ID
			} else if data.User != nil {
				fields["user"] = data.User.Username
				fields["user_id"] = data.User.ID
			}
			fields["channel_id"] = data.ChannelID
			fields["guild_id"] = data.GuildID
			if data.Type == discordgo.InteractionApplicationCommand {
				fields["command"] = data.ApplicationCommandData().Name
			}
		}

		log.Debug(ctx, "event received", fields)

		err := next()

		duration := time.Since(start)
		fields["duration"] = duration.String()

		if err != nil {
			fields["error"] = err.Error()
			log.Error(ctx, "event handler failed", fields)
		} else {
			log.Debug(ctx, "event handled", fields)
		}

		return err
	}
}
