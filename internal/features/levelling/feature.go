// Package levelling tracks user activity and assigns levels based on XP.
package levelling

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling/handlers"
	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling/persistence"
	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Config configures the levelling system.
type Config struct {
	XPPerMessage int   // XP gained per message (default: 10)
	Thresholds   []int // XP needed for each level (default: [100, 300, 600, 1000, 1500])
	AnnounceUp   bool  // Announce level-ups in channel
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		XPPerMessage: 10,
		Thresholds:   []int{100, 300, 600, 1000, 1500, 2100, 2800, 3600, 4500, 5500},
		AnnounceUp:   true,
	}
}

// Deps holds dependencies for the levelling feature.
type Deps struct {
	Bot    *discord.Bot
	Logger logger.Logger
	Config Config
}

// Feature wires together the levelling feature components.
type Feature struct {
	svc     service.Service
	persist *persistence.Persist
	handler *handlers.Handler
}

// Register sets up the levelling feature with the bot.
func Register(deps Deps) *Feature {
	// Apply defaults if not set
	if deps.Config.XPPerMessage == 0 {
		deps.Config.XPPerMessage = 10
	}
	if len(deps.Config.Thresholds) == 0 {
		deps.Config.Thresholds = DefaultConfig().Thresholds
	}
	if deps.Logger == nil {
		deps.Logger = logger.New()
	}

	// Create service with config
	svcConfig := service.Config{
		XPPerMessage: deps.Config.XPPerMessage,
		Thresholds:   deps.Config.Thresholds,
	}
	svc := service.New(svcConfig)

	// Create persistence handler
	persist := persistence.New(svc, deps.Logger)

	// Create handler with service dependency
	handler := handlers.New(svc, deps.Config.AnnounceUp)

	f := &Feature{
		svc:     svc,
		persist: persist,
		handler: handler,
	}

	// Register handlers
	deps.Bot.OnMessage(handler.OnMessage)
	deps.Bot.SlashCommand("level", "Check your level", handler.CheckLevel,
		discord.UserOption("user", "User to check (optional)", false),
	)
	deps.Bot.OnShutdown(persist.Save)

	deps.Logger.Info(context.Background(), "levelling feature registered", logger.Fields{
		"component":      "levelling",
		"xp_per_message": deps.Config.XPPerMessage,
		"levels":         len(deps.Config.Thresholds),
	})

	return f
}

// Service returns the underlying service for external access if needed.
func (f *Feature) Service() service.Service {
	return f.svc
}
