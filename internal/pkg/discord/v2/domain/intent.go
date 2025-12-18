package domain

// Intent represents Discord gateway intents.
// These are v2-owned flags that adapters translate to library-specific constants.
type Intent uint64

const (
	// IntentGuilds enables guild-related events.
	IntentGuilds Intent = 1 << iota

	// IntentGuildMembers enables guild member events (privileged).
	IntentGuildMembers

	// IntentGuildModeration enables guild moderation events.
	IntentGuildModeration

	// IntentGuildEmojisAndStickers enables emoji and sticker events.
	IntentGuildEmojisAndStickers

	// IntentGuildIntegrations enables integration events.
	IntentGuildIntegrations

	// IntentGuildWebhooks enables webhook events.
	IntentGuildWebhooks

	// IntentGuildInvites enables invite events.
	IntentGuildInvites

	// IntentGuildVoiceStates enables voice state events.
	IntentGuildVoiceStates

	// IntentGuildPresences enables presence events (privileged).
	IntentGuildPresences

	// IntentGuildMessages enables guild message events.
	IntentGuildMessages

	// IntentGuildMessageReactions enables guild message reaction events.
	IntentGuildMessageReactions

	// IntentGuildMessageTyping enables guild typing events.
	IntentGuildMessageTyping

	// IntentDirectMessages enables DM events.
	IntentDirectMessages

	// IntentDirectMessageReactions enables DM reaction events.
	IntentDirectMessageReactions

	// IntentDirectMessageTyping enables DM typing events.
	IntentDirectMessageTyping

	// IntentMessageContent enables message content (privileged).
	IntentMessageContent

	// IntentGuildScheduledEvents enables scheduled event events.
	IntentGuildScheduledEvents

	// IntentAutoModerationConfiguration enables auto-mod config events.
	IntentAutoModerationConfiguration

	// IntentAutoModerationExecution enables auto-mod execution events.
	IntentAutoModerationExecution
)

// IntentsDefault is the set of non-privileged intents commonly used.
var IntentsDefault = IntentGuilds |
	IntentGuildMessages |
	IntentGuildMessageReactions |
	IntentGuildVoiceStates |
	IntentDirectMessages |
	IntentDirectMessageReactions

// IntentsAll includes all intents (including privileged).
var IntentsAll = IntentsDefault |
	IntentGuildMembers |
	IntentGuildPresences |
	IntentMessageContent |
	IntentGuildModeration |
	IntentGuildEmojisAndStickers |
	IntentGuildIntegrations |
	IntentGuildWebhooks |
	IntentGuildInvites |
	IntentGuildMessageTyping |
	IntentDirectMessageTyping |
	IntentGuildScheduledEvents |
	IntentAutoModerationConfiguration |
	IntentAutoModerationExecution

// Has returns true if the intent set contains the given intent.
func (i Intent) Has(intent Intent) bool {
	return i&intent == intent
}

// Add adds an intent to the set.
func (i Intent) Add(intent Intent) Intent {
	return i | intent
}

// Remove removes an intent from the set.
func (i Intent) Remove(intent Intent) Intent {
	return i &^ intent
}
