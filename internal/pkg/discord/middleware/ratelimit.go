package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/reply"
	"github.com/bwmarrin/discordgo"
)

// RateLimit returns middleware that limits events per user within a time window.
// Events exceeding the limit are dropped. For interactions, an ephemeral error is sent.
//
// Parameters:
//   - requests: maximum number of requests allowed in the window
//   - window: time window duration
func RateLimit(requests int, window time.Duration) discord.MiddlewareFunc {
	limiter := newRateLimiter(requests, window)

	return func(ctx context.Context, e *discord.Event, next func() error) error {
		userID := discord.ExtractUserID(e)
		if userID == "" {
			return next() // Can't determine user, let it through
		}

		if !limiter.allow(userID) {
			// Rate limited
			if ic, ok := e.Data().(*discordgo.InteractionCreate); ok {
				_ = reply.SendEphemeralError(e.Session(), ic.Interaction, "You're doing that too fast. Please slow down.")
			}
			return nil
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
