// Package persistence handles saving and loading of voice state for the voice feature.
package persistence

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Persist handles saving and loading of voice state.
type Persist struct {
	logger logger.Logger
}

// New creates a new persistence handler.
func New(log logger.Logger) *Persist {
	return &Persist{
		logger: log,
	}
}

// Save persists voice state to storage.
func (p *Persist) Save(ctx context.Context) error {
	p.logger.Info(ctx, "voice feature shutting down", logger.Fields{
		"component": "voice",
	})
	// TODO: Persist queue state to database
	return nil
}

// Load loads voice state from storage.
func (p *Persist) Load(ctx context.Context) error {
	// TODO: Load queue state from database
	p.logger.Info(ctx, "voice feature loaded", logger.Fields{
		"component": "voice",
	})
	return nil
}
