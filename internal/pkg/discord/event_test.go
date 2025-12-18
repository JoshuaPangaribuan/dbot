package discord

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      string
	}{
		{
			name:      "message event",
			eventType: EventTypeMessage,
			want:      "message",
		},
		{
			name:      "command event",
			eventType: EventTypeCommand,
			want:      "command",
		},
		{
			name:      "component event",
			eventType: EventTypeComponent,
			want:      "component",
		},
		{
			name:      "modal event",
			eventType: EventTypeModal,
			want:      "modal",
		},
		{
			name:      "voice state update event",
			eventType: EventTypeVoiceStateUpdate,
			want:      "voice_state_update",
		},
		{
			name:      "reaction add event",
			eventType: EventTypeReactionAdd,
			want:      "reaction_add",
		},
		{
			name:      "reaction remove event",
			eventType: EventTypeReactionRemove,
			want:      "reaction_remove",
		},
		{
			name:      "unknown event type",
			eventType: EventType(999),
			want:      "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.eventType.String())
		})
	}
}

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		ctx       context.Context
		data      interface{}
	}{
		{
			name:      "creates message event",
			eventType: EventTypeMessage,
			ctx:       context.Background(),
			data:      "test-data",
		},
		{
			name:      "creates command event with nil data",
			eventType: EventTypeCommand,
			ctx:       context.Background(),
			data:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(tt.eventType, tt.ctx, nil, tt.data)

			assert.Equal(t, tt.eventType, event.Type())
			assert.Equal(t, tt.ctx, event.Context())
			assert.Equal(t, tt.data, event.Data())
		})
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	tests := []struct {
		name     string
		subs     []*Subscription
		wantSubs map[EventType]int
	}{
		{
			name:     "nil subscription is ignored",
			subs:     []*Subscription{nil},
			wantSubs: map[EventType]int{EventTypeMessage: 0},
		},
		{
			name: "single subscription",
			subs: []*Subscription{
				{Type: EventTypeMessage, Handler: func(ctx context.Context, e *Event) error { return nil }},
			},
			wantSubs: map[EventType]int{EventTypeMessage: 1},
		},
		{
			name: "multiple subscriptions same type",
			subs: []*Subscription{
				{Type: EventTypeMessage, Handler: func(ctx context.Context, e *Event) error { return nil }},
				{Type: EventTypeMessage, Handler: func(ctx context.Context, e *Event) error { return nil }},
			},
			wantSubs: map[EventType]int{EventTypeMessage: 2},
		},
		{
			name: "multiple subscriptions different types",
			subs: []*Subscription{
				{Type: EventTypeMessage, Handler: func(ctx context.Context, e *Event) error { return nil }},
				{Type: EventTypeCommand, Handler: func(ctx context.Context, e *Event) error { return nil }},
			},
			wantSubs: map[EventType]int{EventTypeMessage: 1, EventTypeCommand: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()

			for _, sub := range tt.subs {
				eb.Subscribe(sub)
			}

			for eventType, count := range tt.wantSubs {
				assert.Len(t, eb.Subscriptions(eventType), count)
			}
		})
	}
}

func TestEventBus_Publish_Priority(t *testing.T) {
	tests := []struct {
		name           string
		priorities     []int
		expectedOrder  []int
	}{
		{
			name:          "single handler",
			priorities:    []int{0},
			expectedOrder: []int{0},
		},
		{
			name:          "ascending priorities",
			priorities:    []int{1, 2, 3},
			expectedOrder: []int{1, 2, 3},
		},
		{
			name:          "descending priorities",
			priorities:    []int{3, 2, 1},
			expectedOrder: []int{1, 2, 3},
		},
		{
			name:          "mixed priorities",
			priorities:    []int{10, 1, 5},
			expectedOrder: []int{1, 5, 10},
		},
		{
			name:          "duplicate priorities preserve order",
			priorities:    []int{1, 1, 1},
			expectedOrder: []int{1, 1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()
			var order []int
			var mu sync.Mutex

			for _, priority := range tt.priorities {
				p := priority
				eb.Subscribe(&Subscription{
					Type:     EventTypeMessage,
					Priority: p,
					Handler: func(ctx context.Context, e *Event) error {
						mu.Lock()
						order = append(order, p)
						mu.Unlock()
						return nil
					},
				})
			}

			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			err := eb.Publish(event)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOrder, order)
		})
	}
}

func TestEventBus_Publish_Middleware(t *testing.T) {
	tests := []struct {
		name          string
		setupBus      func() (*EventBus, *[]string)
		expectedCalls []string
	}{
		{
			name: "global middleware wraps handler",
			setupBus: func() (*EventBus, *[]string) {
				eb := NewEventBus()
				calls := &[]string{}

				eb.Use(func(ctx context.Context, e *Event, next func() error) error {
					*calls = append(*calls, "global-before")
					err := next()
					*calls = append(*calls, "global-after")
					return err
				})

				eb.Subscribe(&Subscription{
					Type: EventTypeMessage,
					Handler: func(ctx context.Context, e *Event) error {
						*calls = append(*calls, "handler")
						return nil
					},
				})

				return eb, calls
			},
			expectedCalls: []string{"global-before", "handler", "global-after"},
		},
		{
			name: "local middleware wraps handler",
			setupBus: func() (*EventBus, *[]string) {
				eb := NewEventBus()
				calls := &[]string{}

				eb.Subscribe(&Subscription{
					Type: EventTypeMessage,
					Middleware: []MiddlewareFunc{
						func(ctx context.Context, e *Event, next func() error) error {
							*calls = append(*calls, "local-before")
							err := next()
							*calls = append(*calls, "local-after")
							return err
						},
					},
					Handler: func(ctx context.Context, e *Event) error {
						*calls = append(*calls, "handler")
						return nil
					},
				})

				return eb, calls
			},
			expectedCalls: []string{"local-before", "handler", "local-after"},
		},
		{
			name: "global then local middleware order",
			setupBus: func() (*EventBus, *[]string) {
				eb := NewEventBus()
				calls := &[]string{}

				eb.Use(func(ctx context.Context, e *Event, next func() error) error {
					*calls = append(*calls, "global-before")
					err := next()
					*calls = append(*calls, "global-after")
					return err
				})

				eb.Subscribe(&Subscription{
					Type: EventTypeMessage,
					Middleware: []MiddlewareFunc{
						func(ctx context.Context, e *Event, next func() error) error {
							*calls = append(*calls, "local-before")
							err := next()
							*calls = append(*calls, "local-after")
							return err
						},
					},
					Handler: func(ctx context.Context, e *Event) error {
						*calls = append(*calls, "handler")
						return nil
					},
				})

				return eb, calls
			},
			expectedCalls: []string{"global-before", "local-before", "handler", "local-after", "global-after"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb, calls := tt.setupBus()

			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			err := eb.Publish(event)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCalls, *calls)
		})
	}
}

func TestEventBus_Publish_MiddlewareShortCircuit(t *testing.T) {
	tests := []struct {
		name          string
		callNext      bool
		wantHandlerCalled bool
	}{
		{
			name:          "middleware calls next",
			callNext:      true,
			wantHandlerCalled: true,
		},
		{
			name:          "middleware short-circuits",
			callNext:      false,
			wantHandlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()
			handlerCalled := false

			eb.Subscribe(&Subscription{
				Type: EventTypeMessage,
				Middleware: []MiddlewareFunc{
					func(ctx context.Context, e *Event, next func() error) error {
						if tt.callNext {
							return next()
						}
						return nil
					},
				},
				Handler: func(ctx context.Context, e *Event) error {
					handlerCalled = true
					return nil
				},
			})

			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			_ = eb.Publish(event)

			assert.Equal(t, tt.wantHandlerCalled, handlerCalled)
		})
	}
}

func TestEventBus_Publish_ErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		handlerErrors    []error
		wantErr          error
		wantHandlerCalls int
	}{
		{
			name:             "no errors",
			handlerErrors:    []error{nil, nil},
			wantErr:          nil,
			wantHandlerCalls: 2,
		},
		{
			name:             "first handler error returned",
			handlerErrors:    []error{errors.New("first error"), nil},
			wantErr:          errors.New("first error"),
			wantHandlerCalls: 2,
		},
		{
			name:             "second handler also called after error",
			handlerErrors:    []error{errors.New("first error"), errors.New("second error")},
			wantErr:          errors.New("first error"),
			wantHandlerCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()
			var callCount int
			var mu sync.Mutex

			for i, handlerErr := range tt.handlerErrors {
				err := handlerErr
				priority := i
				eb.Subscribe(&Subscription{
					Type:     EventTypeMessage,
					Priority: priority,
					Handler: func(ctx context.Context, e *Event) error {
						mu.Lock()
						callCount++
						mu.Unlock()
						return err
					},
				})
			}

			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			err := eb.Publish(event)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantHandlerCalls, callCount)
		})
	}
}

func TestEventBus_Publish_NilEvent(t *testing.T) {
	eb := NewEventBus()
	eb.Subscribe(&Subscription{
		Type:    EventTypeMessage,
		Handler: func(ctx context.Context, e *Event) error { return nil },
	})

	err := eb.Publish(nil)
	assert.NoError(t, err)
}

func TestEventBus_Subscriptions_ReturnsCopy(t *testing.T) {
	eb := NewEventBus()

	eb.Subscribe(&Subscription{
		Type:    EventTypeMessage,
		Handler: func(ctx context.Context, e *Event) error { return nil },
	})

	// Get first copy
	subs1 := eb.Subscriptions(EventTypeMessage)
	require.Len(t, subs1, 1)

	// Add another subscription
	eb.Subscribe(&Subscription{
		Type:    EventTypeMessage,
		Handler: func(ctx context.Context, e *Event) error { return nil },
	})

	// Original copy should be unchanged
	assert.Len(t, subs1, 1, "original subscription copy was modified")

	// New query should show updated count
	subs2 := eb.Subscriptions(EventTypeMessage)
	assert.Len(t, subs2, 2)
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		wantCount  int
	}{
		{
			name:       "100 concurrent publishes",
			goroutines: 100,
			wantCount:  100,
		},
		{
			name:       "1000 concurrent publishes",
			goroutines: 1000,
			wantCount:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()
			var counter atomic.Int32

			eb.Subscribe(&Subscription{
				Type: EventTypeMessage,
				Handler: func(ctx context.Context, e *Event) error {
					counter.Add(1)
					return nil
				},
			})

			var wg sync.WaitGroup
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
					_ = eb.Publish(event)
				}()
			}
			wg.Wait()

			assert.Equal(t, int32(tt.wantCount), counter.Load())
		})
	}
}

func TestEventBus_ConcurrentSubscribeAndPublish(t *testing.T) {
	eb := NewEventBus()
	var publishCount atomic.Int32

	// Start with one subscription
	eb.Subscribe(&Subscription{
		Type: EventTypeMessage,
		Handler: func(ctx context.Context, e *Event) error {
			publishCount.Add(1)
			return nil
		},
	})

	var wg sync.WaitGroup

	// Publish concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			_ = eb.Publish(event)
		}()
	}

	// Subscribe concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Subscribe(&Subscription{
				Type: EventTypeMessage,
				Handler: func(ctx context.Context, e *Event) error {
					publishCount.Add(1)
					return nil
				},
			})
		}()
	}

	wg.Wait()

	// Should have at least 50 publish calls (original handler called 50 times)
	assert.GreaterOrEqual(t, publishCount.Load(), int32(50))
}

func TestEventBus_Use(t *testing.T) {
	tests := []struct {
		name            string
		middlewareCount int
	}{
		{
			name:            "single global middleware",
			middlewareCount: 1,
		},
		{
			name:            "multiple global middleware",
			middlewareCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := NewEventBus()
			var callCount int
			var mu sync.Mutex

			for i := 0; i < tt.middlewareCount; i++ {
				eb.Use(func(ctx context.Context, e *Event, next func() error) error {
					mu.Lock()
					callCount++
					mu.Unlock()
					return next()
				})
			}

			eb.Subscribe(&Subscription{
				Type:    EventTypeMessage,
				Handler: func(ctx context.Context, e *Event) error { return nil },
			})

			event := NewEvent(EventTypeMessage, context.Background(), nil, nil)
			_ = eb.Publish(event)

			assert.Equal(t, tt.middlewareCount, callCount)
		})
	}
}
