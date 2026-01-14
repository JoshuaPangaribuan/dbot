package discord

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/reply"
	"github.com/bwmarrin/discordgo"
)

// Context provides common functionality for all event contexts.
// It wraps an Event and provides helper methods for accessing common data.
type Context struct {
	ctx       context.Context
	session   *discordgo.Session
	event     *Event
	state     *StateStore
	responded bool
}

// NewContext creates a new Context from an Event.
func NewContext(e *Event, state *StateStore) *Context {
	return &Context{
		ctx:     e.Context(),
		session: e.Session(),
		event:   e,
		state:   state,
	}
}

// Context returns the underlying context.Context.
func (c *Context) Context() context.Context { return c.ctx }

// Session returns the discordgo session.
func (c *Context) Session() *discordgo.Session { return c.session }

// Event returns the underlying Event.
func (c *Context) Event() *Event { return c.event }

// State returns the state store.
func (c *Context) State() *StateStore { return c.state }

// GuildID returns the guild ID where the event occurred, or empty for DMs.
func (c *Context) GuildID() string {
	switch data := c.event.Data().(type) {
	case *discordgo.MessageCreate:
		return data.GuildID
	case *discordgo.InteractionCreate:
		return data.GuildID
	case *discordgo.VoiceStateUpdate:
		return data.GuildID
	case *discordgo.MessageReactionAdd:
		return data.GuildID
	case *discordgo.MessageReactionRemove:
		return data.GuildID
	}
	return ""
}

// ChannelID returns the channel ID where the event occurred.
func (c *Context) ChannelID() string {
	switch data := c.event.Data().(type) {
	case *discordgo.MessageCreate:
		return data.ChannelID
	case *discordgo.InteractionCreate:
		return data.ChannelID
	case *discordgo.VoiceStateUpdate:
		return data.ChannelID
	case *discordgo.MessageReactionAdd:
		return data.ChannelID
	case *discordgo.MessageReactionRemove:
		return data.ChannelID
	}
	return ""
}

// UserID returns the user ID who triggered the event.
func (c *Context) UserID() string {
	switch data := c.event.Data().(type) {
	case *discordgo.MessageCreate:
		if data.Author != nil {
			return data.Author.ID
		}
	case *discordgo.InteractionCreate:
		if data.Member != nil && data.Member.User != nil {
			return data.Member.User.ID
		}
		if data.User != nil {
			return data.User.ID
		}
	case *discordgo.VoiceStateUpdate:
		return data.UserID
	case *discordgo.MessageReactionAdd:
		return data.UserID
	case *discordgo.MessageReactionRemove:
		return data.UserID
	}
	return ""
}

// User returns the user who triggered the event.
func (c *Context) User() *discordgo.User {
	switch data := c.event.Data().(type) {
	case *discordgo.MessageCreate:
		return data.Author
	case *discordgo.InteractionCreate:
		if data.Member != nil {
			return data.Member.User
		}
		return data.User
	case *discordgo.MessageReactionAdd:
		if data.Member != nil {
			return data.Member.User
		}
	}
	return nil
}

// Member returns the guild member who triggered the event (nil for DMs).
func (c *Context) Member() *discordgo.Member {
	switch data := c.event.Data().(type) {
	case *discordgo.MessageCreate:
		return data.Member
	case *discordgo.InteractionCreate:
		return data.Member
	case *discordgo.MessageReactionAdd:
		return data.Member
	}
	return nil
}

// Responded returns true if a response has been sent for this context.
func (c *Context) Responded() bool { return c.responded }

// MarkResponded marks the context as having sent a response.
func (c *Context) MarkResponded() { c.responded = true }

// CommandContext wraps Context with command-specific functionality.
type CommandContext struct {
	*Context
	interaction *discordgo.InteractionCreate
	command     *Command
}

// NewCommandContext creates a CommandContext from a Context and interaction data.
func NewCommandContext(c *Context, i *discordgo.InteractionCreate, cmd *Command) *CommandContext {
	return &CommandContext{
		Context:     c,
		interaction: i,
		command:     cmd,
	}
}

// Interaction returns the raw interaction create event.
func (cc *CommandContext) Interaction() *discordgo.InteractionCreate {
	return cc.interaction
}

// Command returns the command being executed.
func (cc *CommandContext) Command() *Command { return cc.command }

// Option returns the value of a command option by name.
func (cc *CommandContext) Option(name string) *cmdOptionValue {
	data := cc.interaction.ApplicationCommandData()
	for _, opt := range data.Options {
		if opt.Name == name {
			return &cmdOptionValue{opt: opt}
		}
	}
	return &cmdOptionValue{opt: nil}
}

// Defer sends a deferred response, indicating the bot is processing.
func (cc *CommandContext) Defer(ephemeral bool) error {
	if cc.responded {
		return nil
	}

	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := cc.session.InteractionRespond(cc.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
	if err != nil {
		return fmt.Errorf("defer interaction: %w", err)
	}
	cc.responded = true
	return nil
}

// Reply returns a new reply builder for this interaction.
func (cc *CommandContext) Reply() *reply.Builder {
	return reply.NewBuilder(cc.session, cc.interaction.Interaction, cc.responded, cc.MarkResponded)
}

// cmdOptionValue wraps a command option value for type-safe access.
// This is the implementation used by CommandContext.Option().
type cmdOptionValue struct {
	opt *discordgo.ApplicationCommandInteractionDataOption
}

// String returns the string value of the option.
func (ov *cmdOptionValue) String() string {
	if ov.opt == nil {
		return ""
	}
	if s, ok := ov.opt.Value.(string); ok {
		return s
	}
	return ""
}

// Int returns the integer value of the option.
func (ov *cmdOptionValue) Int() int64 {
	if ov.opt == nil {
		return 0
	}
	if i, ok := ov.opt.Value.(float64); ok {
		return int64(i)
	}
	return 0
}

// Bool returns the boolean value of the option.
func (ov *cmdOptionValue) Bool() bool {
	if ov.opt == nil {
		return false
	}
	if b, ok := ov.opt.Value.(bool); ok {
		return b
	}
	return false
}

// Float returns the float value of the option.
func (ov *cmdOptionValue) Float() float64 {
	if ov.opt == nil {
		return 0
	}
	if f, ok := ov.opt.Value.(float64); ok {
		return f
	}
	return 0
}

// IsEmpty returns true if the option was not provided.
func (ov *cmdOptionValue) IsEmpty() bool { return ov.opt == nil }

// MessageContext wraps Context with message-specific functionality.
type MessageContext struct {
	*Context
	message *discordgo.MessageCreate
}

// NewMessageContext creates a MessageContext from a Context and message data.
func NewMessageContext(c *Context, m *discordgo.MessageCreate) *MessageContext {
	return &MessageContext{
		Context: c,
		message: m,
	}
}

// Message returns the message that triggered this event.
func (mc *MessageContext) Message() *discordgo.MessageCreate { return mc.message }

// Content returns the message content.
func (mc *MessageContext) Content() string { return mc.message.Content }

// Reply sends a message to the channel.
func (mc *MessageContext) Reply(content string) (*discordgo.Message, error) {
	return mc.session.ChannelMessageSend(mc.message.ChannelID, content)
}

// ReplyEmbed sends an embed to the channel.
func (mc *MessageContext) ReplyEmbed(embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return mc.session.ChannelMessageSendEmbed(mc.message.ChannelID, embed)
}

// ReplyComplex sends a complex message with embeds, components, etc.
func (mc *MessageContext) ReplyComplex(data *discordgo.MessageSend) (*discordgo.Message, error) {
	return mc.session.ChannelMessageSendComplex(mc.message.ChannelID, data)
}

// Delete deletes the triggering message.
func (mc *MessageContext) Delete() error {
	return mc.session.ChannelMessageDelete(mc.message.ChannelID, mc.message.ID)
}

// ComponentContext wraps Context with component interaction functionality.
type ComponentContext struct {
	*Context
	interaction *discordgo.InteractionCreate
}

// NewComponentContext creates a ComponentContext from a Context.
func NewComponentContext(c *Context, i *discordgo.InteractionCreate) *ComponentContext {
	return &ComponentContext{
		Context:     c,
		interaction: i,
	}
}

// Interaction returns the raw interaction create event.
func (cc *ComponentContext) Interaction() *discordgo.InteractionCreate {
	return cc.interaction
}

// CustomID returns the custom ID of the component.
func (cc *ComponentContext) CustomID() string {
	return cc.interaction.MessageComponentData().CustomID
}

// Values returns the selected values (for select menus).
func (cc *ComponentContext) Values() []string {
	return cc.interaction.MessageComponentData().Values
}

// Acknowledge sends an acknowledgement response without content.
func (cc *ComponentContext) Acknowledge() error {
	if cc.responded {
		return nil
	}
	err := cc.session.InteractionRespond(cc.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		return fmt.Errorf("acknowledge component: %w", err)
	}
	cc.responded = true
	return nil
}

// Update updates the message that contains the component.
func (cc *ComponentContext) Update(content string) error {
	return cc.session.InteractionRespond(cc.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// Reply returns a new reply builder for this interaction.
func (cc *ComponentContext) Reply() *reply.Builder {
	return reply.NewBuilder(cc.session, cc.interaction.Interaction, cc.responded, cc.MarkResponded)
}

// ModalContext wraps Context with modal submission functionality.
type ModalContext struct {
	*Context
	interaction *discordgo.InteractionCreate
}

// NewModalContext creates a ModalContext from a Context.
func NewModalContext(c *Context, i *discordgo.InteractionCreate) *ModalContext {
	return &ModalContext{
		Context:     c,
		interaction: i,
	}
}

// Interaction returns the raw interaction create event.
func (mc *ModalContext) Interaction() *discordgo.InteractionCreate {
	return mc.interaction
}

// CustomID returns the custom ID of the modal.
func (mc *ModalContext) CustomID() string {
	return mc.interaction.ModalSubmitData().CustomID
}

// Value returns the value of a text input component by custom ID.
func (mc *ModalContext) Value(customID string) string {
	data := mc.interaction.ModalSubmitData()
	for _, row := range data.Components {
		if ar, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range ar.Components {
				if ti, ok := comp.(*discordgo.TextInput); ok {
					if ti.CustomID == customID {
						return ti.Value
					}
				}
			}
		}
	}
	return ""
}

// Reply returns a new reply builder for this interaction.
func (mc *ModalContext) Reply() *reply.Builder {
	return reply.NewBuilder(mc.session, mc.interaction.Interaction, mc.responded, mc.MarkResponded)
}

// VoiceContext wraps Context with voice state functionality.
type VoiceContext struct {
	*Context
	voiceState *discordgo.VoiceStateUpdate
}

// NewVoiceContext creates a VoiceContext from a Context.
func NewVoiceContext(c *Context, v *discordgo.VoiceStateUpdate) *VoiceContext {
	return &VoiceContext{
		Context:    c,
		voiceState: v,
	}
}

// VoiceState returns the voice state update data.
func (vc *VoiceContext) VoiceState() *discordgo.VoiceStateUpdate { return vc.voiceState }

// VoiceChannelID returns the voice channel ID (empty if user left voice).
func (vc *VoiceContext) VoiceChannelID() string { return vc.voiceState.ChannelID }

// IsJoin returns true if the user joined a voice channel.
func (vc *VoiceContext) IsJoin() bool {
	return vc.voiceState.ChannelID != "" && vc.voiceState.BeforeUpdate != nil && vc.voiceState.BeforeUpdate.ChannelID == ""
}

// IsLeave returns true if the user left a voice channel.
func (vc *VoiceContext) IsLeave() bool {
	return vc.voiceState.ChannelID == "" && vc.voiceState.BeforeUpdate != nil && vc.voiceState.BeforeUpdate.ChannelID != ""
}

// IsMove returns true if the user moved between voice channels.
func (vc *VoiceContext) IsMove() bool {
	if vc.voiceState.BeforeUpdate == nil {
		return false
	}
	return vc.voiceState.ChannelID != "" &&
		vc.voiceState.BeforeUpdate.ChannelID != "" &&
		vc.voiceState.ChannelID != vc.voiceState.BeforeUpdate.ChannelID
}

// ReactionContext wraps Context with reaction functionality.
type ReactionContext struct {
	*Context
	reaction *discordgo.MessageReaction
	member   *discordgo.Member
	removed  bool
}

// NewReactionAddContext creates a ReactionContext for a reaction add event.
func NewReactionAddContext(c *Context, r *discordgo.MessageReactionAdd) *ReactionContext {
	return &ReactionContext{
		Context:  c,
		reaction: r.MessageReaction,
		member:   r.Member,
		removed:  false,
	}
}

// NewReactionRemoveContext creates a ReactionContext for a reaction remove event.
func NewReactionRemoveContext(c *Context, r *discordgo.MessageReactionRemove) *ReactionContext {
	return &ReactionContext{
		Context:  c,
		reaction: r.MessageReaction,
		removed:  true,
	}
}

// MessageID returns the ID of the message that was reacted to.
func (rc *ReactionContext) MessageID() string { return rc.reaction.MessageID }

// Emoji returns the emoji that was added/removed.
func (rc *ReactionContext) Emoji() *discordgo.Emoji { return &rc.reaction.Emoji }

// EmojiName returns the emoji name (for custom emoji) or the unicode emoji.
func (rc *ReactionContext) EmojiName() string { return rc.reaction.Emoji.Name }

// IsRemoved returns true if this is a reaction removal event.
func (rc *ReactionContext) IsRemoved() bool { return rc.removed }

// ReactionMember returns the member who added the reaction (only for add events in guilds).
func (rc *ReactionContext) ReactionMember() *discordgo.Member { return rc.member }
