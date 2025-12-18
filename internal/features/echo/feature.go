package echo

import (
	"context"

	v1 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v1"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// Deps contains dependencies for the echo feature.
type Deps struct {
	Logger logger.Logger
}

// Feature implements v1.Feature for the echo module.
type Feature struct {
	logger logger.Logger
	bot    *v1.Bot
}

func NewFeature(deps Deps) *Feature {
	log := deps.Logger
	if log == nil {
		log = logger.New()
	}
	return &Feature{logger: log}
}

func (f *Feature) Name() string { return "echo" }

func (f *Feature) Init(b *v1.Bot) error {
	f.bot = b
	f.logger.Info(context.Background(), "feature initialized", logger.Fields{
		"component": "echo",
	})
	return nil
}

func (f *Feature) Shutdown(ctx context.Context) error {
	f.logger.Info(ctx, "feature shutdown", logger.Fields{
		"component": "echo",
	})
	return nil
}

func (f *Feature) Subscriptions() []*v1.Subscription {
	return []*v1.Subscription{
		{
			Type:    v1.EventTypeCommand,
			Handler: v1.CommandFilter("echo", f.handleEcho),
		},
	}
}

// Commands implements v1.CommandProvider.
func (f *Feature) Commands() []*v1.Command {
	return []*v1.Command{
		v1.NewSlashCommand("echo", "Echoes back your text").
			WithOptions(&discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to echo back",
				Required:    true,
			}),
	}
}

func (f *Feature) handleEcho(ctx context.Context, e *v1.Event) error {
	ic, ok := e.Data().(*discordgo.InteractionCreate)
	if !ok {
		return nil
	}

	cmdCtx := v1.NewCommandContext(
		v1.NewContext(e, f.bot.State()),
		ic,
		nil,
	)

	text := cmdCtx.Option("text").String()
	_, err := cmdCtx.Reply().
		NoMentions().
		Content(text).
		Send()
	return err
}
