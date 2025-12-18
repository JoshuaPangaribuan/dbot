package v2

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
)

// MiddlewareFunc processes an event before passing to the next handler.
// Call next to continue the chain; return an error to halt execution.
type MiddlewareFunc func(ctx context.Context, e *domain.Event, next func() error) error

// Chain combines multiple middleware functions into one.
func Chain(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(ctx context.Context, e *domain.Event, next func() error) error {
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
	return func(ctx context.Context, e *domain.Event, next func() error) error {
		if user := e.User(); user != nil && user.Bot {
			return nil
		}
		return next()
	}
}

// RequireGuild returns middleware that only processes events from guilds (not DMs).
func RequireGuild() MiddlewareFunc {
	return func(ctx context.Context, e *domain.Event, next func() error) error {
		if e.GuildID().IsEmpty() {
			return nil // Skip DMs
		}
		return next()
	}
}
