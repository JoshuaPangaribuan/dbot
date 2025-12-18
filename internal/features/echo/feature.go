package echo

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// Deps contains dependencies for the echo feature.
type Deps struct {
	Logger logger.Logger
}

// Feature implements discord.Feature for the echo module.
type Feature struct {
	logger logger.Logger
	bot    *discord.Bot
}

func NewFeature(deps Deps) *Feature {
	log := deps.Logger
	if log == nil {
		log = logger.New()
	}
	return &Feature{logger: log}
}

func (f *Feature) Name() string { return "echo" }

func (f *Feature) Init(b *discord.Bot) error {
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

func (f *Feature) Subscriptions() []*discord.Subscription {
	return []*discord.Subscription{
		{
			Type:    discord.EventTypeCommand,
			Handler: discord.CommandFilter("echo", f.handleEcho),
		},
	}
}

// Commands implements discord.CommandProvider.
func (f *Feature) Commands() []*discord.Command {
	return []*discord.Command{
		discord.NewSlashCommand("echo", "Echoes back your text").
			WithOptions(&discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "Text to echo back",
				Required:    true,
			}),
	}
}

func (f *Feature) handleEcho(ctx context.Context, e *discord.Event) error {
	ic, ok := e.Data().(*discordgo.InteractionCreate)
	if !ok {
		return nil
	}

	cmdCtx := discord.NewCommandContext(
		discord.NewContext(e, f.bot.State()),
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
