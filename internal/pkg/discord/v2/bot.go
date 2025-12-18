// Package v2 provides a library-agnostic Discord bot framework.
// This package does not depend on any specific Discord library;
// instead, it uses adapters to connect to Discord.
package v2

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// botState represents the lifecycle state of the bot.
type botState int32

const (
	stateNew    botState = iota // Bot created, not yet connected
	stateOpen                   // Bot connected and running
	stateClosed                 // Bot has been closed
)

func (s botState) String() string {
	switch s {
	case stateNew:
		return "new"
	case stateOpen:
		return "open"
	case stateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Bot manages the Discord bot lifecycle, features, and event handling.
// It uses an adapter to communicate with Discord.
type Bot struct {
	mu       sync.RWMutex
	adapter  port.Adapter
	eventBus *EventBus
	store    *StateStore
	logger   logger.Logger

	// Feature system
	features []Feature
	commands map[string]*domain.Command

	baseCtx context.Context
	guildID domain.GuildID // empty for global commands

	// Lifecycle state (atomic for lock-free reads)
	state atomic.Int32
}

// Option configures a Bot.
type Option func(*Bot)

// WithGuildID sets a guild ID for guild-specific commands (useful for testing).
func WithGuildID(id domain.GuildID) Option {
	return func(b *Bot) {
		b.guildID = id
	}
}

// WithLogger sets a custom logger for the bot.
func WithLogger(l logger.Logger) Option {
	return func(b *Bot) {
		b.logger = l
	}
}

// WithBaseContext sets the base context for all operations.
func WithBaseContext(ctx context.Context) Option {
	return func(b *Bot) {
		if ctx != nil {
			b.baseCtx = ctx
		}
	}
}

// NewWithAdapter creates a new Bot with the given adapter and options.
func NewWithAdapter(adapter port.Adapter, opts ...Option) (*Bot, error) {
	if adapter == nil {
		return nil, errors.New("discord: adapter is required")
	}

	bot := &Bot{
		adapter:  adapter,
		eventBus: NewEventBus(),
		store:    NewStateStore(),
		logger:   logger.New(),
		features: make([]Feature, 0),
		commands: make(map[string]*domain.Command),
		baseCtx:  context.Background(),
	}
	bot.state.Store(int32(stateNew))

	for _, opt := range opts {
		opt(bot)
	}

	// Set up event handler
	adapter.SetEventHandler(bot.handleEvent)

	bot.logger.Info(context.Background(), "bot initialized", logger.Fields{"component": "discord/v2"})

	return bot, nil
}

// getState returns the current bot state.
func (b *Bot) getState() botState {
	return botState(b.state.Load())
}

// isOpen returns true if the bot is currently connected.
func (b *Bot) isOpen() bool {
	return b.getState() == stateOpen
}

// SetBaseContext sets the base context for all interactions.
func (b *Bot) SetBaseContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	b.mu.Lock()
	b.baseCtx = ctx
	b.mu.Unlock()
}

// EventBus returns the bot's event bus for subscribing to events.
func (b *Bot) EventBus() *EventBus {
	return b.eventBus
}

// State returns the bot's shared state store.
func (b *Bot) State() *StateStore {
	return b.store
}

// Logger returns the bot's logger.
func (b *Bot) Logger() logger.Logger {
	return b.logger
}

// Adapter returns the underlying adapter.
func (b *Bot) Adapter() port.Adapter {
	return b.adapter
}

// GuildID returns the configured guild ID (empty for global commands).
func (b *Bot) GuildID() domain.GuildID {
	return b.guildID
}

// RegisterFeature adds a feature to the bot and initializes it.
// Features should be registered before calling Open().
// If the feature implements CommandProvider, its commands will be registered
// with Discord during Open().
func (b *Bot) RegisterFeature(f Feature) error {
	if f == nil {
		return nil
	}

	// Check state - features must be registered before Open()
	if b.getState() != stateNew {
		return errors.New("discord: cannot register feature after bot is opened")
	}

	featureName := f.Name()
	b.logger.Info(context.Background(), "registering feature", logger.Fields{
		"component": "discord/v2",
		"feature":   featureName,
	})

	// Initialize the feature WITHOUT holding the lock (Init may call back into Bot)
	if err := f.Init(b); err != nil {
		return fmt.Errorf("init feature %s: %w", featureName, err)
	}

	// Get subscriptions WITHOUT holding the lock
	subs := f.Subscriptions()

	// Check if feature provides commands
	var featureCmds []*domain.Command
	if cp, ok := f.(CommandProvider); ok {
		featureCmds = cp.Commands()
	}

	// Now acquire lock to update internal state
	b.mu.Lock()
	defer b.mu.Unlock()

	// Register subscriptions (EventBus has its own lock)
	for _, sub := range subs {
		b.eventBus.Subscribe(sub)
		b.logger.Debug(context.Background(), "subscribed to event", logger.Fields{
			"component": "discord/v2",
			"feature":   featureName,
			"event":     sub.Type.String(),
		})
	}

	// Register commands from feature
	for _, cmd := range featureCmds {
		if cmd != nil {
			b.commands[cmd.Name] = cmd
			b.logger.Info(context.Background(), "registered command from feature", logger.Fields{
				"component": "discord/v2",
				"feature":   featureName,
				"command":   cmd.Name,
			})
		}
	}

	b.features = append(b.features, f)
	return nil
}

// Open connects to Discord and registers all commands.
func (b *Bot) Open() error {
	// Check and transition state atomically
	if !b.state.CompareAndSwap(int32(stateNew), int32(stateOpen)) {
		currentState := b.getState()
		if currentState == stateOpen {
			return nil // Already open
		}
		return errors.New("discord: bot has been closed")
	}

	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	// Open gateway WITHOUT holding the lock (network call)
	if err := b.adapter.Open(ctx); err != nil {
		b.state.Store(int32(stateNew)) // Revert state
		return fmt.Errorf("discord: open gateway: %w", err)
	}

	// Sync commands WITHOUT holding the lock (network call)
	if err := b.syncCommands(ctx); err != nil {
		// Use Disconnect instead of Close so the bot can retry Open()
		if closeErr := b.adapter.Disconnect(); closeErr != nil {
			b.logger.Error(ctx, "failed to disconnect adapter after sync commands failure", logger.Fields{
				"component": "discord/v2",
				"error":     closeErr.Error(),
			})
		}
		b.state.Store(int32(stateNew)) // Revert state
		return fmt.Errorf("discord: sync commands: %w", err)
	}

	return nil
}

// Close disconnects from Discord and shuts down features.
func (b *Bot) Close() error {
	// Check and transition state atomically
	if !b.state.CompareAndSwap(int32(stateOpen), int32(stateClosed)) {
		return nil // Already closed or never opened
	}

	// Copy features list under lock to iterate safely
	b.mu.RLock()
	features := make([]Feature, len(b.features))
	copy(features, b.features)
	b.mu.RUnlock()

	// Shutdown features in reverse order WITHOUT holding the lock
	ctx := context.Background()
	for i := len(features) - 1; i >= 0; i-- {
		f := features[i]
		if err := f.Shutdown(ctx); err != nil {
			b.logger.Error(ctx, "feature shutdown error", logger.Fields{
				"component": "discord/v2",
				"feature":   f.Name(),
				"error":     err.Error(),
			})
		}
	}

	return b.adapter.Close()
}

// syncCommands registers all commands with Discord using bulk overwrite.
func (b *Bot) syncCommands(ctx context.Context) error {
	// Copy commands under lock
	b.mu.RLock()
	cmds := make([]*domain.Command, 0, len(b.commands))
	for _, cmd := range b.commands {
		cmds = append(cmds, cmd)
	}
	guildID := b.guildID
	b.mu.RUnlock()

	// Sync commands with Discord (even empty list to clear old commands)
	registered, err := b.adapter.BulkOverwriteCommands(ctx, guildID, cmds)
	if err != nil {
		return fmt.Errorf("bulk register commands: %w", err)
	}

	b.logger.Info(ctx, "synced commands", logger.Fields{
		"component": "discord/v2",
		"count":     len(registered),
		"guild_id":  guildID,
	})

	return nil
}

// handleEvent is called by the adapter when events are received.
func (b *Bot) handleEvent(e *domain.Event) {
	if e == nil {
		return
	}

	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	// Always apply the base context to ensure WithBaseContext propagates to handlers
	e = e.WithContext(ctx)

	// Publish to event bus
	if err := b.eventBus.Publish(e); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord/v2",
			"event":     e.Type().String(),
			"error":     err.Error(),
		})
	}
}

// BotState represents the lifecycle state of the bot (exported for testing).
type BotState = botState

// Exported state constants for testing.
const (
	StateNew    BotState = stateNew
	StateOpen   BotState = stateOpen
	StateClosed BotState = stateClosed
)

// NewTestBot creates a Bot suitable for unit testing (no adapter).
func NewTestBot(log logger.Logger) *Bot {
	bot := &Bot{
		eventBus: NewEventBus(),
		store:    NewStateStore(),
		logger:   log,
		features: make([]Feature, 0),
		commands: make(map[string]*domain.Command),
		baseCtx:  context.Background(),
	}
	bot.state.Store(int32(stateNew))
	return bot
}

// SetStateForTest sets the bot state for testing.
func (b *Bot) SetStateForTest(state BotState) {
	b.state.Store(int32(state))
}

// IsOpenForTest returns whether the bot is open (for testing).
func (b *Bot) IsOpenForTest() bool {
	return b.isOpen()
}

// SetGuildIDForTest sets the guild ID for testing.
func (b *Bot) SetGuildIDForTest(guildID domain.GuildID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.guildID = guildID
}

// FeatureCount returns the number of registered features (for testing).
func (b *Bot) FeatureCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.features)
}

// CommandCount returns the number of registered commands (for testing).
func (b *Bot) CommandCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.commands)
}
