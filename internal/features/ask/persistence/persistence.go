// Package persistence handles saving and loading of ask data.
package persistence

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/features/ask/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Persist handles saving and loading of ask feature data.
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

// Save persists ask data to storage.
func (p *Persist) Save(ctx context.Context) error {
	p.logger.Info(ctx, "ask feature shutting down", logger.Fields{
		"component": "ask",
	})
	// TODO: Persist conversation history or statistics if needed
	return nil
}

// Load loads ask data from storage.
func (p *Persist) Load(ctx context.Context) error {
	// TODO: Load conversation history or statistics if needed
	p.logger.Info(ctx, "ask feature loaded", logger.Fields{
		"component": "ask",
	})
	return nil
}
