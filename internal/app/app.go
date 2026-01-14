package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/features/ask"
	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling"
	"github.com/JoshuaPangaribuan/dbot/internal/features/voice"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/config"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	discordmw "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/middleware"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/langchain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

type Application interface {
	Start()
	Stop(context.Context) error
}

type Flusher func(context.Context) error

type application struct {
	logger   logger.Logger
	config   config.Config
	flushers []Flusher
}

func New() *application {
	return &application{}
}

func (a *application) Start() {
	a.initializeConfig()
	a.initializeLogger()
	a.logger.Info(context.Background(), "Application started", nil)

	token := a.config.GetString("discord.token")
	if token == "" {
		a.logger.Error(context.Background(), "Missing discord.token in config", nil)
		panic("missing discord token")
	}

	shardCfg := a.loadShardingConfig()
	guildID := a.config.GetString("discord.guild_id")

	bot := a.createBot(token, shardCfg, guildID)

	bot.SetBaseContext(context.Background())
	a.setupGlobalMiddleware(bot)
	a.registerFeatures(bot)

	a.connectBot(bot)
	a.waitForShutdown()
}

// createBot creates and configures the Discord bot with sharding options
func (a *application) createBot(token string, shardCfg shardingConfig, guildID string) *discord.Bot {
	botOpts := []discord.Option{
		discord.WithLogger(a.logger),
	}

	if guildID != "" {
		botOpts = append(botOpts, discord.WithGuildID(guildID))
		a.logger.Info(context.Background(), "Using guild-specific commands", logger.Fields{"guild_id": guildID})
	}

	if shardCfg.enabled {
		botOpts = append(botOpts, discord.WithGatewayBot(shardCfg.useGatewayBot))
		if shardCfg.identifyDelay > 0 {
			botOpts = append(botOpts, discord.WithIdentifyDelay(shardCfg.identifyDelay))
		}
		if shardCfg.autoShards {
			botOpts = append(botOpts, discord.WithAutoSharding())
		} else if shardCfg.shardCount > 1 {
			botOpts = append(botOpts, discord.WithShardCount(shardCfg.shardCount))
		}
	}

	bot, err := discord.New(token, botOpts...)
	if err != nil {
		a.logger.Error(context.Background(), "Failed to create discord bot", logger.Fields{"error": err})
		panic(err)
	}

	return bot
}

// connectBot opens the Discord connection and registers cleanup
func (a *application) connectBot(bot *discord.Bot) {
	if err := bot.Open(); err != nil {
		a.logger.Error(context.Background(), "Failed to connect discord bot", logger.Fields{"error": err})
		panic(err)
	}

	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return bot.Close()
	})
	a.logger.Info(context.Background(), "Discord bot connected", nil)
}

// waitForShutdown blocks until SIGINT/SIGTERM is received
func (a *application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.logger.Info(context.Background(), "Shutting down application...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Stop(ctx); err != nil {
		a.logger.Error(ctx, "Failed to gracefully shutdown", logger.Fields{"error": err})
	}
}

// Stop gracefully shuts down the application by flushing all resources in reverse order
func (a *application) Stop(ctx context.Context) error {
	var firstErr error
	// Reverse order: close in LIFO (bot first, logger last)
	for i := len(a.flushers) - 1; i >= 0; i-- {
		flusher := a.flushers[i]
		if err := flusher(ctx); err != nil {
			a.logger.Error(ctx, "Failed to flush", logger.Fields{
				"error": err,
			})
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	a.logger.Info(ctx, "Application stopped", nil)
	return firstErr
}

func (a *application) setupGlobalMiddleware(bot *discord.Bot) {
	bot.EventBus().Use(
		discord.RequireGuild(),
		discord.IgnoreBot(),
		onlyEventTypes(
			discordmw.RateLimit(5, 10*time.Second),
			discord.EventTypeMessage,
			discord.EventTypeCommand,
			discord.EventTypeComponent,
			discord.EventTypeModal,
		),
	)
}

func (a *application) registerFeatures(bot *discord.Bot) {
	a.registerLevellingFeature(bot)
	a.registerAskFeature(bot)
	a.registerVoiceFeature(bot)
}

// registerLevellingFeature registers the levelling system
func (a *application) registerLevellingFeature(bot *discord.Bot) {
	levelling.Register(levelling.Deps{
		Bot:    bot,
		Logger: a.logger,
		Config: levelling.Config{
			XPPerMessage: 10,
			Thresholds:   []int{100, 300, 600, 1000, 1500, 2100, 2800, 3600, 4500, 5500},
			AnnounceUp:   true,
		},
	})
}

// registerAskFeature registers the AI ask feature with LangChain integration
func (a *application) registerAskFeature(bot *discord.Bot) {
	_, err := ask.Register(ask.Deps{
		Bot:    bot,
		Logger: a.logger,
		Config: a.loadLangChainConfig(),
	})
	if err != nil {
		a.logger.Error(context.Background(), "Failed to register ask feature", logger.Fields{"error": err})
		// Don't panic - the bot can run without the ask feature
	}
}

// registerVoiceFeature registers the voice/music feature
func (a *application) registerVoiceFeature(bot *discord.Bot) {
	voice.Register(voice.Deps{
		Bot:    bot,
		Logger: a.logger,
		Config: a.loadVoiceConfig(),
	})
}

func onlyEventTypes(mw discord.MiddlewareFunc, types ...discord.EventType) discord.MiddlewareFunc {
	allowed := make(map[discord.EventType]struct{}, len(types))
	for _, typ := range types {
		allowed[typ] = struct{}{}
	}

	return func(ctx context.Context, e *discord.Event, next func() error) error {
		if _, ok := allowed[e.Type()]; !ok {
			return next()
		}
		return mw(ctx, e, next)
	}
}

func (a *application) initializeConfig() {
	cfg, err := initConfig("opt/config/config.yaml")
	if err != nil {
		fmt.Println("Failed to load config:", err)
		panic(err)
	}
	a.config = cfg
	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return a.config.Close()
	})
}

func (a *application) initializeLogger() {
	log := initLogger()
	if log == nil {
		fmt.Println("Failed to initialize logger")
		panic("Failed to initialize logger")
	}
	a.logger = log
	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return a.logger.Close()
	})
}

// parseBoolWithDefault parses a string to bool, returning defaultValue if parsing fails
func parseBoolWithDefault(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed
	}
	return defaultValue
}

// loadShardingConfig loads sharding configuration from config file
type shardingConfig struct {
	enabled       bool
	autoShards    bool
	shardCount    int
	useGatewayBot bool
	identifyDelay time.Duration
}

func (a *application) loadShardingConfig() shardingConfig {
	cfg := shardingConfig{
		enabled:       parseBoolWithDefault(a.config.GetString("discord.sharding.enabled"), false),
		autoShards:    parseBoolWithDefault(a.config.GetString("discord.sharding.auto"), false),
		shardCount:    a.config.GetInt("discord.sharding.count"),
		useGatewayBot: parseBoolWithDefault(a.config.GetString("discord.sharding.use_gateway_bot"), true),
	}

	// Parse identify delay
	if delayStr := a.config.GetString("discord.sharding.identify_delay"); delayStr != "" {
		if d, err := time.ParseDuration(delayStr); err == nil {
			cfg.identifyDelay = d
		}
	}

	// Auto-enable sharding if auto-sharding or multiple shards are configured
	if cfg.autoShards || cfg.shardCount > 1 {
		cfg.enabled = true
	}

	return cfg
}

// loadLangChainConfig loads LangChain configuration with provider-specific defaults
func (a *application) loadLangChainConfig() ask.Config {
	provider := langchain.Provider(a.config.GetString("langchain.provider"))
	if provider == "" {
		provider = langchain.ProviderOpenAI
	}

	model := a.config.GetString("langchain.model")
	if model == "" {
		// Set default models based on provider
		switch provider {
		case langchain.ProviderAnthropic:
			model = "claude-3-5-sonnet-20241022"
		case langchain.ProviderOllama:
			model = "llama3.2"
		default:
			model = "gpt-4o-mini"
		}
	}

	temperature := a.config.GetFloat64("langchain.temperature")
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := a.config.GetInt("langchain.max_tokens")
	if maxTokens == 0 {
		maxTokens = 1000
	}

	systemPrompt := a.config.GetString("langchain.system_prompt")
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant. Always respond in English."
	}

	return ask.Config{
		Provider:     provider,
		Model:        model,
		APIKey:       a.config.GetString("langchain.api_key"),
		BaseURL:      a.config.GetString("langchain.base_url"),
		Temperature:  float32(temperature),
		MaxTokens:    maxTokens,
		SystemPrompt: systemPrompt,
	}
}

// loadVoiceConfig loads voice feature configuration with sensible defaults
func (a *application) loadVoiceConfig() voice.Config {
	defaultVolume := a.config.GetInt("voice.default_volume")
	if defaultVolume == 0 {
		defaultVolume = 50
	}

	maxQueueSize := a.config.GetInt("voice.max_queue_size")
	if maxQueueSize == 0 {
		maxQueueSize = 20
	}

	return voice.Config{
		Enabled:           a.config.GetString("voice.enabled") == "true",
		DefaultVolume:     defaultVolume,
		MaxQueueSize:      maxQueueSize,
		YouTubeEnabled:    a.config.GetString("voice.youtube_enabled") == "true",
		LocalFilesEnabled: a.config.GetString("voice.local_files_enabled") == "true",
		AllowedChannels:   nil, // TODO: parse from config as array
	}
}
