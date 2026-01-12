// Package handlers provides Discord event handlers for the levelling feature.
package handlers

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/features/levelling/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
)

// Handler manages Discord event handlers for the levelling feature.
// It depends on service.Service interface (DIP).
type Handler struct {
	svc        service.Service
	announceUp bool
}

// New creates a new handler with service dependency injected.
func New(svc service.Service, announceUp bool) *Handler {
	return &Handler{
		svc:        svc,
		announceUp: announceUp,
	}
}

// OnMessage handles incoming messages to award XP and check for level-ups.
func (h *Handler) OnMessage(ctx context.Context, msg *discord.MessageContext) error {
	userID := msg.UserID()
	if userID == "" {
		return nil
	}

	// Award XP and get level change
	oldLevel, newLevel := h.svc.AwardXP(userID)

	// Check for level-up
	if newLevel > oldLevel && h.announceUp {
		session := msg.Session()
		if session != nil {
			_, _ = session.ChannelMessageSend(msg.ChannelID(),
				fmt.Sprintf("🎉 <@%s> reached **Level %d**!", userID, newLevel))
		}
	}

	return nil
}

// CheckLevel handles the /level command to display user level information.
func (h *Handler) CheckLevel(ctx context.Context, cmd *discord.CommandContext) error {
	// Check if user option was provided
	targetUserID := cmd.Option("user").String()
	if targetUserID == "" {
		targetUserID = cmd.UserID()
	}

	xp := h.svc.GetXP(targetUserID)
	level := h.svc.GetLevelForXP(xp)
	nextThreshold := h.svc.GetNextThreshold(level)

	var message string
	if nextThreshold > 0 {
		needed := nextThreshold - xp
		message = fmt.Sprintf("📊 <@%s> is **Level %d** (%d XP)\n%d XP needed for Level %d",
			targetUserID, level, xp, needed, level+1)
	} else {
		message = fmt.Sprintf("📊 <@%s> is **Level %d** (%d XP) — Max level reached! 🏆",
			targetUserID, level, xp)
	}

	_, err := cmd.Reply().Content(message).Send()
	return err
}
