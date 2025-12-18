package pingpong

import (
	"context"
	"regexp"
	"strings"

	"github.com/JoshuaPangaribuan/dbot/internal/features/pingpong/service"
	v1 "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v1"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/bwmarrin/discordgo"
)

// mentionPattern matches @everyone, @here, and role mentions
var mentionPattern = regexp.MustCompile(`@(everyone|here|&\d+)`)

// Deps contains dependencies for the pingpong feature.
type Deps struct {
	Logger logger.Logger
}

// Feature implements v1.Feature for the pingpong module.
// This is the new recommended way to register pingpong functionality.
type Feature struct {
	logger  logger.Logger
	service service.Service
	bot     *v1.Bot
}

// NewFeature creates a new pingpong Feature.
func NewFeature(deps Deps) *Feature {
	log := deps.Logger
	if log == nil {
		log = logger.New()
	}

	return &Feature{
		logger:  log,
		service: service.New(),
	}
}

// Name returns the feature name.
func (f *Feature) Name() string {
	return "pingpong"
}

// Init initializes the feature with the bot.
func (f *Feature) Init(b *v1.Bot) error {
	f.bot = b
	f.logger.Info(context.Background(), "feature initialized", logger.Fields{
		"component": "pingpong",
	})
	return nil
}

// Shutdown cleans up the feature.
func (f *Feature) Shutdown(ctx context.Context) error {
	f.logger.Info(ctx, "feature shutdown", logger.Fields{
		"component": "pingpong",
	})
	return nil
}

// Subscriptions returns the event subscriptions for this feature.
func (f *Feature) Subscriptions() []*v1.Subscription {
	return []*v1.Subscription{
		// /ping command
		{
			Type:     v1.EventTypeCommand,
			Handler:  v1.CommandFilter("ping", f.handlePing),
			Priority: 0,
		},
		// !echo message handler
		{
			Type:     v1.EventTypeMessage,
			Handler:  f.handleEcho,
			Priority: 0,
		},
	}
}

// Commands implements v1.CommandProvider.
// Returns the slash commands this feature provides.
func (f *Feature) Commands() []*v1.Command {
	return []*v1.Command{
		v1.NewSlashCommand("ping", "Replies with pong"),
	}
}

// handlePing handles the /ping slash command.
func (f *Feature) handlePing(ctx context.Context, e *v1.Event) error {
	ic, ok := e.Data().(*discordgo.InteractionCreate)
	if !ok {
		return nil
	}

	// Create a CommandContext for the reply builder
	cmdCtx := v1.NewCommandContext(
		v1.NewContext(e, f.bot.State()),
		ic,
		nil,
	)

	_, err := cmdCtx.Reply().Content(f.service.Pong(ctx)).Send()
	return err
}

// handleEcho handles the !echo message command.
func (f *Feature) handleEcho(ctx context.Context, e *v1.Event) error {
	m, ok := e.Data().(*discordgo.MessageCreate)
	if !ok {
		return nil
	}

	session := e.Session()
	if session == nil {
		return nil
	}

	// Support both:
	//   - !echo hello
	//   - @Bot !echo hello (useful when message content intent isn't available)
	content, ok := parseBangCommand(stripLeadingBotMention(session, m.Content), "echo")
	if !ok {
		return nil
	}

	content = strings.TrimSpace(content)
	if content == "" {
		_, err := session.ChannelMessageSend(m.ChannelID, "Usage: !echo <text>")
		return err
	}

	// Sanitize mentions to prevent injection (zero-width space after @)
	content = mentionPattern.ReplaceAllString(content, "@\u200b$1")

	logFields := logger.Fields{
		"component": "pingpong",
		"handler":   "echo",
		"channel":   m.ChannelID,
	}
	if m.Author != nil {
		logFields["user"] = m.Author.Username
		logFields["user_id"] = m.Author.ID
	}
	f.logger.Info(ctx, "echoing message", logFields)

	_, err := session.ChannelMessageSend(m.ChannelID, content)
	if err != nil {
		f.logger.Error(ctx, "failed to send echo message", logger.Fields{
			"component": "pingpong",
			"handler":   "echo",
			"error":     err.Error(),
		})
	}
	return err
}

func stripLeadingBotMention(s *discordgo.Session, content string) string {
	if s == nil || s.State == nil || s.State.User == nil || s.State.User.ID == "" {
		return content
	}

	trimmed := strings.TrimSpace(content)
	botID := s.State.User.ID

	// Discord user mention formats: <@id> and <@!id>
	mention := "<@" + botID + ">"
	if strings.HasPrefix(trimmed, mention) {
		return strings.TrimSpace(trimmed[len(mention):])
	}

	mentionNick := "<@!" + botID + ">"
	if strings.HasPrefix(trimmed, mentionNick) {
		return strings.TrimSpace(trimmed[len(mentionNick):])
	}

	return content
}

func parseBangCommand(content string, command string) (args string, ok bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", false
	}

	cmd := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(command, "!")))
	if cmd == "" {
		return "", false
	}
	prefix := "!" + cmd

	lower := strings.ToLower(content)
	if lower == prefix {
		return "", true
	}
	if strings.HasPrefix(lower, prefix+" ") {
		return strings.TrimSpace(content[len(prefix):]), true
	}
	return "", false
}
