// Package middleware provides middleware for Discord bot event handling.
//
// # Middleware Types
//
// All middleware in this package works with the EventBus/Feature system
// using v1.MiddlewareFunc:
//
//	func(ctx context.Context, e *v1.Event, next func() error) error
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
// Core middleware in the v1 package:
//   - v1.IgnoreBot() - skips bot users
//   - v1.RequireGuild() - requires guild context
//   - v1.Chain() - chains multiple middleware
//
// # Example
//
//	func (f *Feature) Subscriptions() []*v1.Subscription {
//	    return []*v1.Subscription{
//	        {
//	            Type:       v1.EventTypeCommand,
//	            Handler:    f.handleCommand,
//	            Middleware: []v1.MiddlewareFunc{
//	                v1.IgnoreBot(),
//	                middleware.RateLimit(5, 10*time.Second),
//	            },
//	        },
//	    }
//	}
package middleware

