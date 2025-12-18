// Package domain provides Discord domain types that are library-agnostic.
// These types represent Discord concepts without depending on any specific Discord library.
package domain

// Snowflake ID type aliases for type safety.
// All Discord IDs are snowflakes represented as strings.
type (
	// GuildID represents a Discord guild (server) ID.
	GuildID string

	// ChannelID represents a Discord channel ID.
	ChannelID string

	// UserID represents a Discord user ID.
	UserID string

	// MessageID represents a Discord message ID.
	MessageID string

	// RoleID represents a Discord role ID.
	RoleID string

	// CommandID represents a Discord application command ID.
	CommandID string

	// InteractionID represents a Discord interaction ID.
	InteractionID string

	// EmojiID represents a Discord custom emoji ID.
	EmojiID string

	// AttachmentID represents a Discord attachment ID.
	AttachmentID string

	// WebhookID represents a Discord webhook ID.
	WebhookID string

	// ApplicationID represents a Discord application ID.
	ApplicationID string
)

// String returns the string representation of the ID.
func (id GuildID) String() string       { return string(id) }
func (id ChannelID) String() string     { return string(id) }
func (id UserID) String() string        { return string(id) }
func (id MessageID) String() string     { return string(id) }
func (id RoleID) String() string        { return string(id) }
func (id CommandID) String() string     { return string(id) }
func (id InteractionID) String() string { return string(id) }
func (id EmojiID) String() string       { return string(id) }
func (id AttachmentID) String() string  { return string(id) }
func (id WebhookID) String() string     { return string(id) }
func (id ApplicationID) String() string { return string(id) }

// IsEmpty returns true if the ID is empty.
func (id GuildID) IsEmpty() bool       { return id == "" }
func (id ChannelID) IsEmpty() bool     { return id == "" }
func (id UserID) IsEmpty() bool        { return id == "" }
func (id MessageID) IsEmpty() bool     { return id == "" }
func (id RoleID) IsEmpty() bool        { return id == "" }
func (id CommandID) IsEmpty() bool     { return id == "" }
func (id InteractionID) IsEmpty() bool { return id == "" }
func (id EmojiID) IsEmpty() bool       { return id == "" }
func (id AttachmentID) IsEmpty() bool  { return id == "" }
func (id WebhookID) IsEmpty() bool     { return id == "" }
func (id ApplicationID) IsEmpty() bool { return id == "" }
