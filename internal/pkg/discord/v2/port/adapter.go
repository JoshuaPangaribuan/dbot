package port

import "github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"

// Adapter combines Gateway and RestClient interfaces.
// This is the main interface that adapter implementations should satisfy.
type Adapter interface {
	Gateway
	RestClient
}

// AdapterConfig contains configuration for creating an adapter.
type AdapterConfig struct {
	// Token is the bot token.
	Token string

	// Intents specifies which gateway intents to request.
	Intents domain.Intent

	// ShardID is the shard ID (for sharding).
	ShardID int

	// ShardCount is the total number of shards.
	ShardCount int
}

// DefaultIntents returns the default set of intents for most bots.
// This delegates to domain.IntentsDefault which excludes privileged intents
// (MessageContent, GuildMembers, GuildPresences) that require explicit opt-in.
func DefaultIntents() domain.Intent {
	return domain.IntentsDefault
}
