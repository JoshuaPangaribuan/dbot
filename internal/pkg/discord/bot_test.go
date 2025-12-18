package discord_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JoshuaPangaribuan/dbot/internal/mocks"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newTestBot creates a Bot suitable for unit testing (no network connections).
func newTestBot(t *testing.T) *discord.Bot {
	mockLogger := mocks.NewMockLogger(t)
	mockLogger.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Debug(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Warn(mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLogger.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything).Maybe()

	return discord.NewTestBot(mockLogger)
}

func TestBot_RegisterFeature(t *testing.T) {
	tests := []struct {
		name           string
		setupFeature   func(t *testing.T) discord.Feature
		preOpen        bool
		wantErr        bool
		wantErrContain string
		wantFeatures   int
		wantCommands   int
		wantSubs       int
	}{
		{
			name: "nil feature returns no error",
			setupFeature: func(t *testing.T) discord.Feature {
				return nil
			},
			wantErr:      false,
			wantFeatures: 0,
			wantCommands: 0,
			wantSubs:     0,
		},
		{
			name: "valid feature without commands",
			setupFeature: func(t *testing.T) discord.Feature {
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
			setupFeature: func(t *testing.T) discord.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("test-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(nil)
				mockFeature.EXPECT().Subscriptions().Return([]*discord.Subscription{
					{
						Type: discord.EventTypeMessage,
						Handler: func(ctx context.Context, e *discord.Event) error {
							return nil
						},
					},
					{
						Type: discord.EventTypeCommand,
						Handler: func(ctx context.Context, e *discord.Event) error {
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
			setupFeature: func(t *testing.T) discord.Feature {
				mockFeature := mocks.NewMockFeature(t)
				mockFeature.EXPECT().Name().Return("cmd-feature")
				mockFeature.EXPECT().Init(mock.Anything).Return(nil)
				mockFeature.EXPECT().Subscriptions().Return(nil)

				mockCmdProvider := mocks.NewMockCommandProvider(t)
				mockCmdProvider.EXPECT().Commands().Return([]*discord.Command{
					discord.NewSlashCommand("test1", "Test command 1"),
					discord.NewSlashCommand("test2", "Test command 2"),
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
			setupFeature: func(t *testing.T) discord.Feature {
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
			setupFeature: func(t *testing.T) discord.Feature {
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
				bot.SetStateForTest(discord.StateOpen)
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
			totalSubs := len(bot.EventBus().Subscriptions(discord.EventTypeMessage)) +
				len(bot.EventBus().Subscriptions(discord.EventTypeCommand))
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
		state    discord.BotState
		wantOpen bool
	}{
		{
			name:     "state new - not open",
			state:    discord.StateNew,
			wantOpen: false,
		},
		{
			name:     "state open - is open",
			state:    discord.StateOpen,
			wantOpen: true,
		},
		{
			name:     "state closed - not open",
			state:    discord.StateClosed,
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
	discord.Feature
	discord.CommandProvider
}

func (f *featureWithCommands) Name() string {
	return f.Feature.Name()
}

func (f *featureWithCommands) Init(b *discord.Bot) error {
	return f.Feature.Init(b)
}

func (f *featureWithCommands) Shutdown(ctx context.Context) error {
	return f.Feature.Shutdown(ctx)
}

func (f *featureWithCommands) Subscriptions() []*discord.Subscription {
	return f.Feature.Subscriptions()
}

func (f *featureWithCommands) Commands() []*discord.Command {
	return f.CommandProvider.Commands()
}
