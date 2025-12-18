package domain

import "context"

// EventType identifies the kind of Discord event.
type EventType int

const (
	// EventTypeMessage represents a message create event.
	EventTypeMessage EventType = iota
	// EventTypeCommand represents a slash command interaction.
	EventTypeCommand
	// EventTypeComponent represents a button/select menu interaction.
	EventTypeComponent
	// EventTypeModal represents a modal submission.
	EventTypeModal
	// EventTypeAutocomplete represents an autocomplete interaction.
	EventTypeAutocomplete
	// EventTypeVoiceStateUpdate represents a voice state change.
	EventTypeVoiceStateUpdate
	// EventTypeReactionAdd represents a reaction being added.
	EventTypeReactionAdd
	// EventTypeReactionRemove represents a reaction being removed.
	EventTypeReactionRemove
)

// String returns the string representation of the event type.
func (e EventType) String() string {
	switch e {
	case EventTypeMessage:
		return "message"
	case EventTypeCommand:
		return "command"
	case EventTypeComponent:
		return "component"
	case EventTypeModal:
		return "modal"
	case EventTypeAutocomplete:
		return "autocomplete"
	case EventTypeVoiceStateUpdate:
		return "voice_state_update"
	case EventTypeReactionAdd:
		return "reaction_add"
	case EventTypeReactionRemove:
		return "reaction_remove"
	default:
		return "unknown"
	}
}

// Event represents a Discord event with its associated data.
// This is the v2 library-agnostic event type.
type Event struct {
	typ EventType
	ctx context.Context
	raw any // Raw adapter-specific data for escape hatch

	// Event-specific data (only one populated based on type)
	Message     *Message
	Interaction *Interaction
	VoiceState  *VoiceState
	Reaction    *MessageReaction
}

// Type returns the event type.
func (e *Event) Type() EventType { return e.typ }

// Context returns the context associated with this event.
func (e *Event) Context() context.Context { return e.ctx }

// Raw returns the underlying adapter-specific event data.
// Use this as an escape hatch when v2 types don't expose needed data.
func (e *Event) Raw() any { return e.raw }

// GuildID returns the guild ID where the event occurred, or empty for DMs.
func (e *Event) GuildID() GuildID {
	switch e.typ {
	case EventTypeMessage:
		if e.Message != nil {
			return e.Message.GuildID
		}
	case EventTypeCommand, EventTypeComponent, EventTypeModal, EventTypeAutocomplete:
		if e.Interaction != nil {
			return e.Interaction.GuildID
		}
	case EventTypeVoiceStateUpdate:
		if e.VoiceState != nil {
			return e.VoiceState.GuildID
		}
	case EventTypeReactionAdd, EventTypeReactionRemove:
		if e.Reaction != nil {
			return e.Reaction.GuildID
		}
	}
	return ""
}

// ChannelID returns the channel ID where the event occurred.
func (e *Event) ChannelID() ChannelID {
	switch e.typ {
	case EventTypeMessage:
		if e.Message != nil {
			return e.Message.ChannelID
		}
	case EventTypeCommand, EventTypeComponent, EventTypeModal, EventTypeAutocomplete:
		if e.Interaction != nil {
			return e.Interaction.ChannelID
		}
	case EventTypeVoiceStateUpdate:
		if e.VoiceState != nil {
			return e.VoiceState.ChannelID
		}
	case EventTypeReactionAdd, EventTypeReactionRemove:
		if e.Reaction != nil {
			return e.Reaction.ChannelID
		}
	}
	return ""
}

// UserID returns the user ID who triggered the event.
func (e *Event) UserID() UserID {
	switch e.typ {
	case EventTypeMessage:
		if e.Message != nil && e.Message.Author != nil {
			return e.Message.Author.ID
		}
	case EventTypeCommand, EventTypeComponent, EventTypeModal, EventTypeAutocomplete:
		if e.Interaction != nil {
			return e.Interaction.GetUserID()
		}
	case EventTypeVoiceStateUpdate:
		if e.VoiceState != nil {
			return e.VoiceState.UserID
		}
	case EventTypeReactionAdd, EventTypeReactionRemove:
		if e.Reaction != nil {
			return e.Reaction.UserID
		}
	}
	return ""
}

// User returns the user who triggered the event.
func (e *Event) User() *User {
	switch e.typ {
	case EventTypeMessage:
		if e.Message != nil {
			return e.Message.Author
		}
	case EventTypeCommand, EventTypeComponent, EventTypeModal, EventTypeAutocomplete:
		if e.Interaction != nil {
			return e.Interaction.GetUser()
		}
	case EventTypeReactionAdd:
		if e.Reaction != nil && e.Reaction.Member != nil {
			return e.Reaction.Member.User
		}
	}
	return nil
}

// Member returns the guild member who triggered the event (nil for DMs).
func (e *Event) Member() *Member {
	switch e.typ {
	case EventTypeMessage:
		if e.Message != nil {
			return e.Message.Member
		}
	case EventTypeCommand, EventTypeComponent, EventTypeModal, EventTypeAutocomplete:
		if e.Interaction != nil {
			return e.Interaction.Member
		}
	case EventTypeReactionAdd:
		if e.Reaction != nil {
			return e.Reaction.Member
		}
	}
	return nil
}

// NewEvent creates a new Event.
func NewEvent(typ EventType, ctx context.Context) *Event {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Event{
		typ: typ,
		ctx: ctx,
	}
}

// WithMessage sets the message data.
func (e *Event) WithMessage(m *Message) *Event {
	e.Message = m
	return e
}

// WithInteraction sets the interaction data.
func (e *Event) WithInteraction(i *Interaction) *Event {
	e.Interaction = i
	return e
}

// WithVoiceState sets the voice state data.
func (e *Event) WithVoiceState(v *VoiceState) *Event {
	e.VoiceState = v
	return e
}

// WithReaction sets the reaction data.
func (e *Event) WithReaction(r *MessageReaction) *Event {
	e.Reaction = r
	return e
}

// WithRaw sets the raw adapter-specific data.
func (e *Event) WithRaw(raw any) *Event {
	e.raw = raw
	return e
}

// WithContext sets the context for this event.
func (e *Event) WithContext(ctx context.Context) *Event {
	if ctx != nil {
		e.ctx = ctx
	}
	return e
}

// VoiceState represents a voice state update.
type VoiceState struct {
	GuildID   GuildID
	ChannelID ChannelID
	UserID    UserID
	Member    *Member
	SessionID string
	Deaf      bool
	Mute      bool
	SelfDeaf  bool
	SelfMute  bool
	SelfVideo bool
	Suppress  bool

	// BeforeUpdate contains the previous voice state (if available).
	BeforeUpdate *VoiceState
}

// IsJoin returns true if the user joined a voice channel.
func (v *VoiceState) IsJoin() bool {
	return !v.ChannelID.IsEmpty() && v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID.IsEmpty()
}

// IsLeave returns true if the user left a voice channel.
func (v *VoiceState) IsLeave() bool {
	return v.ChannelID.IsEmpty() && v.BeforeUpdate != nil && !v.BeforeUpdate.ChannelID.IsEmpty()
}

// IsMove returns true if the user moved between voice channels.
func (v *VoiceState) IsMove() bool {
	if v.BeforeUpdate == nil {
		return false
	}
	return !v.ChannelID.IsEmpty() &&
		!v.BeforeUpdate.ChannelID.IsEmpty() &&
		v.ChannelID != v.BeforeUpdate.ChannelID
}

// MessageReaction represents a message reaction event.
type MessageReaction struct {
	UserID    UserID
	ChannelID ChannelID
	MessageID MessageID
	GuildID   GuildID
	Member    *Member
	Emoji     *Emoji
	Removed   bool // True for remove events
}
