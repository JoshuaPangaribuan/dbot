package domain

import "time"

// User represents a Discord user.
type User struct {
	ID            UserID
	Username      string
	Discriminator string
	GlobalName    string
	Avatar        string
	Bot           bool
	System        bool
	Banner        string
	AccentColor   int
}

// Member represents a guild member (user + guild-specific data).
type Member struct {
	User         *User
	GuildID      GuildID
	Nick         string
	Avatar       string
	Roles        []RoleID
	JoinedAt     time.Time
	PremiumSince *time.Time
	Deaf         bool
	Mute         bool
	Pending      bool
	Permissions  int64
}

// Role represents a Discord role.
type Role struct {
	ID          RoleID
	Name        string
	Color       int
	Hoist       bool
	Position    int
	Permissions int64
	Managed     bool
	Mentionable bool
}

// DisplayName returns the user's display name (global name or username).
func (u *User) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

// Mention returns the mention string for the user.
func (u *User) Mention() string {
	return "<@" + u.ID.String() + ">"
}

// DisplayName returns the member's display name (nick or user display name).
func (m *Member) DisplayName() string {
	if m.Nick != "" {
		return m.Nick
	}
	if m.User != nil {
		return m.User.DisplayName()
	}
	return ""
}

// HasRole returns true if the member has the specified role.
func (m *Member) HasRole(roleID RoleID) bool {
	for _, r := range m.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}
