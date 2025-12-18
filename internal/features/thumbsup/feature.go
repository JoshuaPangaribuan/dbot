package thumbsup

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// Deps contains dependencies for the thumbsup feature.
type Deps struct {
	Logger logger.Logger
}

// Feature implements discord.Feature and reacts with 👍 to every message.
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

func (f *Feature) Name() string { return "thumbsup" }

func (f *Feature) Init(b *discord.Bot) error {
	f.bot = b
	f.logger.Info(context.Background(), "feature initialized", logger.Fields{
		"component": "thumbsup",
	})
	return nil
}

func (f *Feature) Shutdown(ctx context.Context) error {
	f.logger.Info(ctx, "feature shutdown", logger.Fields{
		"component": "thumbsup",
	})
	return nil
}

func (f *Feature) Subscriptions() []*discord.Subscription {
	return []*discord.Subscription{
		{
			Type:    discord.EventTypeMessage,
			Handler: f.handleMessage,
		},
	}
}

func (f *Feature) handleMessage(ctx context.Context, e *discord.Event) error {
	m, ok := e.Data().(*discordgo.MessageCreate)
	if !ok || m == nil {
		return nil
	}

	s := e.Session()
	if s == nil || m.ChannelID == "" || m.ID == "" {
		return nil
	}

	if err := s.MessageReactionAdd(m.ChannelID, m.ID, "👍"); err != nil {
		// Don't fail the whole event pipeline if the bot lacks permissions.
		f.logger.Debug(ctx, "failed to add thumbs up reaction", logger.Fields{
			"component": "thumbsup",
			"channel":   m.ChannelID,
			"message":   m.ID,
			"error":     err.Error(),
		})
		return nil
	}

	return nil
}
