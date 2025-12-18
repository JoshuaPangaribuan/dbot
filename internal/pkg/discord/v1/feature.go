package v1

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Feature represents a self-contained bot feature with lifecycle hooks.
// Features register event subscriptions and can manage their own state.
//
// Example implementation:
//
//	type MyFeature struct {
//	    log logger.Logger
//	}
//
//	func (f *MyFeature) Name() string { return "my-feature" }
//
//	func (f *MyFeature) Init(b *Bot) error {
//	    f.log.Info(context.Background(), "feature initialized", nil)
//	    return nil
//	}
//
//	func (f *MyFeature) Shutdown(ctx context.Context) error {
//	    return nil
//	}
//
//	func (f *MyFeature) Subscriptions() []Subscription {
//	    return []Subscription{
//	        {Type: EventTypeCommand, Handler: f.handleCommand},
//	    }
//	}
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
//
// Example:
//
//	func (f *MyFeature) Commands() []*Command {
//	    return []*Command{
//	        NewSlashCommand("hello", "Says hello"),
//	    }
//	}
type CommandProvider interface {
	// Commands returns the slash commands this feature provides.
	Commands() []*Command
}

// CommandFilter returns an EventHandler that only calls the given handler
// for commands matching the specified name.
func CommandFilter(name string, handler EventHandler) EventHandler {
	return func(ctx context.Context, e *Event) error {
		ic, ok := e.Data().(*discordgo.InteractionCreate)
		if !ok {
			return nil
		}
		if ic.Type != discordgo.InteractionApplicationCommand {
			return nil
		}
		data := ic.ApplicationCommandData()
		if data.Name != name {
			return nil
		}
		return handler(ctx, e)
	}
}

// ComponentFilter returns an EventHandler that only calls the given handler
// for component interactions matching the specified custom ID pattern.
func ComponentFilter(customID string, handler EventHandler) EventHandler {
	return func(ctx context.Context, e *Event) error {
		ic, ok := e.Data().(*discordgo.InteractionCreate)
		if !ok {
			return nil
		}
		if ic.Type != discordgo.InteractionMessageComponent {
			return nil
		}
		data := ic.MessageComponentData()
		if data.CustomID != customID {
			return nil
		}
		return handler(ctx, e)
	}
}

