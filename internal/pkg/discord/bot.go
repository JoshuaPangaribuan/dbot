package discord

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
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

// Bot wraps a discordgo session with event bus, features, and lifecycle handling.
type Bot struct {
	mu       sync.RWMutex
	session  *discordgo.Session // primary session (also used for REST calls)
	sessions []*discordgo.Session
	eventBus *EventBus
	store    *StateStore
	logger   logger.Logger

	// Feature system
	features       []Feature
	commands       map[string]*Command
	registeredCmds []*discordgo.ApplicationCommand

	// Shutdown hooks for cleanup
	shutdownHooks []func(context.Context) error

	baseCtx context.Context
	guildID string // empty for global commands

	// Sharding config
	shardCount      int
	autoShards      bool
	useGatewayBot   bool
	identifyDelay   time.Duration
	maxConcurrency  int

	// Lifecycle state (atomic for lock-free reads)
	state atomic.Int32

	// Lifecycle callbacks
	onReady      func(s *discordgo.Session, r *discordgo.Ready)
	onConnect    func(s *discordgo.Session, c *discordgo.Connect)
	onDisconnect func(s *discordgo.Session, d *discordgo.Disconnect)
	onResumed    func(s *discordgo.Session, r *discordgo.Resumed)
}

// Option configures a Bot.
type Option func(*Bot)

// WithGuildID sets a guild ID for guild-specific commands (useful for testing).
func WithGuildID(id string) Option {
	return func(b *Bot) {
		b.guildID = id
	}
}

// WithOnReady sets the Ready event handler.
func WithOnReady(fn func(s *discordgo.Session, r *discordgo.Ready)) Option {
	return func(b *Bot) {
		b.onReady = fn
	}
}

// WithOnConnect sets the Connect event handler.
func WithOnConnect(fn func(s *discordgo.Session, c *discordgo.Connect)) Option {
	return func(b *Bot) {
		b.onConnect = fn
	}
}

// WithOnDisconnect sets the Disconnect event handler.
func WithOnDisconnect(fn func(s *discordgo.Session, d *discordgo.Disconnect)) Option {
	return func(b *Bot) {
		b.onDisconnect = fn
	}
}

// WithOnResumed sets the Resumed event handler.
func WithOnResumed(fn func(s *discordgo.Session, r *discordgo.Resumed)) Option {
	return func(b *Bot) {
		b.onResumed = fn
	}
}

// WithLogger sets a custom logger for the bot.
func WithLogger(l logger.Logger) Option {
	return func(b *Bot) {
		b.logger = l
	}
}

// WithShardCount runs the bot with multiple gateway shards in a single process.
// Shards [0..count-1] will be started. count must be >= 1.
func WithShardCount(count int) Option {
	return func(b *Bot) {
		if count >= 1 {
			b.shardCount = count
		}
	}
}

// WithAutoSharding uses Discord's GatewayBot endpoint to determine the recommended shard count.
// This follows Discord guidance for sharding at scale.
func WithAutoSharding() Option {
	return func(b *Bot) {
		b.autoShards = true
	}
}

// WithGatewayBot enables/disables fetching GatewayBot info (recommended shards + session start limits).
// When sharding is enabled, this is enabled by default.
func WithGatewayBot(enabled bool) Option {
	return func(b *Bot) {
		b.useGatewayBot = enabled
	}
}

// WithIdentifyDelay configures the per-bucket identify delay when starting shards.
// Discord imposes an identify limit per concurrency bucket (commonly 5 seconds).
func WithIdentifyDelay(d time.Duration) Option {
	return func(b *Bot) {
		if d > 0 {
			b.identifyDelay = d
		}
	}
}

// WithShardMaxConcurrency overrides the max_concurrency value used for shard startup.
// Prefer leaving this unset and letting GatewayBot provide the correct value.
func WithShardMaxConcurrency(max int) Option {
	return func(b *Bot) {
		if max > 0 {
			b.maxConcurrency = max
		}
	}
}

// New creates a new Bot with the given token and options.
func New(token string, opts ...Option) (*Bot, error) {
	if token == "" {
		return nil, errors.New("discord: token is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("discord: create session: %w", err)
	}

	bot := &Bot{
		session:       session,
		sessions:      []*discordgo.Session{session},
		eventBus:      NewEventBus(),
		store:         NewStateStore(),
		logger:        logger.New(),
		features:      make([]Feature, 0),
		commands:      make(map[string]*Command),
		baseCtx:       context.Background(),
		shardCount:    1,
		useGatewayBot: true,
		identifyDelay: 5 * time.Second,
	}
	bot.state.Store(int32(stateNew))

	for _, opt := range opts {
		opt(bot)
	}

	// Set required intents for message content, guild messages, voice, and reactions
	bot.session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessageReactions

	// Register event handlers on the primary session.
	bot.registerSessionHandlers(bot.session)

	bot.logger.Info(context.Background(), "bot initialized", logger.Fields{"component": "discord"})

	return bot, nil
}

func (b *Bot) registerSessionHandlers(s *discordgo.Session) {
	if s == nil {
		return
	}
	s.AddHandler(b.handleReady)
	s.AddHandler(b.handleConnect)
	s.AddHandler(b.handleDisconnect)
	s.AddHandler(b.handleResumed)
	s.AddHandler(b.handleInteractionCreate)
	s.AddHandler(b.handleMessageCreate)
	s.AddHandler(b.handleVoiceStateUpdate)
	s.AddHandler(b.handleMessageReactionAdd)
	s.AddHandler(b.handleMessageReactionRemove)
}

func (b *Bot) cloneSessionForShard(shardID, shardCount int) (*discordgo.Session, error) {
	if b.session == nil {
		return nil, errors.New("discord: cannot clone session before primary session exists")
	}

	s, err := discordgo.New(b.session.Identify.Token)
	if err != nil {
		return nil, fmt.Errorf("discord: create shard session: %w", err)
	}

	// Copy user-configurable settings from the primary session so behavior is consistent.
	s.Debug = b.session.Debug
	s.LogLevel = b.session.LogLevel
	s.ShouldReconnectOnError = b.session.ShouldReconnectOnError
	s.ShouldReconnectVoiceOnSessionError = b.session.ShouldReconnectVoiceOnSessionError
	s.ShouldRetryOnRateLimit = b.session.ShouldRetryOnRateLimit
	s.StateEnabled = b.session.StateEnabled
	s.SyncEvents = b.session.SyncEvents
	s.MaxRestRetries = b.session.MaxRestRetries
	s.Client = b.session.Client
	s.Dialer = b.session.Dialer
	s.UserAgent = b.session.UserAgent
	s.Identify = b.session.Identify

	s.ShardID = shardID
	s.ShardCount = shardCount

	b.registerSessionHandlers(s)
	return s, nil
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

// Session returns the underlying discordgo session.
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

// Sessions returns a copy of all shard sessions (len==1 when sharding is disabled).
func (b *Bot) Sessions() []*discordgo.Session {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]*discordgo.Session, len(b.sessions))
	copy(out, b.sessions)
	return out
}

// GuildID returns the configured guild ID (empty for global commands).
func (b *Bot) GuildID() string {
	return b.guildID
}

// SlashCommand registers a slash command with its handler.
// This is the simplest way to add a command - no Feature interface needed.
func (b *Bot) SlashCommand(name, description string, handler CommandHandler, opts ...CommandOption) {
	// Build command options
	var cmdOpts []*discordgo.ApplicationCommandOption
	for _, opt := range opts {
		o := &discordgo.ApplicationCommandOption{}
		opt(o)
		cmdOpts = append(cmdOpts, o)
	}

	// Register command
	cmd := NewSlashCommand(name, description)
	if len(cmdOpts) > 0 {
		cmd.Options = cmdOpts
	}

	b.mu.Lock()
	b.commands[name] = cmd
	b.mu.Unlock()

	// Subscribe to command events
	b.eventBus.Subscribe(&Subscription{
		Type: EventTypeCommand,
		Handler: CommandFilter(name, func(ctx context.Context, e *Event) error {
			ic, ok := e.Data().(*discordgo.InteractionCreate)
			if !ok {
				return nil
			}
			cmdCtx := NewCommandContext(NewContext(e, b.store), ic, cmd)
			return handler(ctx, cmdCtx)
		}),
	})

	b.logger.Info(context.Background(), "registered slash command", logger.Fields{
		"component": "discord",
		"command":   name,
	})
}

// OnMessage registers a message handler.
// This is the simplest way to handle messages - no Feature interface needed.
func (b *Bot) OnMessage(handler MessageHandler, mw ...MiddlewareFunc) {
	b.eventBus.Subscribe(&Subscription{
		Type: EventTypeMessage,
		Handler: func(ctx context.Context, e *Event) error {
			m, ok := e.Data().(*discordgo.MessageCreate)
			if !ok {
				return nil
			}
			msgCtx := NewMessageContext(NewContext(e, b.store), m)
			return handler(ctx, msgCtx)
		},
		Middleware: mw,
	})

	b.logger.Info(context.Background(), "registered message handler", logger.Fields{
		"component": "discord",
	})
}

// OnReaction registers a reaction add handler.
func (b *Bot) OnReaction(handler ReactionHandler, mw ...MiddlewareFunc) {
	b.eventBus.Subscribe(&Subscription{
		Type: EventTypeReactionAdd,
		Handler: func(ctx context.Context, e *Event) error {
			r, ok := e.Data().(*discordgo.MessageReactionAdd)
			if !ok {
				return nil
			}
			rCtx := NewReactionAddContext(NewContext(e, b.store), r)
			return handler(ctx, rCtx)
		},
		Middleware: mw,
	})
}

// OnShutdown registers a cleanup function to run when the bot closes.
func (b *Bot) OnShutdown(fn func(context.Context) error) {
	b.mu.Lock()
	b.shutdownHooks = append(b.shutdownHooks, fn)
	b.mu.Unlock()
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
		"component": "discord",
		"feature":   featureName,
	})

	// Initialize the feature WITHOUT holding the lock (Init may call back into Bot)
	if err := f.Init(b); err != nil {
		return fmt.Errorf("init feature %s: %w", featureName, err)
	}

	// Get subscriptions WITHOUT holding the lock
	subs := f.Subscriptions()

	// Check if feature provides commands
	var featureCmds []*Command
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
			"component": "discord",
			"feature":   featureName,
			"event":     sub.Type.String(),
		})
	}

	// Register commands from feature
	for _, cmd := range featureCmds {
		if cmd != nil {
			b.commands[cmd.Name] = cmd
			b.logger.Info(context.Background(), "registered command from feature", logger.Fields{
				"component": "discord",
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

	// Build shard sessions (may call REST to fetch GatewayBot info).
	sessions, err := b.buildShardSessions()
	if err != nil {
		b.state.Store(int32(stateNew)) // Revert state
		return err
	}

	// Open sessions WITHOUT holding the lock (network calls).
	if err := b.openSessionsWithSharding(sessions); err != nil {
		for _, s := range sessions {
			if s != nil {
				_ = s.Close()
			}
		}
		b.state.Store(int32(stateNew)) // Revert state
		return err
	}

	// Store sessions under lock only after successful startup.
	b.mu.Lock()
	b.sessions = sessions
	// Keep b.session pointing at the primary (shard 0) session when present.
	for _, s := range sessions {
		if s != nil && s.ShardID == 0 {
			b.session = s
			break
		}
	}
	// If shard 0 isn't present (custom shard sets), fall back to the first session.
	if b.session == nil && len(sessions) > 0 {
		b.session = sessions[0]
	}
	b.mu.Unlock()

	// Sync commands WITHOUT holding the lock (network call)
	if err := b.syncCommands(); err != nil {
		for _, s := range sessions {
			if s == nil {
				continue
			}
			if closeErr := s.Close(); closeErr != nil {
				b.logger.Error(context.Background(), "failed to close session after sync commands failure", logger.Fields{
					"component": "discord",
					"error":     closeErr.Error(),
					"shard_id":  s.ShardID,
				})
			}
		}
		b.state.Store(int32(stateNew)) // Revert state
		return fmt.Errorf("discord: sync commands: %w", err)
	}

	return nil
}

// Close disconnects from Discord, shuts down features, and removes registered commands.
func (b *Bot) Close() error {
	// Check and transition state atomically
	if !b.state.CompareAndSwap(int32(stateOpen), int32(stateClosed)) {
		return nil // Already closed or never opened
	}

	// Copy features and hooks under lock to iterate safely
	b.mu.RLock()
	sessions := make([]*discordgo.Session, len(b.sessions))
	copy(sessions, b.sessions)
	features := make([]Feature, len(b.features))
	copy(features, b.features)
	hooks := make([]func(context.Context) error, len(b.shutdownHooks))
	copy(hooks, b.shutdownHooks)
	registeredCmds := make([]*discordgo.ApplicationCommand, len(b.registeredCmds))
	copy(registeredCmds, b.registeredCmds)
	guildID := b.guildID
	b.mu.RUnlock()

	ctx := context.Background()

	// Run shutdown hooks in reverse order
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil {
			b.logger.Error(ctx, "shutdown hook error", logger.Fields{
				"component": "discord",
				"error":     err.Error(),
			})
		}
	}

	// Shutdown features in reverse order WITHOUT holding the lock
	for i := len(features) - 1; i >= 0; i-- {
		f := features[i]
		if err := f.Shutdown(ctx); err != nil {
			b.logger.Error(ctx, "feature shutdown error", logger.Fields{
				"component": "discord",
				"feature":   f.Name(),
				"error":     err.Error(),
			})
		}
	}

	// Remove registered commands WITHOUT holding the lock (network calls)
	if b.session != nil && b.session.State != nil && b.session.State.User != nil {
		appID := b.session.State.User.ID
		for _, cmd := range registeredCmds {
			_ = b.session.ApplicationCommandDelete(appID, guildID, cmd.ID)
		}
	}

	// Clear internal state under lock
	b.mu.Lock()
	b.registeredCmds = nil
	b.mu.Unlock()

	var firstErr error
	for _, s := range sessions {
		if s == nil {
			continue
		}
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// syncCommands registers all commands with Discord using bulk overwrite (idempotent).
func (b *Bot) syncCommands() error {
	// Get application ID - prefer cached State, fall back to API call
	var appID string
	if b.session.State != nil && b.session.State.User != nil && b.session.State.User.ID != "" {
		appID = b.session.State.User.ID
	} else {
		// Fetch via API (State.User may not be populated until Ready event)
		user, err := b.session.User("@me")
		if err != nil {
			return fmt.Errorf("fetch bot user: %w", err)
		}
		appID = user.ID
	}

	// Copy commands under lock
	b.mu.RLock()
	cmds := make([]*discordgo.ApplicationCommand, 0, len(b.commands))
	for _, cmd := range b.commands {
		cmds = append(cmds, cmd.toApplicationCommand())
	}
	guildID := b.guildID
	b.mu.RUnlock()

	// Sync commands with Discord (even empty list to clear old commands)
	registered, err := b.session.ApplicationCommandBulkOverwrite(appID, guildID, cmds)
	if err != nil {
		return fmt.Errorf("bulk register commands: %w", err)
	}

	// Store registered commands under lock
	b.mu.Lock()
	b.registeredCmds = registered
	b.mu.Unlock()

	b.logger.Info(context.Background(), "synced commands", logger.Fields{
		"component": "discord",
		"count":     len(registered),
		"guild_id":  guildID,
	})

	return nil
}

func (b *Bot) buildShardSessions() ([]*discordgo.Session, error) {
	b.mu.RLock()
	shardCount := b.shardCount
	autoShards := b.autoShards
	useGatewayBot := b.useGatewayBot
	maxConcurrency := b.maxConcurrency
	identifyDelay := b.identifyDelay
	baseSession := b.session
	b.mu.RUnlock()

	if baseSession == nil {
		return nil, errors.New("discord: primary session is nil")
	}

	if autoShards && !useGatewayBot {
		return nil, errors.New("discord: auto sharding requires GatewayBot")
	}

	// Auto-determine shard count and/or max concurrency using GatewayBot.
	if useGatewayBot && (autoShards || shardCount > 1) {
		gb, err := baseSession.GatewayBot()
		if err != nil {
			if autoShards {
				return nil, fmt.Errorf("discord: fetch gateway bot info: %w", err)
			}
			// Non-fatal if sharding is configured explicitly; we'll fall back to safe defaults.
		} else {
			if autoShards {
				shardCount = gb.Shards
			}
			if maxConcurrency <= 0 {
				maxConcurrency = gb.SessionStartLimit.MaxConcurrency
			}

			if gb.SessionStartLimit.Remaining < shardCount {
				return nil, fmt.Errorf(
					"discord: session start limit too low to start %d shards (remaining=%d reset_after=%dms)",
					shardCount,
					gb.SessionStartLimit.Remaining,
					gb.SessionStartLimit.ResetAfter,
				)
			}

			b.logger.Info(context.Background(), "gateway bot info", logger.Fields{
				"component":        "discord",
				"recommended":      gb.Shards,
				"max_concurrency":  gb.SessionStartLimit.MaxConcurrency,
				"remaining":        gb.SessionStartLimit.Remaining,
				"reset_after_ms":   gb.SessionStartLimit.ResetAfter,
			})
		}
	}

	if shardCount < 1 {
		shardCount = 1
	}
	if identifyDelay <= 0 {
		identifyDelay = 5 * time.Second
	}

	// Non-sharded.
	if shardCount == 1 {
		baseSession.ShardID = 0
		baseSession.ShardCount = 1
		return []*discordgo.Session{baseSession}, nil
	}

	// Default max concurrency for safety if GatewayBot wasn't available.
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	// Ensure the base session is shard 0.
	baseSession.ShardID = 0
	baseSession.ShardCount = shardCount

	sessions := make([]*discordgo.Session, 0, shardCount)
	sessions = append(sessions, baseSession)

	// Create remaining shard sessions.
	for shardID := 1; shardID < shardCount; shardID++ {
		s, err := b.cloneSessionForShard(shardID, shardCount)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	// Persist the runtime sharding knobs for startup/metrics.
	b.mu.Lock()
	b.shardCount = shardCount
	b.maxConcurrency = maxConcurrency
	b.identifyDelay = identifyDelay
	b.mu.Unlock()

	return sessions, nil
}

func (b *Bot) openSessionsWithSharding(sessions []*discordgo.Session) error {
	if len(sessions) == 0 {
		return nil
	}
	if len(sessions) == 1 {
		if err := sessions[0].Open(); err != nil {
			return fmt.Errorf("discord: open session: %w", err)
		}
		return nil
	}

	b.mu.RLock()
	maxConcurrency := b.maxConcurrency
	identifyDelay := b.identifyDelay
	b.mu.RUnlock()

	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	if identifyDelay <= 0 {
		identifyDelay = 5 * time.Second
	}

	// Group sessions into Discord identify concurrency buckets:
	// bucket = shard_id % max_concurrency
	buckets := make(map[int][]*discordgo.Session, maxConcurrency)
	for _, s := range sessions {
		if s == nil {
			continue
		}
		bucket := s.ShardID % maxConcurrency
		buckets[bucket] = append(buckets[bucket], s)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(sessions))

	for bucket := 0; bucket < maxConcurrency; bucket++ {
		shardSessions := buckets[bucket]
		if len(shardSessions) == 0 {
			continue
		}

		wg.Add(1)
		go func(bucket int, shardSessions []*discordgo.Session) {
			defer wg.Done()
			for i, s := range shardSessions {
				if err := s.Open(); err != nil {
					errCh <- fmt.Errorf("discord: open shard %d/%d (bucket %d): %w", s.ShardID, s.ShardCount, bucket, err)
					return
				}
				// Discord identifies are rate-limited per bucket; wait between shards in the same bucket.
				if i < len(shardSessions)-1 {
					time.Sleep(identifyDelay)
				}
			}
		}(bucket, shardSessions)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// Lifecycle handlers

func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	fields := logger.Fields{
		"component": "discord",
		"state":     b.getState().String(),
	}
	if r != nil && r.User != nil {
		fields["user_id"] = r.User.ID
		fields["user"] = r.User.Username
	}
	if r != nil {
		fields["guild_count"] = len(r.Guilds)
	}
	b.logger.Info(ctx, "lifecycle ready", fields)

	if b.onReady != nil {
		b.onReady(s, r)
	}
}

func (b *Bot) handleConnect(s *discordgo.Session, c *discordgo.Connect) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	b.logger.Info(ctx, "lifecycle connect", logger.Fields{
		"component": "discord",
		"state":     b.getState().String(),
	})

	if b.onConnect != nil {
		b.onConnect(s, c)
	}
}

func (b *Bot) handleDisconnect(s *discordgo.Session, d *discordgo.Disconnect) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	b.logger.Warn(ctx, "lifecycle disconnect", logger.Fields{
		"component": "discord",
		"state":     b.getState().String(),
	})

	if b.onDisconnect != nil {
		b.onDisconnect(s, d)
	}
}

func (b *Bot) handleResumed(s *discordgo.Session, r *discordgo.Resumed) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	b.logger.Info(ctx, "lifecycle resumed", logger.Fields{
		"component": "discord",
		"state":     b.getState().String(),
	})

	if b.onResumed != nil {
		b.onResumed(s, r)
	}
}

func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	// Determine event type based on interaction type
	var eventType EventType
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		eventType = EventTypeCommand
	case discordgo.InteractionMessageComponent:
		eventType = EventTypeComponent
	case discordgo.InteractionModalSubmit:
		eventType = EventTypeModal
	default:
		return
	}

	// Publish to event bus
	event := NewEvent(eventType, ctx, s, i)
	if err := b.eventBus.Publish(event); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord",
			"event":     eventType.String(),
			"error":     err.Error(),
		})
	}
}

func (b *Bot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author != nil && m.Author.Bot {
		return
	}

	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	// Publish to event bus
	event := NewEvent(EventTypeMessage, ctx, s, m)
	if err := b.eventBus.Publish(event); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord",
			"event":     EventTypeMessage.String(),
			"error":     err.Error(),
		})
	}
}

func (b *Bot) handleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	event := NewEvent(EventTypeVoiceStateUpdate, ctx, s, v)
	if err := b.eventBus.Publish(event); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord",
			"event":     EventTypeVoiceStateUpdate.String(),
			"error":     err.Error(),
		})
	}
}

func (b *Bot) handleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	event := NewEvent(EventTypeReactionAdd, ctx, s, r)
	if err := b.eventBus.Publish(event); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord",
			"event":     EventTypeReactionAdd.String(),
			"error":     err.Error(),
		})
	}
}

func (b *Bot) handleMessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	b.mu.RLock()
	ctx := b.baseCtx
	b.mu.RUnlock()

	event := NewEvent(EventTypeReactionRemove, ctx, s, r)
	if err := b.eventBus.Publish(event); err != nil {
		b.logger.Error(ctx, "event bus error", logger.Fields{
			"component": "discord",
			"event":     EventTypeReactionRemove.String(),
			"error":     err.Error(),
		})
	}
}
