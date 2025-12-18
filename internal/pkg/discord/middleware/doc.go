// Package middleware provides middleware for Discord bot event handling.
//
// # Middleware Types
//
// All middleware in this package works with the EventBus/Feature system
// using discord.MiddlewareFunc:
//
//	func(ctx context.Context, e *discord.Event, next func() error) error
//
// # Available Middleware
//
// This package provides:
//   - RateLimit() - rate limits events per user
//   - RequireNSFW() - requires NSFW channel
//   - ContentFilterMiddleware() - filters message content
//   - EventLogger() - logs event handling
//   - RequireEventPermission() - checks permissions
//   - RequireAnyEventPermission() - checks any of multiple permissions
//
// Core middleware in the discord package:
//   - discord.IgnoreBot() - skips bot users
//   - discord.RequireGuild() - requires guild context
//   - discord.Chain() - chains multiple middleware
//
// # Example
//
//	func (f *Feature) Subscriptions() []*discord.Subscription {
//	    return []*discord.Subscription{
//	        {
//	            Type:       discord.EventTypeCommand,
//	            Handler:    f.handleCommand,
//	            Middleware: []discord.MiddlewareFunc{
//	                discord.IgnoreBot(),
//	                middleware.RateLimit(5, 10*time.Second),
//	            },
//	        },
//	    }
//	}
package middleware
