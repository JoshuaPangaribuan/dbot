package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/config"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

// EventPermissionChecker defines the interface for checking user permissions in EventBus handlers.
type EventPermissionChecker interface {
	HasPermissionForEvent(ctx context.Context, e *discord.Event, permission string) bool
}

// HybridPermissionChecker checks permissions from config first, then Discord.
type HybridPermissionChecker struct {
	cfg config.Config
}

// NewHybridPermissionChecker creates a new hybrid permission checker.
func NewHybridPermissionChecker(cfg config.Config) *HybridPermissionChecker {
	return &HybridPermissionChecker{cfg: cfg}
}

// HasPermissionForEvent checks if the user has the specified permission for an event.
func (h *HybridPermissionChecker) HasPermissionForEvent(ctx context.Context, e *discord.Event, permission string) bool {
	var userID string
	var member *discordgo.Member
	var guildID string

	switch data := e.Data().(type) {
	case *discordgo.InteractionCreate:
		if data.Member != nil && data.Member.User != nil {
			userID = data.Member.User.ID
			member = data.Member
		} else if data.User != nil {
			userID = data.User.ID
		}
		guildID = data.GuildID
	case *discordgo.MessageCreate:
		if data.Author != nil {
			userID = data.Author.ID
		}
		member = data.Member
		guildID = data.GuildID
	}

	if userID == "" {
		return false
	}

	// Check config first
	key := fmt.Sprintf("discord.permissions.%s", userID)
	perms := h.cfg.GetSlice(key)
	for _, p := range perms {
		if pStr, ok := p.(string); ok {
			if strings.EqualFold(pStr, permission) || pStr == "*" {
				return true
			}
		}
	}

	// Fallback: check Discord roles
	if member == nil || guildID == "" {
		return false
	}

	session := e.Session()
	if session == nil {
		return false
	}

	roles, err := session.GuildRoles(guildID)
	if err != nil {
		return false
	}

	roleMap := make(map[string]string)
	for _, r := range roles {
		roleMap[r.ID] = strings.ToLower(r.Name)
	}

	for _, roleID := range member.Roles {
		roleName := roleMap[roleID]
		if strings.EqualFold(roleName, permission) {
			return true
		}
	}

	return false
}

// RequireEventPermission returns middleware for EventBus that checks for a specific permission.
func RequireEventPermission(checker EventPermissionChecker, permission string) discord.MiddlewareFunc {
	return func(ctx context.Context, e *discord.Event, next func() error) error {
		if !checker.HasPermissionForEvent(ctx, e, permission) {
			// For interactions, send ephemeral error response
			if ic, ok := e.Data().(*discordgo.InteractionCreate); ok {
				_ = e.Session().InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You don't have permission to use this command.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
			return nil
		}
		return next()
	}
}

// RequireAnyEventPermission returns middleware for EventBus that checks if user has any of the permissions.
func RequireAnyEventPermission(checker EventPermissionChecker, permissions ...string) discord.MiddlewareFunc {
	return func(ctx context.Context, e *discord.Event, next func() error) error {
		for _, perm := range permissions {
			if checker.HasPermissionForEvent(ctx, e, perm) {
				return next()
			}
		}
		// For interactions, send ephemeral error response
		if ic, ok := e.Data().(*discordgo.InteractionCreate); ok {
			_ = e.Session().InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You don't have permission to use this command.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return nil
	}
}
