package domain

// Guild represents a Discord guild (server).
type Guild struct {
	ID          GuildID
	Name        string
	Icon        string
	OwnerID     UserID
	Permissions int64
	Features    []string
}

// Channel represents a Discord channel.
type Channel struct {
	ID                   ChannelID
	Type                 ChannelType
	GuildID              GuildID
	Name                 string
	Topic                string
	NSFW                 bool
	Position             int
	ParentID             ChannelID
	PermissionOverwrites []PermissionOverwrite
}

// ChannelType represents the type of Discord channel.
type ChannelType int

const (
	ChannelTypeGuildText ChannelType = iota
	ChannelTypeDM
	ChannelTypeGuildVoice
	ChannelTypeGroupDM
	ChannelTypeGuildCategory
	ChannelTypeGuildAnnouncement
	ChannelTypeAnnouncementThread ChannelType = 10
	ChannelTypePublicThread       ChannelType = 11
	ChannelTypePrivateThread      ChannelType = 12
	ChannelTypeGuildStageVoice    ChannelType = 13
	ChannelTypeGuildDirectory     ChannelType = 14
	ChannelTypeGuildForum         ChannelType = 15
	ChannelTypeGuildMedia         ChannelType = 16
)

// PermissionOverwrite represents a channel permission overwrite.
type PermissionOverwrite struct {
	ID    string // Role or User ID
	Type  PermissionOverwriteType
	Allow int64
	Deny  int64
}

// PermissionOverwriteType represents the type of permission overwrite.
type PermissionOverwriteType int

const (
	PermissionOverwriteTypeRole PermissionOverwriteType = iota
	PermissionOverwriteTypeMember
)

// IsText returns true if this is a text-based channel.
func (c *Channel) IsText() bool {
	switch c.Type {
	case ChannelTypeGuildText, ChannelTypeDM, ChannelTypeGroupDM,
		ChannelTypeGuildAnnouncement, ChannelTypeAnnouncementThread,
		ChannelTypePublicThread, ChannelTypePrivateThread, ChannelTypeGuildForum:
		return true
	}
	return false
}

// IsVoice returns true if this is a voice-based channel.
func (c *Channel) IsVoice() bool {
	switch c.Type {
	case ChannelTypeGuildVoice, ChannelTypeGuildStageVoice:
		return true
	}
	return false
}

// IsThread returns true if this is a thread channel.
func (c *Channel) IsThread() bool {
	switch c.Type {
	case ChannelTypeAnnouncementThread, ChannelTypePublicThread, ChannelTypePrivateThread:
		return true
	}
	return false
}
