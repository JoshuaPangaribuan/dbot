package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/config"
	v2 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
)

// EventPermissionChecker defines the interface for checking user permissions.
type EventPermissionChecker interface {
	HasPermissionForEvent(ctx context.Context, e *domain.Event, permission string) bool
}

// HybridPermissionChecker checks permissions from config first, then Discord roles.
type HybridPermissionChecker struct {
	cfg  config.Config
	rest port.RestClient
}

// NewHybridPermissionChecker creates a new hybrid permission checker.
func NewHybridPermissionChecker(cfg config.Config, rest port.RestClient) *HybridPermissionChecker {
	return &HybridPermissionChecker{cfg: cfg, rest: rest}
}

// HasPermissionForEvent checks if the user has the specified permission.
func (h *HybridPermissionChecker) HasPermissionForEvent(ctx context.Context, e *domain.Event, permission string) bool {
	userID := e.UserID()
	if userID.IsEmpty() {
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
	member := e.Member()
	guildID := e.GuildID()
	if member == nil || guildID.IsEmpty() || h.rest == nil {
		return false
	}

	roles, err := h.rest.GetGuildRoles(ctx, guildID)
	if err != nil {
		return false
	}

	roleMap := make(map[domain.RoleID]string)
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

// RequireEventPermission returns middleware that checks for a specific permission.
func RequireEventPermission(checker EventPermissionChecker, permission string, rest port.RestClient) v2.MiddlewareFunc {
	return func(ctx context.Context, e *domain.Event, next func() error) error {
		if !checker.HasPermissionForEvent(ctx, e, permission) {
			// For interactions, send ephemeral error response
			if e.Interaction != nil && rest != nil {
				resp := &domain.Response{
					Type: domain.ResponseTypeChannelMessage,
					Data: &domain.ResponseData{
						Content: "You don't have permission to use this command.",
						Flags:   domain.ResponseFlagsEphemeral,
					},
				}
				_ = rest.RespondInteraction(ctx, e.Interaction.ID, e.Interaction.Token, resp)
			}
			return nil
		}
		return next()
	}
}

// RequireAnyEventPermission returns middleware that checks if user has any of the permissions.
func RequireAnyEventPermission(checker EventPermissionChecker, rest port.RestClient, permissions ...string) v2.MiddlewareFunc {
	return func(ctx context.Context, e *domain.Event, next func() error) error {
		for _, perm := range permissions {
			if checker.HasPermissionForEvent(ctx, e, perm) {
				return next()
			}
		}
		// For interactions, send ephemeral error response
		if e.Interaction != nil && rest != nil {
			resp := &domain.Response{
				Type: domain.ResponseTypeChannelMessage,
				Data: &domain.ResponseData{
					Content: "You don't have permission to use this command.",
					Flags:   domain.ResponseFlagsEphemeral,
				},
			}
			_ = rest.RespondInteraction(ctx, e.Interaction.ID, e.Interaction.Token, resp)
		}
		return nil
	}
}
