// Package handlers provides Discord event handlers for the ask feature.
package handlers

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/features/ask/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/reply"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Handler manages Discord event handlers for the ask feature.
// It depends on service.Service interface (DIP).
type Handler struct {
	svc    service.Service
	logger logger.Logger
}

// New creates a new handler with service dependency injected.
func New(svc service.Service, log logger.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: log,
	}
}

// HandleAsk handles the /ask command.
func (h *Handler) HandleAsk(ctx context.Context, cmd *discord.CommandContext) error {
	prompt := cmd.Option("prompt").String()

	h.logger.Info(ctx, "ask command received", logger.Fields{
		"prompt":        prompt,
		"prompt_length": len(prompt),
	})

	// Send initial response
	sent, err := cmd.Reply().Content("Thinking...").Send()
	if err != nil {
		return fmt.Errorf("send initial response: %w", err)
	}

	return h.completeAndRespond(ctx, sent, prompt)
}

// completeAndRespond calls the AI and updates the message with the response.
func (h *Handler) completeAndRespond(ctx context.Context, sent *reply.SentMessage, prompt string) error {
	response, err := h.svc.Complete(ctx, prompt)
	if err != nil {
		return h.handleError(ctx, sent, err)
	}

	return sent.Edit(response)
}

// handleError handles errors by updating the message with error info.
func (h *Handler) handleError(ctx context.Context, sent *reply.SentMessage, err error) error {
	errorMsg := fmt.Sprintf("Error: %s", err.Error())

	// Truncate error message if needed
	const maxLen = 2000
	if len(errorMsg) > maxLen {
		errorMsg = errorMsg[:maxLen-3] + "..."
	}

	if editErr := sent.Edit(errorMsg); editErr != nil {
		h.logger.Error(ctx, "failed to send error message", logger.Fields{"error": editErr})
		return editErr
	}
	return err
}
