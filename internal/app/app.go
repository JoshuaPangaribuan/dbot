package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/features/echo"
	"github.com/JoshuaPangaribuan/dbot/internal/features/pingpong"
	"github.com/JoshuaPangaribuan/dbot/internal/features/thumbsup"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/config"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	discordmw "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/middleware"
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
		configPath = "opt/config/db.yaml"
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

	// Setup Bot
	bot, err := discord.New(token)
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
	features := []struct {
		name string
		f    discord.Feature
	}{
		{
			name: "pingpong",
			f: pingpong.NewFeature(pingpong.Deps{
				Logger: a.logger,
			}),
		},
		{
			name: "echo",
			f: echo.NewFeature(echo.Deps{
				Logger: a.logger,
			}),
		},
		{
			name: "thumbsup",
			f: thumbsup.NewFeature(thumbsup.Deps{
				Logger: a.logger,
			}),
		},
	}

	for _, feature := range features {
		if err := bot.RegisterFeature(feature.f); err != nil {
			a.logger.Error(context.Background(), "Failed to register feature", logger.Fields{
				"feature": feature.name,
				"error":   err,
			})
			panic(err)
		}
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
