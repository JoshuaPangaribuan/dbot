package v2

import (
	"context"
	"sort"
	"sync"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
)

// EventHandler processes an event.
type EventHandler func(ctx context.Context, e *domain.Event) error

// Subscription represents a handler subscription to an event type.
type Subscription struct {
	Type       domain.EventType
	Handler    EventHandler
	Middleware []MiddlewareFunc
	Priority   int // Lower values run first
}

// EventBus manages event subscriptions and dispatching.
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[domain.EventType][]*Subscription
	middleware    []MiddlewareFunc
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscriptions: make(map[domain.EventType][]*Subscription),
	}
}

// Use adds global middleware that runs for all events.
func (eb *EventBus) Use(mw ...MiddlewareFunc) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.middleware = append(eb.middleware, mw...)
}

// Subscribe registers a subscription for an event type.
func (eb *EventBus) Subscribe(sub *Subscription) {
	if sub == nil {
		return
	}
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscriptions[sub.Type] = append(eb.subscriptions[sub.Type], sub)
	// Sort by priority (stable sort preserves registration order for equal priorities)
	sort.SliceStable(eb.subscriptions[sub.Type], func(i, j int) bool {
		return eb.subscriptions[sub.Type][i].Priority < eb.subscriptions[sub.Type][j].Priority
	})
}

// Publish dispatches an event to all registered handlers.
// Handlers are called in priority order. If a handler returns an error,
// subsequent handlers are still called but the first error is returned.
func (eb *EventBus) Publish(e *domain.Event) error {
	if e == nil {
		return nil
	}

	// Copy slices under lock to avoid races with concurrent Subscribe/Use calls
	eb.mu.RLock()
	originalSubs := eb.subscriptions[e.Type()]
	subs := make([]*Subscription, len(originalSubs))
	copy(subs, originalSubs)
	globalMW := make([]MiddlewareFunc, len(eb.middleware))
	copy(globalMW, eb.middleware)
	eb.mu.RUnlock()

	var firstErr error
	for _, sub := range subs {
		// Skip subscriptions with nil handlers
		if sub.Handler == nil {
			continue
		}

		// Build middleware chain: global middleware -> subscription middleware -> handler
		handler := sub.Handler

		// Wrap with subscription-specific middleware (reverse order)
		for i := len(sub.Middleware) - 1; i >= 0; i-- {
			handler = wrapHandler(sub.Middleware[i], handler)
		}

		// Wrap with global middleware (reverse order)
		for i := len(globalMW) - 1; i >= 0; i-- {
			handler = wrapHandler(globalMW[i], handler)
		}

		if err := handler(e.Context(), e); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// wrapHandler wraps a handler with middleware.
func wrapHandler(mw MiddlewareFunc, h EventHandler) EventHandler {
	return func(ctx context.Context, e *domain.Event) error {
		return mw(ctx, e, func() error {
			return h(ctx, e)
		})
	}
}

// Subscriptions returns a copy of all subscriptions for an event type.
func (eb *EventBus) Subscriptions(t domain.EventType) []*Subscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	subs := eb.subscriptions[t]
	result := make([]*Subscription, len(subs))
	copy(result, subs)
	return result
}
