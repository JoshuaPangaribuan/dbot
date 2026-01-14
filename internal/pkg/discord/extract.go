package discord

import "github.com/bwmarrin/discordgo"

// EventData holds common data extracted from Discord events.
type EventData struct {
	UserID    string
	GuildID   string
	ChannelID string
}

// ExtractEventData extracts common IDs from a Discord event.
// Returns empty strings for fields that cannot be determined from the event type.
func ExtractEventData(e *Event) EventData {
	switch data := e.Data().(type) {
	case *discordgo.MessageCreate:
		return EventData{
			UserID:    extractUserIDFromMessage(data),
			GuildID:   data.GuildID,
			ChannelID: data.ChannelID,
		}
	case *discordgo.InteractionCreate:
		return EventData{
			UserID:    extractUserIDFromInteraction(data),
			GuildID:   data.GuildID,
			ChannelID: data.ChannelID,
		}
	case *discordgo.VoiceStateUpdate:
		return EventData{
			UserID:    data.UserID,
			GuildID:   data.GuildID,
			ChannelID: data.ChannelID,
		}
	case *discordgo.MessageReactionAdd:
		return EventData{
			UserID:    data.UserID,
			GuildID:   data.GuildID,
			ChannelID: data.ChannelID,
		}
	case *discordgo.MessageReactionRemove:
		return EventData{
			UserID:    data.UserID,
			GuildID:   data.GuildID,
			ChannelID: data.ChannelID,
		}
	}
	return EventData{}
}

// ExtractUserID extracts the user ID from a Discord event.
// Returns empty string if the user ID cannot be determined.
func ExtractUserID(e *Event) string {
	switch data := e.Data().(type) {
	case *discordgo.MessageCreate:
		return extractUserIDFromMessage(data)
	case *discordgo.InteractionCreate:
		return extractUserIDFromInteraction(data)
	case *discordgo.VoiceStateUpdate:
		return data.UserID
	case *discordgo.MessageReactionAdd:
		return data.UserID
	case *discordgo.MessageReactionRemove:
		return data.UserID
	}
	return ""
}

// ExtractGuildID extracts the guild ID from a Discord event.
// Returns empty string for DMs or if the guild ID cannot be determined.
func ExtractGuildID(e *Event) string {
	switch data := e.Data().(type) {
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

// ExtractChannelID extracts the channel ID from a Discord event.
// Returns empty string if the channel ID cannot be determined.
func ExtractChannelID(e *Event) string {
	switch data := e.Data().(type) {
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

// extractUserIDFromMessage extracts user ID from MessageCreate event.
func extractUserIDFromMessage(m *discordgo.MessageCreate) string {
	if m.Author != nil {
		return m.Author.ID
	}
	return ""
}

// extractUserIDFromInteraction extracts user ID from InteractionCreate event.
func extractUserIDFromInteraction(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}
