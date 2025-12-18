package discordgo

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/port"
	"github.com/bwmarrin/discordgo"
)

// --- RestClient Interface Implementation ---

// BulkOverwriteCommands overwrites all commands for the application.
func (a *Adapter) BulkOverwriteCommands(ctx context.Context, guildID domain.GuildID, cmds []*domain.Command) ([]*domain.Command, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	// Get application ID
	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	if appID.IsEmpty() {
		// Try to get from session state
		if a.session.State != nil && a.session.State.User != nil {
			appID = domain.ApplicationID(a.session.State.User.ID)
		}
		if appID.IsEmpty() {
			// Fetch via API
			user, err := a.session.User("@me")
			if err != nil {
				return nil, fmt.Errorf("discordgo: fetch bot user: %w", err)
			}
			appID = domain.ApplicationID(user.ID)
		}
	}

	// Convert commands
	dgCmds := make([]*discordgo.ApplicationCommand, len(cmds))
	for i, cmd := range cmds {
		dgCmds[i] = commandToDiscordgo(cmd)
	}

	// Bulk overwrite
	registered, err := a.session.ApplicationCommandBulkOverwrite(string(appID), string(guildID), dgCmds)
	if err != nil {
		return nil, fmt.Errorf("discordgo: bulk overwrite commands: %w", err)
	}

	// Convert response
	result := make([]*domain.Command, len(registered))
	for i, cmd := range registered {
		result[i] = commandFromDiscordgoResponse(cmd)
	}

	return result, nil
}

// DeleteCommand deletes a command.
func (a *Adapter) DeleteCommand(ctx context.Context, guildID domain.GuildID, cmdID domain.CommandID) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	if appID.IsEmpty() {
		return fmt.Errorf("discordgo: application ID not available")
	}

	return a.session.ApplicationCommandDelete(string(appID), string(guildID), string(cmdID))
}

// RespondInteraction sends an initial response to an interaction.
func (a *Adapter) RespondInteraction(ctx context.Context, interactionID domain.InteractionID, token string, resp *domain.Response) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	interaction := &discordgo.Interaction{
		ID:    string(interactionID),
		Token: token,
	}

	err := a.session.InteractionRespond(interaction, responseToDiscordgo(resp))
	if err != nil {
		return fmt.Errorf("discordgo: respond interaction: %w", err)
	}
	return nil
}

// DeferInteraction sends a deferred response to an interaction.
func (a *Adapter) DeferInteraction(ctx context.Context, interactionID domain.InteractionID, token string, flags domain.ResponseFlags) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	interaction := &discordgo.Interaction{
		ID:    string(interactionID),
		Token: token,
	}

	err := a.session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlags(flags),
		},
	})
	if err != nil {
		return fmt.Errorf("discordgo: defer interaction: %w", err)
	}
	return nil
}

// EditInteractionResponse edits the initial response to an interaction.
func (a *Adapter) EditInteractionResponse(ctx context.Context, token string, edit *domain.ResponseEdit) (*domain.Message, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	webhookEdit := &discordgo.WebhookEdit{}
	if edit.Content != nil {
		webhookEdit.Content = edit.Content
	}
	if edit.Embeds != nil {
		embeds := make([]*discordgo.MessageEmbed, len(*edit.Embeds))
		for i, e := range *edit.Embeds {
			embeds[i] = embedToDiscordgo(e)
		}
		webhookEdit.Embeds = &embeds
	}
	if edit.Components != nil {
		comps := componentsToDiscordgo(*edit.Components)
		webhookEdit.Components = &comps
	}
	if edit.Files != nil {
		webhookEdit.Files = filesToDiscordgo(edit.Files)
	}
	if edit.AllowedMentions != nil {
		webhookEdit.AllowedMentions = allowedMentionsToDiscordgo(edit.AllowedMentions)
	}

	// Need to use the right interaction for webhook edit
	interaction := &discordgo.Interaction{
		AppID: string(appID),
		Token: token,
	}

	msg, err := a.session.InteractionResponseEdit(interaction, webhookEdit)
	if err != nil {
		return nil, fmt.Errorf("discordgo: edit interaction response: %w", err)
	}

	return messageFromDiscordgo(msg), nil
}

// GetInteractionResponse gets the initial response to an interaction.
func (a *Adapter) GetInteractionResponse(ctx context.Context, token string) (*domain.Message, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	interaction := &discordgo.Interaction{
		AppID: string(appID),
		Token: token,
	}

	msg, err := a.session.InteractionResponse(interaction)
	if err != nil {
		return nil, fmt.Errorf("discordgo: get interaction response: %w", err)
	}

	return messageFromDiscordgo(msg), nil
}

// DeleteInteractionResponse deletes the initial response to an interaction.
func (a *Adapter) DeleteInteractionResponse(ctx context.Context, token string) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	interaction := &discordgo.Interaction{
		AppID: string(appID),
		Token: token,
	}

	return a.session.InteractionResponseDelete(interaction)
}

// FollowupMessage sends a followup message to an interaction.
func (a *Adapter) FollowupMessage(ctx context.Context, token string, msg *domain.ResponseData) (*domain.Message, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	a.mu.RLock()
	appID := a.appID
	a.mu.RUnlock()

	interaction := &discordgo.Interaction{
		AppID: string(appID),
		Token: token,
	}

	params := &discordgo.WebhookParams{
		TTS:     msg.TTS,
		Content: msg.Content,
		Flags:   discordgo.MessageFlags(msg.Flags),
	}
	if len(msg.Embeds) > 0 {
		params.Embeds = make([]*discordgo.MessageEmbed, len(msg.Embeds))
		for i, e := range msg.Embeds {
			params.Embeds[i] = embedToDiscordgo(e)
		}
	}
	if len(msg.Components) > 0 {
		params.Components = componentsToDiscordgo(msg.Components)
	}
	if len(msg.Files) > 0 {
		params.Files = filesToDiscordgo(msg.Files)
	}
	if msg.AllowedMentions != nil {
		params.AllowedMentions = allowedMentionsToDiscordgo(msg.AllowedMentions)
	}

	result, err := a.session.FollowupMessageCreate(interaction, true, params)
	if err != nil {
		return nil, fmt.Errorf("discordgo: create followup message: %w", err)
	}

	return messageFromDiscordgo(result), nil
}

// SendMessage sends a message to a channel.
func (a *Adapter) SendMessage(ctx context.Context, channelID domain.ChannelID, msg *domain.MessageSend) (*domain.Message, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	send := &discordgo.MessageSend{
		Content: msg.Content,
		TTS:     msg.TTS,
	}
	if len(msg.Embeds) > 0 {
		send.Embeds = make([]*discordgo.MessageEmbed, len(msg.Embeds))
		for i, e := range msg.Embeds {
			send.Embeds[i] = embedToDiscordgo(e)
		}
	}
	if len(msg.Components) > 0 {
		send.Components = componentsToDiscordgo(msg.Components)
	}
	if len(msg.Files) > 0 {
		send.Files = filesToDiscordgo(msg.Files)
	}
	if msg.AllowedMentions != nil {
		send.AllowedMentions = allowedMentionsToDiscordgo(msg.AllowedMentions)
	}
	if msg.Reference != nil {
		send.Reference = &discordgo.MessageReference{
			MessageID: string(msg.Reference.MessageID),
			ChannelID: string(msg.Reference.ChannelID),
			GuildID:   string(msg.Reference.GuildID),
		}
	}

	result, err := a.session.ChannelMessageSendComplex(string(channelID), send)
	if err != nil {
		return nil, fmt.Errorf("discordgo: send message: %w", err)
	}

	return messageFromDiscordgo(result), nil
}

// EditMessage edits a message.
func (a *Adapter) EditMessage(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, edit *domain.MessageEdit) (*domain.Message, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	msgEdit := &discordgo.MessageEdit{
		Channel: string(channelID),
		ID:      string(msgID),
	}
	if edit.Content != nil {
		msgEdit.Content = edit.Content
	}
	if edit.Embeds != nil {
		embeds := make([]*discordgo.MessageEmbed, len(*edit.Embeds))
		for i, e := range *edit.Embeds {
			embeds[i] = embedToDiscordgo(e)
		}
		msgEdit.Embeds = &embeds
	}
	if edit.Components != nil {
		comps := componentsToDiscordgo(*edit.Components)
		msgEdit.Components = &comps
	}
	if edit.AllowedMentions != nil {
		msgEdit.AllowedMentions = allowedMentionsToDiscordgo(edit.AllowedMentions)
	}

	result, err := a.session.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return nil, fmt.Errorf("discordgo: edit message: %w", err)
	}

	return messageFromDiscordgo(result), nil
}

// DeleteMessage deletes a message.
func (a *Adapter) DeleteMessage(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	return a.session.ChannelMessageDelete(string(channelID), string(msgID))
}

// AddReaction adds a reaction to a message.
func (a *Adapter) AddReaction(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, emoji string) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	return a.session.MessageReactionAdd(string(channelID), string(msgID), emoji)
}

// RemoveReaction removes the bot's reaction from a message.
func (a *Adapter) RemoveReaction(ctx context.Context, channelID domain.ChannelID, msgID domain.MessageID, emoji string) error {
	if !a.connected.Load() {
		return port.ErrNotConnected
	}

	return a.session.MessageReactionRemove(string(channelID), string(msgID), emoji, "@me")
}

// GetChannel gets a channel by ID.
func (a *Adapter) GetChannel(ctx context.Context, channelID domain.ChannelID) (*domain.Channel, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	ch, err := a.session.Channel(string(channelID))
	if err != nil {
		return nil, fmt.Errorf("discordgo: get channel: %w", err)
	}

	return channelFromDiscordgo(ch), nil
}

// GetGuildRoles gets all roles in a guild.
func (a *Adapter) GetGuildRoles(ctx context.Context, guildID domain.GuildID) ([]*domain.Role, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	roles, err := a.session.GuildRoles(string(guildID))
	if err != nil {
		return nil, fmt.Errorf("discordgo: get guild roles: %w", err)
	}

	result := make([]*domain.Role, len(roles))
	for i, r := range roles {
		result[i] = roleFromDiscordgo(r)
	}

	return result, nil
}

// GetUser gets a user by ID.
func (a *Adapter) GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	user, err := a.session.User(string(userID))
	if err != nil {
		return nil, fmt.Errorf("discordgo: get user: %w", err)
	}

	return userFromDiscordgo(user), nil
}

// GetMember gets a guild member.
func (a *Adapter) GetMember(ctx context.Context, guildID domain.GuildID, userID domain.UserID) (*domain.Member, error) {
	if !a.connected.Load() {
		return nil, port.ErrNotConnected
	}

	member, err := a.session.GuildMember(string(guildID), string(userID))
	if err != nil {
		return nil, fmt.Errorf("discordgo: get member: %w", err)
	}

	return memberFromDiscordgo(member, string(guildID)), nil
}
