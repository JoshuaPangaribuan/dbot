package v2

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
)

// Feature represents a self-contained bot feature with lifecycle hooks.
// Features register event subscriptions and can manage their own state.
type Feature interface {
	// Name returns a unique identifier for the feature.
	Name() string

	// Init is called when the feature is registered with the bot.
	// Use this to set up resources, connections, or state.
	Init(b *Bot) error

	// Shutdown is called when the bot is closing.
	// Use this to clean up resources. The context may have a deadline.
	Shutdown(ctx context.Context) error

	// Subscriptions returns the event subscriptions for this feature.
	Subscriptions() []*Subscription
}

// CommandProvider is an optional interface that features can implement
// to provide slash commands that should be registered with Discord.
// If a feature implements this interface, its commands will be included
// in the bot's command sync during Open().
type CommandProvider interface {
	// Commands returns the slash commands this feature provides.
	Commands() []*domain.Command
}

// CommandFilter returns an EventHandler that only calls the given handler
// for commands matching the specified name.
func CommandFilter(name string, handler EventHandler) EventHandler {
	return func(ctx context.Context, e *domain.Event) error {
		if e.Type() != domain.EventTypeCommand {
			return nil
		}
		if e.Interaction == nil || e.Interaction.CommandData == nil {
			return nil
		}
		if e.Interaction.CommandData.Name != name {
			return nil
		}
		return handler(ctx, e)
	}
}

// ComponentFilter returns an EventHandler that only calls the given handler
// for component interactions matching the specified custom ID.
func ComponentFilter(customID string, handler EventHandler) EventHandler {
	return func(ctx context.Context, e *domain.Event) error {
		if e.Type() != domain.EventTypeComponent {
			return nil
		}
		if e.Interaction == nil || e.Interaction.ComponentData == nil {
			return nil
		}
		if e.Interaction.ComponentData.CustomID != customID {
			return nil
		}
		return handler(ctx, e)
	}
}

// ModalFilter returns an EventHandler that only calls the given handler
// for modal submissions matching the specified custom ID.
func ModalFilter(customID string, handler EventHandler) EventHandler {
	return func(ctx context.Context, e *domain.Event) error {
		if e.Type() != domain.EventTypeModal {
			return nil
		}
		if e.Interaction == nil || e.Interaction.ModalData == nil {
			return nil
		}
		if e.Interaction.ModalData.CustomID != customID {
			return nil
		}
		return handler(ctx, e)
	}
}
