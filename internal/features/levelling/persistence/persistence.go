// Package persistence handles saving and loading of XP data for the levelling feature.
package persistence

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Persist handles saving and loading of XP data.
type Persist struct {
	svc    service.Service
	logger logger.Logger
}

// New creates a new persistence handler.
func New(svc service.Service, log logger.Logger) *Persist {
	return &Persist{
		svc:    svc,
		logger: log,
	}
}

// Save persists XP data to storage.
func (p *Persist) Save(ctx context.Context) error {
	p.logger.Info(ctx, "levelling feature shutting down", logger.Fields{
		"component": "levelling",
		"users":     p.svc.UserCount(ctx),
	})
	// TODO: Persist to database
	return nil
}

// Load loads XP data from storage.
func (p *Persist) Load(ctx context.Context) error {
	// TODO: Load from database
	p.logger.Info(ctx, "levelling feature loaded", logger.Fields{
		"component": "levelling",
		"users":     0,
	})
	return nil
}
