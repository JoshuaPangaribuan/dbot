// Package middleware provides middleware for Discord bot event handling in v2.
//
// # Middleware Types
//
// All middleware in this package works with the v2 EventBus/Feature system
// using v2.MiddlewareFunc:
//
//	func(ctx context.Context, e *domain.Event, next func() error) error
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
// Core middleware in the v2 package:
//   - v2.IgnoreBot() - skips bot users
//   - v2.RequireGuild() - requires guild context
//   - v2.Chain() - chains multiple middleware
//
// # Example
//
//	func (f *Feature) Subscriptions() []*v2.Subscription {
//	    return []*v2.Subscription{
//	        {
//	            Type:       domain.EventTypeCommand,
//	            Handler:    f.handleCommand,
//	            Middleware: []v2.MiddlewareFunc{
//	                v2.IgnoreBot(),
//	                middleware.RateLimit(5, 10*time.Second),
//	            },
//	        },
//	    }
//	}
package middleware
