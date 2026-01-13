// Package service provides core business logic for the ask feature.
package service

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/langchain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Service defines the interface for AI completion operations (DIP).
type Service interface {
	Complete(ctx context.Context, prompt string) (string, error)
	Close() error
}

// impl manages AI completion operations using langchain.
type impl struct {
	svc    langchain.Service
	config Config
	logger logger.Logger
}

// Config holds configuration for the completion service.
type Config struct {
	MaxResponseLength int
}

// Ensure impl implements the Service interface.
var _ Service = (*impl)(nil)

// New creates a new completion service with the given langchain service.
func New(svc langchain.Service, cfg Config, log logger.Logger) Service {
	if cfg.MaxResponseLength == 0 {
		cfg.MaxResponseLength = 1900 // Discord message limit with safety margin
	}
	return &impl{
		svc:    svc,
		config: cfg,
		logger: log,
	}
}

// Complete performs an AI completion for the given prompt.
func (s *impl) Complete(ctx context.Context, prompt string) (string, error) {
	s.logger.Info(ctx, "starting completion", logger.Fields{
		"prompt_length": len(prompt),
	})

	response, err := s.svc.Complete(ctx, prompt)
	if err != nil {
		s.logger.Error(ctx, "completion failed", logger.Fields{
			"error":         err,
			"error_type":    fmt.Sprintf("%T", err),
			"prompt_length": len(prompt),
		})
		return "", err
	}

	s.logger.Info(ctx, "completion succeeded", logger.Fields{
		"response_length":  len(response),
		"response_preview": truncateForLog(response, 100),
	})

	// Truncate if too long
	if len(response) > s.config.MaxResponseLength {
		response = response[:s.config.MaxResponseLength] + "..."
	}

	return response, nil
}

// Close cleans up the underlying langchain service.
func (s *impl) Close() error {
	return s.svc.Close()
}

// truncateForLog truncates a string for logging purposes.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
