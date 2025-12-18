package middleware

import (
	"context"
	"time"

	v2 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// EventLogger returns middleware that logs events for the EventBus system.
func EventLogger(log logger.Logger) v2.MiddlewareFunc {
	return func(ctx context.Context, e *domain.Event, next func() error) error {
		start := time.Now()

		// Extract common fields based on event type
		fields := logger.Fields{
			"event_type": e.Type().String(),
		}

		// Add user info
		if user := e.User(); user != nil {
			fields["user"] = user.Username
			fields["user_id"] = user.ID
		}

		// Add location info
		if !e.ChannelID().IsEmpty() {
			fields["channel_id"] = e.ChannelID()
		}
		if !e.GuildID().IsEmpty() {
			fields["guild_id"] = e.GuildID()
		}

		// Add command name for command events
		if e.Interaction != nil && e.Interaction.CommandData != nil {
			fields["command"] = e.Interaction.CommandData.Name
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
