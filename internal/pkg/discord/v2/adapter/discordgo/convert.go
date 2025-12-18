// Package discordgo provides an adapter for the bwmarrin/discordgo library.
package discordgo

import (
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/bwmarrin/discordgo"
)

// --- Localization Helpers ---

// localizationsToDiscordgo converts string-keyed localizations to discordgo.Locale-keyed.
func localizationsToDiscordgo(locs map[string]string) map[discordgo.Locale]string {
	if len(locs) == 0 {
		return nil
	}
	result := make(map[discordgo.Locale]string, len(locs))
	for k, v := range locs {
		result[discordgo.Locale(k)] = v
	}
	return result
}

// --- User/Member Conversions ---

func userFromDiscordgo(u *discordgo.User) *domain.User {
	if u == nil {
		return nil
	}
	return &domain.User{
		ID:            domain.UserID(u.ID),
		Username:      u.Username,
		Discriminator: u.Discriminator,
		GlobalName:    u.GlobalName,
		Avatar:        u.Avatar,
		Bot:           u.Bot,
		System:        u.System,
		Banner:        u.Banner,
		AccentColor:   u.AccentColor,
	}
}

func memberFromDiscordgo(m *discordgo.Member, guildID string) *domain.Member {
	if m == nil {
		return nil
	}
	roles := make([]domain.RoleID, len(m.Roles))
	for i, r := range m.Roles {
		roles[i] = domain.RoleID(r)
	}
	return &domain.Member{
		User:        userFromDiscordgo(m.User),
		GuildID:     domain.GuildID(guildID),
		Nick:        m.Nick,
		Avatar:      m.Avatar,
		Roles:       roles,
		JoinedAt:    m.JoinedAt,
		Deaf:        m.Deaf,
		Mute:        m.Mute,
		Pending:     m.Pending,
		Permissions: m.Permissions,
	}
}

func roleFromDiscordgo(r *discordgo.Role) *domain.Role {
	if r == nil {
		return nil
	}
	return &domain.Role{
		ID:          domain.RoleID(r.ID),
		Name:        r.Name,
		Color:       r.Color,
		Hoist:       r.Hoist,
		Position:    r.Position,
		Permissions: r.Permissions,
		Managed:     r.Managed,
		Mentionable: r.Mentionable,
	}
}

// --- Channel Conversions ---

func channelFromDiscordgo(c *discordgo.Channel) *domain.Channel {
	if c == nil {
		return nil
	}
	overwrites := make([]domain.PermissionOverwrite, len(c.PermissionOverwrites))
	for i, o := range c.PermissionOverwrites {
		overwrites[i] = domain.PermissionOverwrite{
			ID:    o.ID,
			Type:  domain.PermissionOverwriteType(o.Type),
			Allow: o.Allow,
			Deny:  o.Deny,
		}
	}
	return &domain.Channel{
		ID:                   domain.ChannelID(c.ID),
		Type:                 domain.ChannelType(c.Type),
		GuildID:              domain.GuildID(c.GuildID),
		Name:                 c.Name,
		Topic:                c.Topic,
		NSFW:                 c.NSFW,
		Position:             c.Position,
		ParentID:             domain.ChannelID(c.ParentID),
		PermissionOverwrites: overwrites,
	}
}

// --- Message Conversions ---

func messageFromDiscordgo(m *discordgo.Message) *domain.Message {
	if m == nil {
		return nil
	}
	mentions := make([]*domain.User, len(m.Mentions))
	for i, u := range m.Mentions {
		mentions[i] = userFromDiscordgo(u)
	}
	mentionRoles := make([]domain.RoleID, len(m.MentionRoles))
	for i, r := range m.MentionRoles {
		mentionRoles[i] = domain.RoleID(r)
	}
	attachments := make([]*domain.Attachment, len(m.Attachments))
	for i, a := range m.Attachments {
		attachments[i] = attachmentFromDiscordgo(a)
	}
	embeds := make([]*domain.Embed, len(m.Embeds))
	for i, e := range m.Embeds {
		embeds[i] = embedFromDiscordgo(e)
	}
	reactions := make([]*domain.Reaction, len(m.Reactions))
	for i, r := range m.Reactions {
		reactions[i] = &domain.Reaction{
			Count: r.Count,
			Me:    r.Me,
			Emoji: emojiPtrFromDiscordgo(r.Emoji),
		}
	}
	msg := &domain.Message{
		ID:              domain.MessageID(m.ID),
		ChannelID:       domain.ChannelID(m.ChannelID),
		GuildID:         domain.GuildID(m.GuildID),
		Author:          userFromDiscordgo(m.Author),
		Member:          memberFromDiscordgo(m.Member, m.GuildID),
		Content:         m.Content,
		Timestamp:       m.Timestamp,
		TTS:             m.TTS,
		MentionEveryone: m.MentionEveryone,
		Mentions:        mentions,
		MentionRoles:    mentionRoles,
		Attachments:     attachments,
		Embeds:          embeds,
		Reactions:       reactions,
		Pinned:          m.Pinned,
		Type:            domain.MessageType(m.Type),
	}
	if m.EditedTimestamp != nil {
		msg.EditedTimestamp = m.EditedTimestamp
	}
	if m.ReferencedMessage != nil {
		msg.ReferencedMsg = messageFromDiscordgo(m.ReferencedMessage)
	}
	return msg
}

func attachmentFromDiscordgo(a *discordgo.MessageAttachment) *domain.Attachment {
	if a == nil {
		return nil
	}
	return &domain.Attachment{
		ID:          domain.AttachmentID(a.ID),
		URL:         a.URL,
		ProxyURL:    a.ProxyURL,
		Filename:    a.Filename,
		ContentType: a.ContentType,
		Size:        a.Size,
		Height:      a.Height,
		Width:       a.Width,
		Ephemeral:   a.Ephemeral,
	}
}

func emojiFromDiscordgo(e discordgo.Emoji) *domain.Emoji {
	return &domain.Emoji{
		ID:       domain.EmojiID(e.ID),
		Name:     e.Name,
		Animated: e.Animated,
	}
}

func emojiPtrFromDiscordgo(e *discordgo.Emoji) *domain.Emoji {
	if e == nil {
		return nil
	}
	return &domain.Emoji{
		ID:       domain.EmojiID(e.ID),
		Name:     e.Name,
		Animated: e.Animated,
	}
}

// --- Embed Conversions ---

func embedFromDiscordgo(e *discordgo.MessageEmbed) *domain.Embed {
	if e == nil {
		return nil
	}
	embed := &domain.Embed{
		Title:       e.Title,
		Type:        string(e.Type),
		Description: e.Description,
		URL:         e.URL,
		Timestamp:   e.Timestamp,
		Color:       e.Color,
	}
	if e.Footer != nil {
		embed.Footer = &domain.EmbedFooter{
			Text:         e.Footer.Text,
			IconURL:      e.Footer.IconURL,
			ProxyIconURL: e.Footer.ProxyIconURL,
		}
	}
	if e.Image != nil {
		embed.Image = &domain.EmbedImage{
			URL:      e.Image.URL,
			ProxyURL: e.Image.ProxyURL,
			Height:   e.Image.Height,
			Width:    e.Image.Width,
		}
	}
	if e.Thumbnail != nil {
		embed.Thumbnail = &domain.EmbedThumbnail{
			URL:      e.Thumbnail.URL,
			ProxyURL: e.Thumbnail.ProxyURL,
			Height:   e.Thumbnail.Height,
			Width:    e.Thumbnail.Width,
		}
	}
	if e.Video != nil {
		embed.Video = &domain.EmbedVideo{
			URL:    e.Video.URL,
			Height: e.Video.Height,
			Width:  e.Video.Width,
		}
	}
	if e.Provider != nil {
		embed.Provider = &domain.EmbedProvider{
			Name: e.Provider.Name,
			URL:  e.Provider.URL,
		}
	}
	if e.Author != nil {
		embed.Author = &domain.EmbedAuthor{
			Name:         e.Author.Name,
			URL:          e.Author.URL,
			IconURL:      e.Author.IconURL,
			ProxyIconURL: e.Author.ProxyIconURL,
		}
	}
	if len(e.Fields) > 0 {
		embed.Fields = make([]*domain.EmbedField, len(e.Fields))
		for i, f := range e.Fields {
			embed.Fields[i] = &domain.EmbedField{
				Name:   f.Name,
				Value:  f.Value,
				Inline: f.Inline,
			}
		}
	}
	return embed
}

func embedToDiscordgo(e *domain.Embed) *discordgo.MessageEmbed {
	if e == nil {
		return nil
	}
	embed := &discordgo.MessageEmbed{
		Title:       e.Title,
		Type:        discordgo.EmbedType(e.Type),
		Description: e.Description,
		URL:         e.URL,
		Timestamp:   e.Timestamp,
		Color:       e.Color,
	}
	if e.Footer != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text:    e.Footer.Text,
			IconURL: e.Footer.IconURL,
		}
	}
	if e.Image != nil {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: e.Image.URL,
		}
	}
	if e.Thumbnail != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: e.Thumbnail.URL,
		}
	}
	if e.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    e.Author.Name,
			URL:     e.Author.URL,
			IconURL: e.Author.IconURL,
		}
	}
	if len(e.Fields) > 0 {
		embed.Fields = make([]*discordgo.MessageEmbedField, len(e.Fields))
		for i, f := range e.Fields {
			embed.Fields[i] = &discordgo.MessageEmbedField{
				Name:   f.Name,
				Value:  f.Value,
				Inline: f.Inline,
			}
		}
	}
	return embed
}

// --- Interaction Conversions ---

func interactionFromDiscordgo(i *discordgo.InteractionCreate) *domain.Interaction {
	if i == nil {
		return nil
	}
	interaction := &domain.Interaction{
		ID:        domain.InteractionID(i.ID),
		Type:      domain.InteractionType(i.Type),
		Token:     i.Token,
		GuildID:   domain.GuildID(i.GuildID),
		ChannelID: domain.ChannelID(i.ChannelID),
		User:      userFromDiscordgo(i.User),
		Member:    memberFromDiscordgo(i.Member, i.GuildID),
		AppID:     domain.ApplicationID(i.AppID),
		Locale:    string(i.Locale),
		Version:   i.Version,
	}
	if i.Message != nil {
		interaction.Message = messageFromDiscordgo(i.Message)
	}
	if i.GuildLocale != nil {
		interaction.GuildLocale = string(*i.GuildLocale)
	}

	// Parse interaction data based on type
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		interaction.CommandData = &domain.CommandInteractionData{
			ID:       domain.CommandID(data.ID),
			Name:     data.Name,
			Type:     domain.CommandType(data.CommandType),
			Options:  commandOptionsFromDiscordgo(data.Options),
			Resolved: resolvedDataFromDiscordgo(data.Resolved, i.GuildID),
			TargetID: data.TargetID,
		}
	case discordgo.InteractionMessageComponent:
		data := i.MessageComponentData()
		interaction.ComponentData = &domain.ComponentInteractionData{
			CustomID:      data.CustomID,
			ComponentType: domain.ComponentType(data.ComponentType),
			Values:        data.Values,
			// Component interactions don't have the same resolved structure
		}
	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		interaction.ModalData = &domain.ModalInteractionData{
			CustomID:   data.CustomID,
			Components: componentsFromDiscordgo(data.Components),
		}
	}

	return interaction
}

func commandOptionsFromDiscordgo(opts []*discordgo.ApplicationCommandInteractionDataOption) []*domain.CommandOption {
	if len(opts) == 0 {
		return nil
	}
	result := make([]*domain.CommandOption, len(opts))
	for i, opt := range opts {
		result[i] = &domain.CommandOption{
			Name:    opt.Name,
			Type:    domain.CommandOptionType(opt.Type),
			Value:   opt.Value,
			Options: commandOptionsFromDiscordgo(opt.Options),
			Focused: opt.Focused,
		}
	}
	return result
}

func resolvedDataFromDiscordgo(r *discordgo.ApplicationCommandInteractionDataResolved, guildID string) *domain.ResolvedData {
	if r == nil {
		return nil
	}
	resolved := &domain.ResolvedData{}
	if len(r.Users) > 0 {
		resolved.Users = make(map[domain.UserID]*domain.User)
		for id, u := range r.Users {
			resolved.Users[domain.UserID(id)] = userFromDiscordgo(u)
		}
	}
	if len(r.Members) > 0 {
		resolved.Members = make(map[domain.UserID]*domain.Member)
		for id, m := range r.Members {
			resolved.Members[domain.UserID(id)] = memberFromDiscordgo(m, guildID)
		}
	}
	if len(r.Roles) > 0 {
		resolved.Roles = make(map[domain.RoleID]*domain.Role)
		for id, role := range r.Roles {
			resolved.Roles[domain.RoleID(id)] = roleFromDiscordgo(role)
		}
	}
	if len(r.Channels) > 0 {
		resolved.Channels = make(map[domain.ChannelID]*domain.Channel)
		for id, ch := range r.Channels {
			resolved.Channels[domain.ChannelID(id)] = channelFromDiscordgo(ch)
		}
	}
	if len(r.Messages) > 0 {
		resolved.Messages = make(map[domain.MessageID]*domain.Message)
		for id, msg := range r.Messages {
			resolved.Messages[domain.MessageID(id)] = messageFromDiscordgo(msg)
		}
	}
	return resolved
}

func componentsFromDiscordgo(comps []discordgo.MessageComponent) []domain.Component {
	if len(comps) == 0 {
		return nil
	}
	result := make([]domain.Component, len(comps))
	for i, comp := range comps {
		switch c := comp.(type) {
		case *discordgo.ActionsRow:
			result[i] = &domain.ActionRow{
				Components: componentsFromDiscordgo(c.Components),
			}
		case *discordgo.TextInput:
			result[i] = &domain.TextInput{
				CustomID:    c.CustomID,
				Style:       domain.TextInputStyle(c.Style),
				Label:       c.Label,
				MinLength:   c.MinLength,
				MaxLength:   c.MaxLength,
				Required:    c.Required,
				Value:       c.Value,
				Placeholder: c.Placeholder,
			}
		}
	}
	return result
}

// --- Voice State Conversions ---

func voiceStateFromDiscordgo(v *discordgo.VoiceStateUpdate) *domain.VoiceState {
	if v == nil {
		return nil
	}
	vs := &domain.VoiceState{
		GuildID:   domain.GuildID(v.GuildID),
		ChannelID: domain.ChannelID(v.ChannelID),
		UserID:    domain.UserID(v.UserID),
		Member:    memberFromDiscordgo(v.Member, v.GuildID),
		SessionID: v.SessionID,
		Deaf:      v.Deaf,
		Mute:      v.Mute,
		SelfDeaf:  v.SelfDeaf,
		SelfMute:  v.SelfMute,
		SelfVideo: v.SelfVideo,
		Suppress:  v.Suppress,
	}
	if v.BeforeUpdate != nil {
		vs.BeforeUpdate = &domain.VoiceState{
			GuildID:   domain.GuildID(v.BeforeUpdate.GuildID),
			ChannelID: domain.ChannelID(v.BeforeUpdate.ChannelID),
			UserID:    domain.UserID(v.BeforeUpdate.UserID),
			SessionID: v.BeforeUpdate.SessionID,
			Deaf:      v.BeforeUpdate.Deaf,
			Mute:      v.BeforeUpdate.Mute,
			SelfDeaf:  v.BeforeUpdate.SelfDeaf,
			SelfMute:  v.BeforeUpdate.SelfMute,
			SelfVideo: v.BeforeUpdate.SelfVideo,
			Suppress:  v.BeforeUpdate.Suppress,
		}
	}
	return vs
}

// --- Command Conversions (to discordgo) ---

func commandToDiscordgo(cmd *domain.Command) *discordgo.ApplicationCommand {
	if cmd == nil {
		return nil
	}
	return &discordgo.ApplicationCommand{
		ID:            string(cmd.ID),
		ApplicationID: string(cmd.ApplicationID),
		GuildID:       string(cmd.GuildID),
		Name:          cmd.Name,
		Description:   cmd.Description,
		Type:          discordgo.ApplicationCommandType(cmd.Type),
		Options:       commandOptionsToDiscordgo(cmd.Options),
		NSFW:          &cmd.NSFW,
	}
}

func commandOptionsToDiscordgo(opts []*domain.CommandOptionDefinition) []*discordgo.ApplicationCommandOption {
	if len(opts) == 0 {
		return nil
	}
	result := make([]*discordgo.ApplicationCommandOption, len(opts))
	for i, opt := range opts {
		result[i] = &discordgo.ApplicationCommandOption{
			Type:                     discordgo.ApplicationCommandOptionType(opt.Type),
			Name:                     opt.Name,
			Description:              opt.Description,
			Required:                 opt.Required,
			Choices:                  commandChoicesToDiscordgo(opt.Choices),
			Options:                  commandOptionsToDiscordgo(opt.Options),
			Autocomplete:             opt.Autocomplete,
			NameLocalizations:        localizationsToDiscordgo(opt.NameLocalizations),
			DescriptionLocalizations: localizationsToDiscordgo(opt.DescriptionLocalizations),
		}
		// Map channel types
		if len(opt.ChannelTypes) > 0 {
			channelTypes := make([]discordgo.ChannelType, len(opt.ChannelTypes))
			for j, ct := range opt.ChannelTypes {
				channelTypes[j] = discordgo.ChannelType(ct)
			}
			result[i].ChannelTypes = channelTypes
		}
		if opt.MinValue != nil {
			result[i].MinValue = opt.MinValue
		}
		if opt.MaxValue != nil {
			result[i].MaxValue = *opt.MaxValue
		}
		if opt.MinLength != nil {
			result[i].MinLength = opt.MinLength
		}
		if opt.MaxLength != nil {
			result[i].MaxLength = *opt.MaxLength
		}
	}
	return result
}

func commandChoicesToDiscordgo(choices []*domain.CommandChoice) []*discordgo.ApplicationCommandOptionChoice {
	if len(choices) == 0 {
		return nil
	}
	result := make([]*discordgo.ApplicationCommandOptionChoice, len(choices))
	for i, c := range choices {
		result[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:              c.Name,
			Value:             c.Value,
			NameLocalizations: localizationsToDiscordgo(c.NameLocalizations),
		}
	}
	return result
}

func commandFromDiscordgoResponse(cmd *discordgo.ApplicationCommand) *domain.Command {
	if cmd == nil {
		return nil
	}
	return &domain.Command{
		ID:            domain.CommandID(cmd.ID),
		ApplicationID: domain.ApplicationID(cmd.ApplicationID),
		GuildID:       domain.GuildID(cmd.GuildID),
		Name:          cmd.Name,
		Description:   cmd.Description,
		Type:          domain.CommandType(cmd.Type),
		Version:       cmd.Version,
	}
}

// --- Response Conversions ---

func responseToDiscordgo(resp *domain.Response) *discordgo.InteractionResponse {
	if resp == nil {
		return nil
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(resp.Type),
		Data: responseDataToDiscordgo(resp.Data),
	}
}

func responseDataToDiscordgo(data *domain.ResponseData) *discordgo.InteractionResponseData {
	if data == nil {
		return nil
	}
	result := &discordgo.InteractionResponseData{
		TTS:      data.TTS,
		Content:  data.Content,
		Flags:    discordgo.MessageFlags(data.Flags),
		CustomID: data.CustomID,
		Title:    data.Title,
	}
	if len(data.Embeds) > 0 {
		result.Embeds = make([]*discordgo.MessageEmbed, len(data.Embeds))
		for i, e := range data.Embeds {
			result.Embeds[i] = embedToDiscordgo(e)
		}
	}
	if len(data.Components) > 0 {
		result.Components = componentsToDiscordgo(data.Components)
	}
	if len(data.Files) > 0 {
		result.Files = filesToDiscordgo(data.Files)
	}
	if data.AllowedMentions != nil {
		result.AllowedMentions = allowedMentionsToDiscordgo(data.AllowedMentions)
	}
	return result
}

func componentsToDiscordgo(comps []domain.Component) []discordgo.MessageComponent {
	if len(comps) == 0 {
		return nil
	}
	result := make([]discordgo.MessageComponent, len(comps))
	for i, comp := range comps {
		switch c := comp.(type) {
		case *domain.ActionRow:
			result[i] = discordgo.ActionsRow{
				Components: componentsToDiscordgo(c.Components),
			}
		case *domain.Button:
			btn := discordgo.Button{
				Style:    discordgo.ButtonStyle(c.Style),
				Label:    c.Label,
				CustomID: c.CustomID,
				URL:      c.URL,
				Disabled: c.Disabled,
			}
			if c.Emoji != nil {
				btn.Emoji = &discordgo.ComponentEmoji{
					ID:       string(c.Emoji.ID),
					Name:     c.Emoji.Name,
					Animated: c.Emoji.Animated,
				}
			}
			result[i] = btn
		case *domain.SelectMenu:
			sm := discordgo.SelectMenu{
				MenuType:    discordgo.SelectMenuType(c.Type),
				CustomID:    c.CustomID,
				Placeholder: c.Placeholder,
				MaxValues:   c.MaxValues,
				Disabled:    c.Disabled,
			}
			if c.MinValues != nil {
				sm.MinValues = c.MinValues
			}
			// Map channel types for channel select menus
			if len(c.ChannelTypes) > 0 {
				channelTypes := make([]discordgo.ChannelType, len(c.ChannelTypes))
				for j, ct := range c.ChannelTypes {
					channelTypes[j] = discordgo.ChannelType(ct)
				}
				sm.ChannelTypes = channelTypes
			}
			if len(c.Options) > 0 {
				sm.Options = make([]discordgo.SelectMenuOption, len(c.Options))
				for j, opt := range c.Options {
					sm.Options[j] = discordgo.SelectMenuOption{
						Label:       opt.Label,
						Value:       opt.Value,
						Description: opt.Description,
						Default:     opt.Default,
					}
					if opt.Emoji != nil {
						sm.Options[j].Emoji = &discordgo.ComponentEmoji{
							ID:       string(opt.Emoji.ID),
							Name:     opt.Emoji.Name,
							Animated: opt.Emoji.Animated,
						}
					}
				}
			}
			result[i] = sm
		case *domain.TextInput:
			result[i] = discordgo.TextInput{
				CustomID:    c.CustomID,
				Style:       discordgo.TextInputStyle(c.Style),
				Label:       c.Label,
				MinLength:   c.MinLength,
				MaxLength:   c.MaxLength,
				Required:    c.Required,
				Value:       c.Value,
				Placeholder: c.Placeholder,
			}
		}
	}
	return result
}

func filesToDiscordgo(files []*domain.File) []*discordgo.File {
	if len(files) == 0 {
		return nil
	}
	result := make([]*discordgo.File, len(files))
	for i, f := range files {
		result[i] = &discordgo.File{
			Name:        f.Name,
			ContentType: f.ContentType,
			Reader:      f.Reader,
		}
	}
	return result
}

func allowedMentionsToDiscordgo(am *domain.AllowedMentions) *discordgo.MessageAllowedMentions {
	if am == nil {
		return nil
	}
	result := &discordgo.MessageAllowedMentions{
		RepliedUser: am.RepliedUser,
	}
	if len(am.Parse) > 0 {
		result.Parse = make([]discordgo.AllowedMentionType, len(am.Parse))
		for i, p := range am.Parse {
			result.Parse[i] = discordgo.AllowedMentionType(p)
		}
	}
	if len(am.Roles) > 0 {
		result.Roles = make([]string, len(am.Roles))
		for i, r := range am.Roles {
			result.Roles[i] = string(r)
		}
	}
	if len(am.Users) > 0 {
		result.Users = make([]string, len(am.Users))
		for i, u := range am.Users {
			result.Users[i] = string(u)
		}
	}
	return result
}
