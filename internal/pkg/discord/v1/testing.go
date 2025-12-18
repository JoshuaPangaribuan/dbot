package v1

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// BotState represents the lifecycle state of the bot (exported for testing).
type BotState = botState

// Exported state constants for testing.
const (
	StateNew    BotState = stateNew
	StateOpen   BotState = stateOpen
	StateClosed BotState = stateClosed
)

// NewTestBot creates a Bot suitable for unit testing (no network connections).
// This is exported for use in external test packages.
func NewTestBot(log logger.Logger) *Bot {
	bot := &Bot{
		eventBus: NewEventBus(),
		store:    NewStateStore(),
		logger:   log,
		features: make([]Feature, 0),
		commands: make(map[string]*Command),
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
func (b *Bot) SetGuildIDForTest(guildID string) {
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

