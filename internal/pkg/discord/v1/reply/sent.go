package reply

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// SentMessage represents a sent Discord message with methods for editing, deleting, and reacting.
type SentMessage struct {
	session     *discordgo.Session
	interaction *discordgo.Interaction
	message     *discordgo.Message
}

// NewSentMessage creates a new SentMessage wrapper.
func NewSentMessage(s *discordgo.Session, i *discordgo.Interaction, m *discordgo.Message) *SentMessage {
	return &SentMessage{
		session:     s,
		interaction: i,
		message:     m,
	}
}

// ID returns the message ID.
func (sm *SentMessage) ID() string {
	if sm.message != nil {
		return sm.message.ID
	}
	return ""
}

// ChannelID returns the channel ID.
func (sm *SentMessage) ChannelID() string {
	if sm.message != nil {
		return sm.message.ChannelID
	}
	return ""
}

// Message returns the underlying discordgo.Message.
func (sm *SentMessage) Message() *discordgo.Message {
	return sm.message
}

// Edit modifies the sent message content.
func (sm *SentMessage) Edit(content string) error {
	_, err := sm.session.InteractionResponseEdit(sm.interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		return fmt.Errorf("edit message: %w", err)
	}
	return nil
}

// EditEmbed modifies the sent message with new embeds.
func (sm *SentMessage) EditEmbed(embeds ...*discordgo.MessageEmbed) error {
	_, err := sm.session.InteractionResponseEdit(sm.interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
	})
	if err != nil {
		return fmt.Errorf("edit embeds: %w", err)
	}
	return nil
}

// Delete removes the sent message.
func (sm *SentMessage) Delete() error {
	err := sm.session.InteractionResponseDelete(sm.interaction)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	return nil
}

// React adds an emoji reaction to the message.
func (sm *SentMessage) React(emoji string) error {
	if sm.message == nil {
		return fmt.Errorf("no message to react to")
	}
	err := sm.session.MessageReactionAdd(sm.message.ChannelID, sm.message.ID, emoji)
	if err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}
	return nil
}

// RemoveReaction removes an emoji reaction from the message.
func (sm *SentMessage) RemoveReaction(emoji string) error {
	if sm.message == nil {
		return fmt.Errorf("no message reference")
	}
	err := sm.session.MessageReactionRemove(sm.message.ChannelID, sm.message.ID, emoji, "@me")
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	return nil
}

