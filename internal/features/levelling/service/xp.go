// Package service provides core business logic for the levelling feature.
package service

import (
	"context"
	"sync"
)

// Config holds configuration for the levelling service.
type Config struct {
	XPPerMessage int   // XP gained per message
	Thresholds   []int // XP needed for each level
}

// Service defines the interface for XP and level operations (DIP).
type Service interface {
	AwardXP(ctx context.Context, userID string) (oldLevel, newLevel int)
	GetXP(ctx context.Context, userID string) int
	SetXP(ctx context.Context, userID string, xp int)
	GetLevelForXP(ctx context.Context, xp int) int
	GetNextThreshold(ctx context.Context, level int) int
	GetAllUsers(ctx context.Context) map[string]int
	UserCount(ctx context.Context) int
}

// impl manages XP tracking and level calculations.
type impl struct {
	config Config
	mu     sync.RWMutex
	xp     map[string]int // userID -> XP
}

// Ensure impl implements the Service interface.
var _ Service = (*impl)(nil)

// New creates a new levelling service with the given configuration.
func New(cfg Config) Service {
	return &impl{
		config: cfg,
		xp:     make(map[string]int),
	}
}

// AwardXP adds XP for a user and returns the old and new levels.
func (s *impl) AwardXP(ctx context.Context, userID string) (oldLevel, newLevel int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldXP := s.xp[userID]
	newXP := oldXP + s.config.XPPerMessage
	s.xp[userID] = newXP

	oldLevel = s.getLevelForXP(oldXP)
	newLevel = s.getLevelForXP(newXP)
	return
}

// GetXP returns the current XP for a user.
func (s *impl) GetXP(ctx context.Context, userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.xp[userID]
}

// SetXP sets the XP for a user (used for loading from persistence).
func (s *impl) SetXP(ctx context.Context, userID string, xp int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.xp[userID] = xp
}

// GetLevelForXP returns the level for a given XP amount.
func (s *impl) GetLevelForXP(ctx context.Context, xp int) int {
	level := 0
	for _, threshold := range s.config.Thresholds {
		if xp >= threshold {
			level++
		} else {
			break
		}
	}
	return level
}

// GetNextThreshold returns XP needed for next level, or 0 if max level.
func (s *impl) GetNextThreshold(ctx context.Context, level int) int {
	if level < len(s.config.Thresholds) {
		return s.config.Thresholds[level]
	}
	return 0
}

// GetAllUsers returns all user IDs with their XP (for persistence).
func (s *impl) GetAllUsers(ctx context.Context) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int, len(s.xp))
	for userID, xp := range s.xp {
		result[userID] = xp
		// nolint:mapsloop // Using explicit loop for compatibility with Go <1.23
	}
	return result
}

// UserCount returns the number of users tracked.
func (s *impl) UserCount(ctx context.Context) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.xp)
}

// getLevelForXP is the internal implementation without locking.
func (s *impl) getLevelForXP(xp int) int {
	level := 0
	for _, threshold := range s.config.Thresholds {
		if xp >= threshold {
			level++
		} else {
			break
		}
	}
	return level
}
