package domain

import (
	"io"
	"time"
)

// Message represents a Discord message.
type Message struct {
	ID              MessageID
	ChannelID       ChannelID
	GuildID         GuildID
	Author          *User
	Member          *Member
	Content         string
	Timestamp       time.Time
	EditedTimestamp *time.Time
	TTS             bool
	MentionEveryone bool
	Mentions        []*User
	MentionRoles    []RoleID
	Attachments     []*Attachment
	Embeds          []*Embed
	Reactions       []*Reaction
	Pinned          bool
	Type            MessageType
	ReferencedMsg   *Message
}

// MessageType represents the type of message.
type MessageType int

const (
	MessageTypeDefault MessageType = iota
	MessageTypeRecipientAdd
	MessageTypeRecipientRemove
	MessageTypeCall
	MessageTypeChannelNameChange
	MessageTypeChannelIconChange
	MessageTypePinnedMessage
	MessageTypeGuildMemberJoin
	MessageTypeUserPremiumSub
	MessageTypeUserPremiumSubTier1
	MessageTypeUserPremiumSubTier2
	MessageTypeUserPremiumSubTier3
	MessageTypeChannelFollowAdd
	MessageTypeGuildDiscoveryDisqualified MessageType = 14
	MessageTypeGuildDiscoveryRequalified  MessageType = 15
	MessageTypeGuildDiscoveryGraceInitial MessageType = 16
	MessageTypeGuildDiscoveryGraceFinal   MessageType = 17
	MessageTypeThreadCreated              MessageType = 18
	MessageTypeReply                      MessageType = 19
	MessageTypeChatInputCommand           MessageType = 20
	MessageTypeThreadStarterMessage       MessageType = 21
	MessageTypeGuildInviteReminder        MessageType = 22
	MessageTypeContextMenuCommand         MessageType = 23
	MessageTypeAutoModerationAction       MessageType = 24
	MessageTypeRoleSubscriptionPurchase   MessageType = 25
	MessageTypeInteractionPremiumUpsell   MessageType = 26
	MessageTypeStageStart                 MessageType = 27
	MessageTypeStageEnd                   MessageType = 28
	MessageTypeStageSpeaker               MessageType = 29
	MessageTypeStageTopic                 MessageType = 31
	MessageTypeGuildApplicationPremiumSub MessageType = 32
)

// Attachment represents a message attachment.
type Attachment struct {
	ID          AttachmentID
	URL         string
	ProxyURL    string
	Filename    string
	ContentType string
	Size        int
	Height      int
	Width       int
	Ephemeral   bool
}

// Reaction represents a message reaction.
type Reaction struct {
	Count int
	Me    bool
	Emoji *Emoji
}

// Emoji represents a Discord emoji.
type Emoji struct {
	ID       EmojiID
	Name     string
	Animated bool
}

// Embed represents a Discord message embed.
type Embed struct {
	Title       string
	Type        string
	Description string
	URL         string
	Timestamp   string
	Color       int
	Footer      *EmbedFooter
	Image       *EmbedImage
	Thumbnail   *EmbedThumbnail
	Video       *EmbedVideo
	Provider    *EmbedProvider
	Author      *EmbedAuthor
	Fields      []*EmbedField
}

// EmbedFooter represents an embed footer.
type EmbedFooter struct {
	Text         string
	IconURL      string
	ProxyIconURL string
}

// EmbedImage represents an embed image.
type EmbedImage struct {
	URL      string
	ProxyURL string
	Height   int
	Width    int
}

// EmbedThumbnail represents an embed thumbnail.
type EmbedThumbnail struct {
	URL      string
	ProxyURL string
	Height   int
	Width    int
}

// EmbedVideo represents an embed video.
type EmbedVideo struct {
	URL    string
	Height int
	Width  int
}

// EmbedProvider represents an embed provider.
type EmbedProvider struct {
	Name string
	URL  string
}

// EmbedAuthor represents an embed author.
type EmbedAuthor struct {
	Name         string
	URL          string
	IconURL      string
	ProxyIconURL string
}

// EmbedField represents an embed field.
type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

// MessageSend represents a message to be sent.
type MessageSend struct {
	Content         string
	TTS             bool
	Embeds          []*Embed
	Components      []Component
	Files           []*File
	AllowedMentions *AllowedMentions
	Reference       *MessageReference
}

// MessageEdit represents a message edit.
type MessageEdit struct {
	Content         *string
	Embeds          *[]*Embed
	Components      *[]Component
	Files           []*File
	AllowedMentions *AllowedMentions
}

// MessageReference represents a message reference (for replies).
type MessageReference struct {
	MessageID MessageID
	ChannelID ChannelID
	GuildID   GuildID
}

// File represents a file attachment.
type File struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

// AllowedMentions specifies which mentions are allowed.
type AllowedMentions struct {
	Parse       []AllowedMentionType
	Roles       []RoleID
	Users       []UserID
	RepliedUser bool
}

// AllowedMentionType represents a type of mention that's allowed.
type AllowedMentionType string

const (
	AllowedMentionTypeRoles    AllowedMentionType = "roles"
	AllowedMentionTypeUsers    AllowedMentionType = "users"
	AllowedMentionTypeEveryone AllowedMentionType = "everyone"
)
