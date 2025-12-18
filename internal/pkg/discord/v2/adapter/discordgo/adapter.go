package discordgo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
	"github.com/bwmarrin/discordgo"
)

// Adapter implements port.Adapter using the bwmarrin/discordgo library.
type Adapter struct {
	mu           sync.RWMutex
	session      *discordgo.Session
	eventHandler port.EventHandler
	botUser      *domain.User
	appID        domain.ApplicationID
	connected    atomic.Bool
	closed       atomic.Bool
	latency      atomic.Int64
}

// New creates a new discordgo adapter.
func New(cfg port.AdapterConfig) (*Adapter, error) {
	if cfg.Token == "" {
		return nil, errors.New("discordgo: token is required")
	}

	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("discordgo: create session: %w", err)
	}

	// Set intents
	intents := intentsToDiscordgo(cfg.Intents)
	if intents == 0 {
		intents = intentsToDiscordgo(port.DefaultIntents())
	}
	session.Identify.Intents = intents

	// Apply sharding configuration if specified
	if cfg.ShardCount > 0 {
		session.ShardID = cfg.ShardID
		session.ShardCount = cfg.ShardCount
	}

	adapter := &Adapter{
		session: session,
	}

	// Register event handlers
	session.AddHandler(adapter.handleReady)
	session.AddHandler(adapter.handleInteractionCreate)
	session.AddHandler(adapter.handleMessageCreate)
	session.AddHandler(adapter.handleVoiceStateUpdate)
	session.AddHandler(adapter.handleMessageReactionAdd)
	session.AddHandler(adapter.handleMessageReactionRemove)

	return adapter, nil
}

// intentsToDiscordgo converts domain intents to discordgo intents.
func intentsToDiscordgo(intents domain.Intent) discordgo.Intent {
	var result discordgo.Intent
	if intents.Has(domain.IntentGuilds) {
		result |= discordgo.IntentsGuilds
	}
	if intents.Has(domain.IntentGuildMembers) {
		result |= discordgo.IntentsGuildMembers
	}
	if intents.Has(domain.IntentGuildModeration) {
		result |= discordgo.IntentGuildBans
	}
	if intents.Has(domain.IntentGuildEmojisAndStickers) {
		result |= discordgo.IntentGuildEmojis
	}
	if intents.Has(domain.IntentGuildIntegrations) {
		result |= discordgo.IntentsGuildIntegrations
	}
	if intents.Has(domain.IntentGuildWebhooks) {
		result |= discordgo.IntentsGuildWebhooks
	}
	if intents.Has(domain.IntentGuildInvites) {
		result |= discordgo.IntentsGuildInvites
	}
	if intents.Has(domain.IntentGuildVoiceStates) {
		result |= discordgo.IntentsGuildVoiceStates
	}
	if intents.Has(domain.IntentGuildPresences) {
		result |= discordgo.IntentsGuildPresences
	}
	if intents.Has(domain.IntentGuildMessages) {
		result |= discordgo.IntentsGuildMessages
	}
	if intents.Has(domain.IntentGuildMessageReactions) {
		result |= discordgo.IntentsGuildMessageReactions
	}
	if intents.Has(domain.IntentGuildMessageTyping) {
		result |= discordgo.IntentsGuildMessageTyping
	}
	if intents.Has(domain.IntentDirectMessages) {
		result |= discordgo.IntentsDirectMessages
	}
	if intents.Has(domain.IntentDirectMessageReactions) {
		result |= discordgo.IntentsDirectMessageReactions
	}
	if intents.Has(domain.IntentDirectMessageTyping) {
		result |= discordgo.IntentsDirectMessageTyping
	}
	if intents.Has(domain.IntentMessageContent) {
		result |= discordgo.IntentMessageContent
	}
	if intents.Has(domain.IntentGuildScheduledEvents) {
		result |= discordgo.IntentsGuildScheduledEvents
	}
	// AutoModeration intents are not available in this version of discordgo
	// if intents.Has(domain.IntentAutoModerationConfiguration) { ... }
	// if intents.Has(domain.IntentAutoModerationExecution) { ... }
	return result
}

// --- Gateway Interface ---

// Open establishes the connection to Discord's gateway.
func (a *Adapter) Open(ctx context.Context) error {
	if a.closed.Load() {
		return port.ErrClosed
	}
	if a.connected.Load() {
		return port.ErrAlreadyConnected
	}

	if err := a.session.Open(); err != nil {
		return fmt.Errorf("discordgo: open session: %w", err)
	}

	a.connected.Store(true)
	return nil
}

// Close disconnects from the gateway gracefully and marks the adapter as permanently closed.
func (a *Adapter) Close() error {
	if !a.connected.Load() {
		return nil
	}

	a.closed.Store(true)
	a.connected.Store(false)

	return a.session.Close()
}

// Disconnect disconnects from the gateway without marking the adapter as permanently closed.
// This allows the adapter to be reopened with Open().
func (a *Adapter) Disconnect() error {
	if !a.connected.Load() {
		return nil
	}

	a.connected.Store(false)
	return a.session.Close()
}

// Reset clears the closed state, allowing the adapter to be reopened.
// This is useful for retry scenarios where Open() failed after connection.
func (a *Adapter) Reset() {
	a.closed.Store(false)
	a.connected.Store(false)
}

// SetEventHandler sets the callback function for received events.
func (a *Adapter) SetEventHandler(handler port.EventHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventHandler = handler
}

// BotUser returns the bot's user after connection is established.
func (a *Adapter) BotUser() *domain.User {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.botUser
}

// AppID returns the application ID after connection is established.
func (a *Adapter) AppID() domain.ApplicationID {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.appID
}

// Latency returns the gateway latency (heartbeat RTT).
func (a *Adapter) Latency() int64 {
	return a.latency.Load()
}

// Session returns the underlying discordgo session for advanced use cases.
func (a *Adapter) Session() *discordgo.Session {
	return a.session
}

// --- Event Handlers ---

func (a *Adapter) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	a.mu.Lock()
	a.botUser = userFromDiscordgo(r.User)
	if r.User != nil {
		a.appID = domain.ApplicationID(r.User.ID)
	}
	a.mu.Unlock()
}

func (a *Adapter) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	// Determine event type based on interaction type
	var eventType domain.EventType
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		eventType = domain.EventTypeCommand
	case discordgo.InteractionMessageComponent:
		eventType = domain.EventTypeComponent
	case discordgo.InteractionModalSubmit:
		eventType = domain.EventTypeModal
	case discordgo.InteractionApplicationCommandAutocomplete:
		eventType = domain.EventTypeAutocomplete
	default:
		return
	}

	event := domain.NewEvent(eventType, context.Background()).
		WithInteraction(interactionFromDiscordgo(i)).
		WithRaw(i)

	handler(event)
}

func (a *Adapter) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	// Skip bot messages at adapter level
	if m.Author != nil && m.Author.Bot {
		return
	}

	event := domain.NewEvent(domain.EventTypeMessage, context.Background()).
		WithMessage(messageFromDiscordgo(m.Message)).
		WithRaw(m)

	handler(event)
}

func (a *Adapter) handleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	event := domain.NewEvent(domain.EventTypeVoiceStateUpdate, context.Background()).
		WithVoiceState(voiceStateFromDiscordgo(v)).
		WithRaw(v)

	handler(event)
}

func (a *Adapter) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	event := domain.NewEvent(domain.EventTypeReactionAdd, context.Background()).
		WithReaction(&domain.MessageReaction{
			UserID:    domain.UserID(r.UserID),
			ChannelID: domain.ChannelID(r.ChannelID),
			MessageID: domain.MessageID(r.MessageID),
			GuildID:   domain.GuildID(r.GuildID),
			Member:    memberFromDiscordgo(r.Member, r.GuildID),
			Emoji:     emojiFromDiscordgo(r.Emoji),
			Removed:   false,
		}).
		WithRaw(r)

	handler(event)
}

func (a *Adapter) handleMessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	a.mu.RLock()
	handler := a.eventHandler
	a.mu.RUnlock()

	if handler == nil {
		return
	}

	event := domain.NewEvent(domain.EventTypeReactionRemove, context.Background()).
		WithReaction(&domain.MessageReaction{
			UserID:    domain.UserID(r.UserID),
			ChannelID: domain.ChannelID(r.ChannelID),
			MessageID: domain.MessageID(r.MessageID),
			GuildID:   domain.GuildID(r.GuildID),
			Emoji:     emojiFromDiscordgo(r.Emoji),
			Removed:   true,
		}).
		WithRaw(r)

	handler(event)
}

// Verify Adapter implements port.Adapter
var _ port.Adapter = (*Adapter)(nil)
