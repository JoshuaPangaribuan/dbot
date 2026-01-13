// Package handlers provides Discord event handlers for the voice feature.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaPangaribuan/dbot/internal/features/voice/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Handler manages Discord event handlers for the voice feature.
type Handler struct {
	svc    service.VoiceService
	logger logger.Logger
	config Config
}

// Config holds configuration for the voice handlers.
type Config struct {
	YouTubeEnabled    bool
	LocalFilesEnabled bool
}

// New creates a new handler with service dependency injected.
func New(svc service.VoiceService, logger logger.Logger, cfg Config) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
		config: cfg,
	}
}

const mvpNote = "\n\n*Note: This is an MVP skeleton - audio streaming is not yet implemented.*"

// HandlePlay handles the /play command.
func (h *Handler) HandlePlay(ctx context.Context, cmdCtx *discord.CommandContext) error {
	query := cmdCtx.Option("query").String()

	if query == "" {
		_, err := cmdCtx.Reply().Content("Please provide a URL or search query.").Send()
		return err
	}

	// Get user's voice channel
	guildID := cmdCtx.GuildID()
	userID := cmdCtx.UserID()

	voiceChannelID, err := h.svc.GetUserVoiceChannel(ctx, guildID, userID)
	if err != nil {
		_, err := cmdCtx.Reply().Content("You must be in a voice channel to use this command.").Send()
		return err
	}

	// "Connect" to voice channel (MVP: simulated connection)
	if !h.svc.IsConnected() || h.svc.GetChannelID() != voiceChannelID {
		if err := h.svc.Connect(ctx, guildID, voiceChannelID); err != nil {
			h.logger.Error(ctx, "Failed to connect to voice channel", logger.Fields{"error": err})
			_, err := cmdCtx.Reply().Content("Failed to join voice channel.").Send()
			return err
		}
	}

	// Determine if this is a URL or search query
	var track *service.Track
	if isYouTubeURL(query) {
		if !h.config.YouTubeEnabled {
			_, err := cmdCtx.Reply().Content("YouTube playback is disabled.").Send()
			return err
		}
		track = &service.Track{
			URL:    query,
			Title:  "YouTube Track",
			Source: "youtube",
			UserID: userID,
		}
	} else if isLocalFile(query) {
		if !h.config.LocalFilesEnabled {
			_, err := cmdCtx.Reply().Content("Local file playback is disabled.").Send()
			return err
		}
		track = &service.Track{
			URL:    query,
			Title:  "Local File",
			Source: "local",
			UserID: userID,
		}
	} else {
		track = &service.Track{
			URL:    query,
			Title:  query,
			Source: "direct",
			UserID: userID,
		}
	}

	// Add to queue or play immediately
	queueSize := h.svc.QueueSize()
	if h.svc.IsPlaying() {
		if err := h.svc.Enqueue(ctx, track); err != nil {
			_, err := cmdCtx.Reply().Content(fmt.Sprintf("Failed to add to queue: %s", err.Error())).Send()
			return err
		}
		_, err := cmdCtx.Reply().Content(fmt.Sprintf("Added **%s** to queue (position %d)%s", track.Title, queueSize+1, mvpNote)).Send()
		return err
	}

	// Play immediately (MVP: simulated)
	if err := h.svc.Play(ctx, track); err != nil {
		h.logger.Error(ctx, "Failed to play track", logger.Fields{"error": err})
		_, err := cmdCtx.Reply().Content(fmt.Sprintf("Failed to play track: %s", err.Error())).Send()
		return err
	}

	_, err = cmdCtx.Reply().Content(fmt.Sprintf("Now playing: **%s**%s", track.Title, mvpNote)).Send()
	return err
}

// HandleSkip handles the /skip command.
func (h *Handler) HandleSkip(ctx context.Context, cmdCtx *discord.CommandContext) error {
	if !h.svc.IsPlaying() {
		_, err := cmdCtx.Reply().Content("Nothing is playing.").Send()
		return err
	}

	current := h.svc.Current()
	if err := h.svc.Skip(ctx); err != nil {
		h.logger.Error(ctx, "Failed to skip track", logger.Fields{"error": err})
		_, err := cmdCtx.Reply().Content("Failed to skip track.").Send()
		return err
	}

	// Try to play next track
	next := h.svc.Dequeue(ctx)
	if next != nil {
		if err := h.svc.Play(ctx, next); err != nil {
			h.logger.Error(ctx, "Failed to play next track", logger.Fields{"error": err})
			_, err := cmdCtx.Reply().Content(fmt.Sprintf("Skipped **%s**. Failed to play next track.", current.Title)).Send()
			return err
		}
		_, err := cmdCtx.Reply().Content(fmt.Sprintf("Skipped **%s**. Now playing: **%s**%s", current.Title, next.Title, mvpNote)).Send()
		return err
	}

	_, err := cmdCtx.Reply().Content(fmt.Sprintf("Skipped **%s**. Queue is empty.", current.Title)).Send()
	return err
}

// HandleQueue handles the /queue command.
func (h *Handler) HandleQueue(ctx context.Context, cmdCtx *discord.CommandContext) error {
	queue := h.svc.Queue()
	current := h.svc.Current()

	if current == nil && len(queue) == 0 {
		_, err := cmdCtx.Reply().Content("The queue is empty.").Send()
		return err
	}

	// Build queue display
	var content strings.Builder
	if current != nil {
		fmt.Fprintf(&content, "**Now Playing:** %s\n\n", current.Title)
	}

	if len(queue) > 0 {
		content.WriteString("**Up Next:**\n")
		for i, track := range queue {
			fmt.Fprintf(&content, "%d. %s\n", i+1, track.Title)
		}
	}

	content.WriteString(mvpNote)

	_, err := cmdCtx.Reply().Content(content.String()).Send()
	return err
}

// Helper functions

func isYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

func isLocalFile(url string) bool {
	return strings.Contains(url, "cdn.discordapp.com") || strings.Contains(url, "media.discordapp.net")
}
