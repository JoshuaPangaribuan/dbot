// Package service provides core business logic for the voice feature.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// Config holds configuration for the voice service.
type Config struct {
	DefaultVolume int // Default volume (0-100)
	MaxQueueSize  int // Maximum tracks in queue
}

// VoiceService defines the interface for voice operations (DIP).
type VoiceService interface {
	// Connection management
	Connect(ctx context.Context, guildID, channelID string) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetChannelID() string
	GetGuildID() string

	// Playback control
	Play(ctx context.Context, track *Track) error
	Skip(ctx context.Context) error
	Stop(ctx context.Context) error
	IsPlaying() bool
	Current() *Track

	// Queue management
	Enqueue(ctx context.Context, track *Track) error
	Dequeue(ctx context.Context) *Track
	Queue() []*Track
	QueueSize() int
	ClearQueue(ctx context.Context)

	// User voice state
	GetUserVoiceChannel(ctx context.Context, guildID, userID string) (string, error)
}

// Track represents a playable audio track.
type Track struct {
	URL      string
	Title    string
	Duration int    // Duration in seconds
	Source   string // "youtube", "local", etc.
	UserID   string // User who requested the track
}

// impl manages voice connections and audio playback.
// MVP: This is a skeleton implementation - no actual audio streaming.
type impl struct {
	bot    *discord.Bot
	logger logger.Logger
	config Config

	// Simulated voice connection state (MVP: no actual Discord voice connection)
	guildID   string
	channelID string
	connMu    sync.RWMutex

	// Playback state
	playing bool
	current *Track
	queue   []*Track
	queueMu sync.RWMutex
}

// Ensure impl implements the VoiceService interface.
var _ VoiceService = (*impl)(nil)

// New creates a new voice service with the given configuration.
func New(bot *discord.Bot, logger logger.Logger, config Config) VoiceService {
	return &impl{
		bot:    bot,
		logger: logger,
		config: config,
		queue:  make([]*Track, 0, config.MaxQueueSize),
	}
}

// Connect joins a voice channel.
// MVP: Simulates connection - no actual Discord voice API call.
func (s *impl) Connect(ctx context.Context, guildID, channelID string) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	// Check if already connected to this channel
	if s.channelID == channelID {
		return nil // Already connected
	}

	s.guildID = guildID
	s.channelID = channelID

	s.logger.Info(ctx, "Voice connection simulated (MVP: no actual audio)", logger.Fields{
		"component":  "voice",
		"guild_id":   guildID,
		"channel_id": channelID,
		"mvp_note":   "Audio streaming not implemented - skeleton only",
	})

	return nil
}

// Disconnect leaves the current voice channel.
// MVP: Clears simulated connection state.
func (s *impl) Disconnect(ctx context.Context) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	s.guildID = ""
	s.channelID = ""

	s.logger.Info(ctx, "Voice disconnected", logger.Fields{
		"component": "voice",
	})

	return nil
}

// IsConnected returns true if connected to a voice channel.
func (s *impl) IsConnected() bool {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	return s.channelID != ""
}

// GetChannelID returns the current voice channel ID.
func (s *impl) GetChannelID() string {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	return s.channelID
}

// GetGuildID returns the current guild ID.
func (s *impl) GetGuildID() string {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	return s.guildID
}

// Enqueue adds a track to the playback queue.
func (s *impl) Enqueue(ctx context.Context, track *Track) error {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	if len(s.queue) >= s.config.MaxQueueSize {
		return fmt.Errorf("queue is full (max %d tracks)", s.config.MaxQueueSize)
	}

	s.queue = append(s.queue, track)

	s.logger.Debug(ctx, "Track added to queue", logger.Fields{
		"component":   "voice",
		"title":       track.Title,
		"queue_size":  len(s.queue),
		"queue_limit": s.config.MaxQueueSize,
	})

	return nil
}

// Dequeue removes and returns the next track from the queue.
func (s *impl) Dequeue(ctx context.Context) *Track {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()

	if len(s.queue) == 0 {
		return nil
	}

	track := s.queue[0]
	s.queue = s.queue[1:]
	return track
}

// Queue returns a copy of the current queue.
func (s *impl) Queue() []*Track {
	s.queueMu.RLock()
	defer s.queueMu.RUnlock()

	queue := make([]*Track, len(s.queue))
	copy(queue, s.queue)
	return queue
}

// QueueSize returns the current queue size.
func (s *impl) QueueSize() int {
	s.queueMu.RLock()
	defer s.queueMu.RUnlock()
	return len(s.queue)
}

// Current returns the currently playing track.
func (s *impl) Current() *Track {
	return s.current
}

// IsPlaying returns true if currently playing audio.
func (s *impl) IsPlaying() bool {
	return s.playing
}

// Skip stops the current track and advances to the next.
func (s *impl) Skip(ctx context.Context) error {
	if !s.playing {
		return fmt.Errorf("not playing")
	}

	s.playing = false
	s.current = nil

	s.logger.Debug(ctx, "Track skipped", logger.Fields{
		"component": "voice",
	})

	return nil
}

// Play starts playing a track.
// MVP: Simulates playback - no actual audio streaming.
func (s *impl) Play(ctx context.Context, track *Track) error {
	s.current = track
	s.playing = true

	s.logger.Info(ctx, "Playback simulated (MVP: no actual audio)", logger.Fields{
		"component": "voice",
		"title":     track.Title,
		"url":       track.URL,
		"source":    track.Source,
		"mvp_note":  "Audio streaming not implemented - skeleton only",
	})

	return nil
}

// Stop stops the current playback.
func (s *impl) Stop(ctx context.Context) error {
	s.playing = false
	s.current = nil

	return nil
}

// ClearQueue clears all tracks from the queue.
func (s *impl) ClearQueue(ctx context.Context) {
	s.queueMu.Lock()
	defer s.queueMu.Unlock()
	s.queue = s.queue[:0]
}

// GetUserVoiceChannel returns the voice channel ID for a user in a guild.
func (s *impl) GetUserVoiceChannel(ctx context.Context, guildID, userID string) (string, error) {
	var session *discordgo.Session = s.bot.Session()

	// Get guild voice states
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return "", fmt.Errorf("failed to get guild: %w", err)
	}

	// Find user's voice state
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID, nil
		}
	}

	return "", fmt.Errorf("user not in voice channel")
}
