package middleware

import (
	"context"
	"regexp"
	"strings"

	v2 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
)

// ContentFilter is a function that checks message content.
// It returns true if the content should be blocked.
type ContentFilter func(content string) bool

// FilterAction defines what happens when content is filtered.
type FilterAction int

const (
	// FilterActionDelete deletes the message silently.
	FilterActionDelete FilterAction = iota
	// FilterActionWarn deletes the message and sends a warning.
	FilterActionWarn
	// FilterActionBlock blocks the interaction with an ephemeral message.
	FilterActionBlock
)

// ContentFilterConfig configures the content filter middleware.
type ContentFilterConfig struct {
	Filters []ContentFilter
	Action  FilterAction
	Message string // Custom message for warn/block actions
}

// ContentFilterMiddleware returns middleware that filters message content.
// It only applies to message events.
func ContentFilterMiddleware(cfg ContentFilterConfig, rest port.RestClient) v2.MiddlewareFunc {
	if cfg.Message == "" {
		cfg.Message = "Your message was removed for violating content rules."
	}

	return func(ctx context.Context, e *domain.Event, next func() error) error {
		// Only filter message events
		if e.Type() != domain.EventTypeMessage || e.Message == nil {
			return next()
		}

		// Check all filters
		for _, filter := range cfg.Filters {
			if filter(e.Message.Content) {
				return handleFilterMatch(ctx, rest, e.Message, cfg)
			}
		}

		return next()
	}
}

func handleFilterMatch(ctx context.Context, rest port.RestClient, msg *domain.Message, cfg ContentFilterConfig) error {
	if rest == nil {
		return nil
	}

	switch cfg.Action {
	case FilterActionDelete:
		_ = rest.DeleteMessage(ctx, msg.ChannelID, msg.ID)
	case FilterActionWarn:
		_ = rest.DeleteMessage(ctx, msg.ChannelID, msg.ID)
		// Send a warning message
		_, _ = rest.SendMessage(ctx, msg.ChannelID, &domain.MessageSend{
			Content: cfg.Message,
		})
	case FilterActionBlock:
		_ = rest.DeleteMessage(ctx, msg.ChannelID, msg.ID)
	}
	return nil
}

// WordFilter returns a ContentFilter that blocks messages containing any of the given words.
// Matching is case-insensitive.
func WordFilter(words ...string) ContentFilter {
	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}
	return func(content string) bool {
		contentLower := strings.ToLower(content)
		for _, word := range lower {
			if strings.Contains(contentLower, word) {
				return true
			}
		}
		return false
	}
}

// RegexFilter returns a ContentFilter that blocks messages matching any of the given patterns.
func RegexFilter(patterns ...*regexp.Regexp) ContentFilter {
	return func(content string) bool {
		for _, p := range patterns {
			if p.MatchString(content) {
				return true
			}
		}
		return false
	}
}

// InviteFilter returns a ContentFilter that blocks Discord invite links.
func InviteFilter() ContentFilter {
	pattern := regexp.MustCompile(`(?i)(discord\.gg|discord\.com/invite|discordapp\.com/invite)/[\w-]+`)
	return func(content string) bool {
		return pattern.MatchString(content)
	}
}

// LinkFilter returns a ContentFilter that blocks all URLs.
func LinkFilter() ContentFilter {
	pattern := regexp.MustCompile(`https?://[^\s]+`)
	return func(content string) bool {
		return pattern.MatchString(content)
	}
}

// MentionSpamFilter returns a ContentFilter that blocks messages with too many mentions.
func MentionSpamFilter(maxMentions int) ContentFilter {
	pattern := regexp.MustCompile(`<@!?\d+>`)
	return func(content string) bool {
		matches := pattern.FindAllString(content, -1)
		return len(matches) > maxMentions
	}
}

// CapsFilter returns a ContentFilter that blocks messages with too many capital letters.
// threshold is the percentage of caps (0.0-1.0) that triggers the filter.
func CapsFilter(threshold float64, minLength int) ContentFilter {
	return func(content string) bool {
		if len(content) < minLength {
			return false
		}
		caps := 0
		total := 0
		for _, r := range content {
			if r >= 'A' && r <= 'Z' {
				caps++
				total++
			} else if r >= 'a' && r <= 'z' {
				total++
			}
		}
		if total == 0 {
			return false
		}
		return float64(caps)/float64(total) > threshold
	}
}
