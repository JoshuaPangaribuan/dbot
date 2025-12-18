// Package discord provides a framework for building Discord bots.
package discord

import (
	"context"
	"sort"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// EventType identifies the kind of Discord event.
type EventType int

const (
	// EventTypeMessage represents a message create event.
	EventTypeMessage EventType = iota
	// EventTypeCommand represents a slash command interaction.
	EventTypeCommand
	// EventTypeComponent represents a button/select menu interaction.
	EventTypeComponent
	// EventTypeModal represents a modal submission.
	EventTypeModal
	// EventTypeVoiceStateUpdate represents a voice state change.
	EventTypeVoiceStateUpdate
	// EventTypeReactionAdd represents a reaction being added.
	EventTypeReactionAdd
	// EventTypeReactionRemove represents a reaction being removed.
	EventTypeReactionRemove
)

// String returns the string representation of the event type.
func (e EventType) String() string {
	switch e {
	case EventTypeMessage:
		return "message"
	case EventTypeCommand:
		return "command"
	case EventTypeComponent:
		return "component"
	case EventTypeModal:
		return "modal"
	case EventTypeVoiceStateUpdate:
		return "voice_state_update"
	case EventTypeReactionAdd:
		return "reaction_add"
	case EventTypeReactionRemove:
		return "reaction_remove"
	default:
		return "unknown"
	}
}

// Event represents a Discord event with its associated data.
type Event struct {
	typ     EventType
	ctx     context.Context
	session *discordgo.Session
	data    interface{}
}

// Type returns the event type.
func (e *Event) Type() EventType { return e.typ }

// Context returns the context associated with this event.
func (e *Event) Context() context.Context { return e.ctx }

// Session returns the discordgo session.
func (e *Event) Session() *discordgo.Session { return e.session }

// Data returns the underlying discordgo event data.
// Use type assertions to access specific fields:
//
//	switch e.Type() {
//	case EventTypeMessage:
//	    msg := e.Data().(*discordgo.MessageCreate)
//	case EventTypeCommand:
//	    ic := e.Data().(*discordgo.InteractionCreate)
//	}
func (e *Event) Data() interface{} { return e.data }

// NewEvent creates a new Event with the given parameters.
func NewEvent(typ EventType, ctx context.Context, s *discordgo.Session, data interface{}) *Event {
	return &Event{
		typ:     typ,
		ctx:     ctx,
		session: s,
		data:    data,
	}
}

// EventHandler processes an event.
type EventHandler func(ctx context.Context, e *Event) error

// Subscription represents a handler subscription to an event type.
type Subscription struct {
	Type       EventType
	Handler    EventHandler
	Middleware []MiddlewareFunc
	Priority   int // Lower values run first
}

// EventBus manages event subscriptions and dispatching.
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[EventType][]*Subscription
	middleware    []MiddlewareFunc
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscriptions: make(map[EventType][]*Subscription),
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
func (eb *EventBus) Publish(e *Event) error {
	if e == nil {
		return nil
	}

	// Copy slices under lock to avoid races with concurrent Subscribe/Use calls
	eb.mu.RLock()
	originalSubs := eb.subscriptions[e.typ]
	subs := make([]*Subscription, len(originalSubs))
	copy(subs, originalSubs)
	globalMW := make([]MiddlewareFunc, len(eb.middleware))
	copy(globalMW, eb.middleware)
	eb.mu.RUnlock()

	var firstErr error
	for _, sub := range subs {
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

		if err := handler(e.ctx, e); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// wrapHandler wraps a handler with middleware.
func wrapHandler(mw MiddlewareFunc, h EventHandler) EventHandler {
	return func(ctx context.Context, e *Event) error {
		return mw(ctx, e, func() error {
			return h(ctx, e)
		})
	}
}

// Subscriptions returns a copy of all subscriptions for an event type.
func (eb *EventBus) Subscriptions(t EventType) []*Subscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	subs := eb.subscriptions[t]
	result := make([]*Subscription, len(subs))
	copy(result, subs)
	return result
}
