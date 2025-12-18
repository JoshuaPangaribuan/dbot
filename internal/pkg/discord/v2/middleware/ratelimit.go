package middleware

import (
	"context"
	"sync"
	"time"

	v2 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
)

// RateLimit returns middleware that limits events per user within a time window.
// Events exceeding the limit are dropped. For interactions, an ephemeral error is sent.
//
// Parameters:
//   - requests: maximum number of requests allowed in the window
//   - window: time window duration
func RateLimit(requests int, window time.Duration, rest port.RestClient) v2.MiddlewareFunc {
	limiter := newRateLimiter(requests, window)

	return func(ctx context.Context, e *domain.Event, next func() error) error {
		userID := e.UserID()
		if userID.IsEmpty() {
			return next() // Can't determine user, let it through
		}

		if !limiter.allow(string(userID)) {
			// Rate limited - send ephemeral error for interactions
			if e.Interaction != nil && rest != nil {
				resp := &domain.Response{
					Type: domain.ResponseTypeChannelMessage,
					Data: &domain.ResponseData{
						Content: "You're doing that too fast. Please slow down.",
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

// RateLimitSimple returns a simpler rate limiter that doesn't send responses.
// Use this when you don't have access to the RestClient.
func RateLimitSimple(requests int, window time.Duration) v2.MiddlewareFunc {
	limiter := newRateLimiter(requests, window)

	return func(ctx context.Context, e *domain.Event, next func() error) error {
		userID := e.UserID()
		if userID.IsEmpty() {
			return next()
		}

		if !limiter.allow(string(userID)) {
			return nil // Silently drop
		}

		return next()
	}
}

// rateLimiter implements a sliding window rate limiter.
type rateLimiter struct {
	mu       sync.Mutex
	requests int
	window   time.Duration
	buckets  map[string]*bucket
}

type bucket struct {
	timestamps []time.Time
}

func newRateLimiter(requests int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: requests,
		window:   window,
		buckets:  make(map[string]*bucket),
	}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	b, ok := r.buckets[key]
	if !ok {
		b = &bucket{}
		r.buckets[key] = b
	}

	// Remove expired timestamps
	valid := b.timestamps[:0]
	for _, ts := range b.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	b.timestamps = valid

	// Check if under limit
	if len(b.timestamps) >= r.requests {
		return false
	}

	// Add current request
	b.timestamps = append(b.timestamps, now)
	return true
}
