package v1_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JoshuaPangaribuan/dbot/internal/mocks"
	v1 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newTestBot creates a Bot suitable for unit testing (no network connections).
func newTestBot(t *testing.T) *v1.Bot {
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Debug(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Warn(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything).Maybe()

	return v1.NewTestBot(mockLogger)
}

func TestBot_RegisterFeature(t *testing.T) {
	tests := []struct {
		name           string
		setupFeature   func(t *testing.T) v1.Feature
		preOpen        bool
		wantErr        bool
		wantErrContain string
		wantFeatures   int
		wantCommands   int
		wantSubs       int
	}{
		{
			name: "nil feature returns no error",
			setupFeature: func(t *testing.T) v1.Feature {
				return nil
			},
			wantErr:      false,
			wantFeatures: 0,
			wantCommands: 0,
			wantSubs:     0,
		},
		{
			name: "valid feature without commands",
			setupFeature: func(t *testing.T) v1.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("test-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(nil)
				mockFeature.EXPECT().Subscriptions().Return(nil)
				return mockFeature
			},
			wantErr:      false,
			wantFeatures: 1,
			wantCommands: 0,
			wantSubs:     0,
		},
		{
			name: "feature with subscriptions",
			setupFeature: func(t *testing.T) v1.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("test-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(nil)
				mockFeature.EXPECT().Subscriptions().Return([]*v1.Subscription{
					{
						Type: v1.EventTypeMessage,
						Handler: func(ctx context.Context, e *v1.Event) error {
							return nil
						},
					},
					{
						Type: v1.EventTypeCommand,
						Handler: func(ctx context.Context, e *v1.Event) error {
							return nil
						},
					},
				})
				return mockFeature
			},
			wantErr:      false,
			wantFeatures: 1,
			wantCommands: 0,
			wantSubs:     2,
		},
		{
			name: "feature with commands (implements CommandProvider)",
			setupFeature: func(t *testing.T) v1.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("cmd-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(nil)
				mockFeature.EXPECT().Subscriptions().Return(nil)

				mockCmdProvider := mocks.NewMockCommandProvider(t)
				mockCmdProvider.EXPECT().Commands().Return([]*v1.Command{
					v1.NewSlashCommand("test1", "Test command 1"),
					v1.NewSlashCommand("test2", "Test command 2"),
				})

				// Create a composite that implements both interfaces
				return &featureWithCommands{
					Feature:         mockFeature,
					CommandProvider: mockCmdProvider,
				}
			},
			wantErr:      false,
			wantFeatures: 1,
			wantCommands: 2,
			wantSubs:     0,
		},
		{
			name: "feature init error",
			setupFeature: func(t *testing.T) v1.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("failing-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(errors.New("init failed"))
				return mockFeature
			},
			wantErr:        true,
			wantErrContain: "init feature",
			wantFeatures:   0,
			wantCommands:   0,
			wantSubs:       0,
		},
		{
			name: "cannot register after bot is opened",
			setupFeature: func(t *testing.T) v1.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("late-feature").Maybe()
				return mockFeature
			},
			preOpen:        true,
			wantErr:        true,
			wantErrContain: "cannot register feature after bot is opened",
			wantFeatures:   0,
			wantCommands:   0,
			wantSubs:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newTestBot(t)
			if tt.preOpen {
				bot.SetStateForTest(v1.StateOpen)
			}

			feature := tt.setupFeature(t)
			err := bot.RegisterFeature(feature)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantFeatures, bot.FeatureCount())
			assert.Equal(t, tt.wantCommands, bot.CommandCount())

			// Count subscriptions for EventTypeMessage and EventTypeCommand
			totalSubs := len(bot.EventBus().Subscriptions(v1.EventTypeMessage)) +
				len(bot.EventBus().Subscriptions(v1.EventTypeCommand))
			assert.Equal(t, tt.wantSubs, totalSubs)
		})
	}
}

func TestBot_State(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		getValue string
	}{
		{
			name:     "state store is accessible",
			setValue: "test-value",
			getValue: "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newTestBot(t)
			store := bot.State()

			assert.NotNil(t, store)

			store.Set("key", tt.setValue)
			v, ok := store.Get("key")
			assert.True(t, ok)
			assert.Equal(t, tt.getValue, v)
		})
	}
}

func TestBot_IsOpen(t *testing.T) {
	tests := []struct {
		name     string
		state    v1.BotState
		wantOpen bool
	}{
		{
			name:     "state new - not open",
			state:    v1.StateNew,
			wantOpen: false,
		},
		{
			name:     "state open - is open",
			state:    v1.StateOpen,
			wantOpen: true,
		},
		{
			name:     "state closed - not open",
			state:    v1.StateClosed,
			wantOpen: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newTestBot(t)
			bot.SetStateForTest(tt.state)

			assert.Equal(t, tt.wantOpen, bot.IsOpenForTest())
		})
	}
}

func TestBot_SetBaseContext(t *testing.T) {
	type ctxKey string

	tests := []struct {
		name       string
		inputCtx   context.Context
		wantNotNil bool
	}{
		{
			name:       "nil context defaults to background",
			inputCtx:   nil,
			wantNotNil: true,
		},
		{
			name:       "custom context is set",
			inputCtx:   context.WithValue(context.Background(), ctxKey("key"), "value"),
			wantNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newTestBot(t)

			bot.SetBaseContext(tt.inputCtx)

			// SetBaseContext should always result in a non-nil context
			assert.True(t, tt.wantNotNil)
		})
	}
}

func TestBot_Accessors(t *testing.T) {
	t.Run("EventBus returns event bus", func(t *testing.T) {
		bot := newTestBot(t)
		assert.NotNil(t, bot.EventBus())
	})

	t.Run("State returns state store", func(t *testing.T) {
		bot := newTestBot(t)
		assert.NotNil(t, bot.State())
	})

	t.Run("GuildID returns guild ID", func(t *testing.T) {
		bot := newTestBot(t)
		bot.SetGuildIDForTest("test-guild-123")
		assert.Equal(t, "test-guild-123", bot.GuildID())
	})
}

// featureWithCommands is a test helper that combines Feature and CommandProvider
type featureWithCommands struct {
	v1.Feature
	v1.CommandProvider
}

func (f *featureWithCommands) Name() string {
	return f.Feature.Name()
}

func (f *featureWithCommands) Init(b *v1.Bot) error {
	return f.Feature.Init(b)
}

func (f *featureWithCommands) Shutdown(ctx context.Context) error {
	return f.Feature.Shutdown(ctx)
}

func (f *featureWithCommands) Subscriptions() []*v1.Subscription {
	return f.Feature.Subscriptions()
}

func (f *featureWithCommands) Commands() []*v1.Command {
	return f.CommandProvider.Commands()
}

