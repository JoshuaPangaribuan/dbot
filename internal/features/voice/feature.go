// Package voice provides music/audio playback functionality for Discord voice channels.
package voice

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/features/voice/handlers"
	"github.com/JoshuaPangaribuan/dbot/internal/features/voice/persistence"
	"github.com/JoshuaPangaribuan/dbot/internal/features/voice/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Config configures the voice feature.
type Config struct {
	Enabled          bool     // Enable voice feature
	DefaultVolume    int      // Default volume (0-100)
	MaxQueueSize     int      // Maximum tracks in queue
	YouTubeEnabled   bool     // Enable YouTube audio source
	LocalFilesEnabled bool    // Enable local file uploads
	AllowedChannels  []string // Whitelist voice channel IDs (empty = all channels)
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		DefaultVolume:     50,
		MaxQueueSize:      20,
		YouTubeEnabled:    true,
		LocalFilesEnabled: true,
		AllowedChannels:   nil, // nil = all channels allowed
	}
}

// Deps holds dependencies for the voice feature.
type Deps struct {
	Bot    *discord.Bot
	Logger logger.Logger
	Config Config
}

// Feature wires together the voice feature components.
type Feature struct {
	svc      service.VoiceService
	persist  *persistence.Persist
	handler  *handlers.Handler
}

// Register sets up the voice feature with the bot.
func Register(deps Deps) *Feature {
	// Apply defaults if not set
	if deps.Config.DefaultVolume == 0 {
		deps.Config.DefaultVolume = DefaultConfig().DefaultVolume
	}
	if deps.Config.MaxQueueSize == 0 {
		deps.Config.MaxQueueSize = DefaultConfig().MaxQueueSize
	}
	if deps.Logger == nil {
		deps.Logger = logger.New()
	}

	// Create service with config
	svcConfig := service.Config{
		DefaultVolume: deps.Config.DefaultVolume,
		MaxQueueSize:  deps.Config.MaxQueueSize,
	}
	svc := service.New(deps.Bot, deps.Logger, svcConfig)

	// Create persistence handler
	persist := persistence.New(deps.Logger)

	// Create handler with service dependency
	handlerConfig := handlers.Config{
		YouTubeEnabled:    deps.Config.YouTubeEnabled,
		LocalFilesEnabled: deps.Config.LocalFilesEnabled,
	}
	handler := handlers.New(svc, deps.Logger, handlerConfig)

	f := &Feature{
		svc:     svc,
		persist: persist,
		handler: handler,
	}

	// Register handlers
	deps.Bot.SlashCommand("play", "Play audio from a URL or attachment", handler.HandlePlay,
		discord.StringOption("query", "URL, search query, or attachment", false),
	)
	deps.Bot.SlashCommand("skip", "Skip the current track", handler.HandleSkip)
	deps.Bot.SlashCommand("queue", "Show the current playback queue", handler.HandleQueue)

	// Register shutdown hook
	deps.Bot.OnShutdown(persist.Save)

	deps.Logger.Info(context.Background(), "voice feature registered", logger.Fields{
		"component":      "voice",
		"enabled":        deps.Config.Enabled,
		"max_queue_size": deps.Config.MaxQueueSize,
		"youtube":        deps.Config.YouTubeEnabled,
		"local_files":    deps.Config.LocalFilesEnabled,
	})

	return f
}

// Service returns the underlying service for external access if needed.
func (f *Feature) Service() service.VoiceService {
	return f.svc
}
