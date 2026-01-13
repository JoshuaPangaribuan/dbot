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
	return &application{
		flushers: make([]Flusher, 0),
	}
}

func (a *application) Start() {
	configPath := os.Getenv("DBOT_CONFIG_PATH")
	if configPath == "" {
		configPath = "opt/config/config.yaml"
	}

	cfg, err := initConfig(configPath)
	if err != nil {
		fmt.Println("Failed to load config:", err)
		panic(err)
	}
	a.config = cfg
	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return a.config.Close()
	})

	log := initLogger()
	if log == nil {
		fmt.Println("Failed to initialize logger")
		panic("Failed to initialize logger")
	}
	a.logger = log
	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return a.logger.Close()
	})

	a.logger.Info(context.Background(), "Application started", nil)

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		token = a.config.GetString("discord.token")
	}
	if token == "" {
		a.logger.Error(context.Background(), "Missing discord token (set DISCORD_TOKEN or discord.token in config)", nil)
		panic("missing discord token")
	}

	// Sharding config (single binary, many shards)
	// Defaults keep sharding disabled unless explicitly enabled.
	shardEnabled := false
	autoShards := false
	shardCount := 1
	useGatewayBot := true
	identifyDelayStr := "5s"

	if v := a.config.GetString("discord.sharding.enabled"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			shardEnabled = b
		}
	}
	if v := a.config.GetString("discord.sharding.auto"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			autoShards = b
		}
	}
	if n := a.config.GetInt("discord.sharding.count"); n > 0 {
		shardCount = n
	}
	if v := a.config.GetString("discord.sharding.use_gateway_bot"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			useGatewayBot = b
		}
	}
	if v := a.config.GetString("discord.sharding.identify_delay"); v != "" {
		identifyDelayStr = v
	}

	// Env overrides (takes precedence over config)
	if v := os.Getenv("DISCORD_SHARDING"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			shardEnabled = b
		}
	}
	if v := os.Getenv("DISCORD_AUTO_SHARDING"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			autoShards = b
		}
	}
	if v := os.Getenv("DISCORD_SHARD_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			shardCount = n
		}
	}
	if v := os.Getenv("DISCORD_USE_GATEWAY_BOT"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			useGatewayBot = b
		}
	}
	if v := os.Getenv("DISCORD_IDENTIFY_DELAY"); v != "" {
		identifyDelayStr = v
	}

	if autoShards || shardCount > 1 {
		shardEnabled = true
	}

	var identifyDelay time.Duration
	if identifyDelayStr != "" {
		if d, err := time.ParseDuration(identifyDelayStr); err == nil {
			identifyDelay = d
		}
	}

	// Guild ID for guild-specific commands (instant sync, recommended for development)
	guildID := a.config.GetString("discord.guild_id")
	if v := os.Getenv("DISCORD_GUILD_ID"); v != "" {
		guildID = v
	}

	// Setup Bot
	botOpts := []discord.Option{
		discord.WithLogger(a.logger),
	}
	if guildID != "" {
		botOpts = append(botOpts, discord.WithGuildID(guildID))
		a.logger.Info(context.Background(), "Using guild-specific commands", logger.Fields{"guild_id": guildID})
	}
	if shardEnabled {
		botOpts = append(botOpts, discord.WithGatewayBot(useGatewayBot))
		if identifyDelay > 0 {
			botOpts = append(botOpts, discord.WithIdentifyDelay(identifyDelay))
		}
		if autoShards {
			botOpts = append(botOpts, discord.WithAutoSharding())
		} else if shardCount > 1 {
			botOpts = append(botOpts, discord.WithShardCount(shardCount))
		}
	}

	bot, err := discord.New(token, botOpts...)
	if err != nil {
		a.logger.Error(context.Background(), "Failed to create discord bot", logger.Fields{"error": err})
		panic(err)
	}
	bot.SetBaseContext(context.Background())
	a.setupGlobalMiddleware(bot)
	a.registerFeatures(bot)

	// Connect to Discord
	if err := bot.Open(); err != nil {
		a.logger.Error(context.Background(), "Failed to connect discord bot", logger.Fields{"error": err})
		panic(err)
	}
	a.flushers = append(a.flushers, func(ctx context.Context) error {
		return bot.Close()
	})
	a.logger.Info(context.Background(), "Discord bot connected", nil)

	// Graceful shutdown
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
	// Register levelling feature using the new simplified pattern
	levelling.Register(levelling.Deps{
		Bot:    bot,
		Logger: a.logger,
		Config: levelling.Config{
			XPPerMessage: 10,
			Thresholds:   []int{100, 300, 600, 1000, 1500, 2100, 2800, 3600, 4500, 5500},
			AnnounceUp:   true,
		},
	})

	// Register ask feature with LangChain integration
	// Get langchain config from config file
	langchainProvider := langchain.Provider(a.config.GetString("langchain.provider"))
	if langchainProvider == "" {
		langchainProvider = langchain.ProviderOpenAI
	}

	// Get config values
	apiKey := a.config.GetString("langchain.api_key")
	baseURL := a.config.GetString("langchain.base_url")
	model := a.config.GetString("langchain.model")

	// Set default models based on provider if not specified
	switch langchainProvider {
	case langchain.ProviderAnthropic:
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}
	case langchain.ProviderOllama:
		if model == "" {
			model = "llama3.2"
		}
	default:
		if model == "" {
			model = "gpt-4o-mini"
		}
	}

	// Get temperature and max_tokens from config, with defaults
	temperature := a.config.GetFloat64("langchain.temperature")
	if temperature == 0 {
		temperature = 0.7
	}
	maxTokens := a.config.GetInt("langchain.max_tokens")
	if maxTokens == 0 {
		maxTokens = 1000
	}

	// Get system prompt from config, with default
	systemPrompt := a.config.GetString("langchain.system_prompt")
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant. Always respond in English."
	}

	_, err := ask.Register(ask.Deps{
		Bot:    bot,
		Logger: a.logger,
		Config: ask.Config{
			Provider:     langchainProvider,
			Model:        model,
			APIKey:       apiKey,
			BaseURL:      baseURL,
			Temperature:  float32(temperature),
			MaxTokens:    maxTokens,
			SystemPrompt: systemPrompt,
		},
	})
	if err != nil {
		a.logger.Error(context.Background(), "Failed to register ask feature", logger.Fields{"error": err})
		// Don't panic - the bot can run without the ask feature
	}
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
