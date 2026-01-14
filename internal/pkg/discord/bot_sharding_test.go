package discord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBot_buildShardSessions_SingleShard(t *testing.T) {
	bot, err := New("test-token", WithGatewayBot(false))
	require.NoError(t, err)

	sessions, err := bot.buildShardSessions()
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	assert.Same(t, bot.session, sessions[0])
	assert.Equal(t, 0, sessions[0].ShardID)
	assert.Equal(t, 1, sessions[0].ShardCount)
}

func TestBot_buildShardSessions_MultiShard(t *testing.T) {
	bot, err := New(
		"test-token",
		WithShardCount(3),
		WithGatewayBot(false),
		WithShardMaxConcurrency(2),
		WithIdentifyDelay(123*time.Millisecond),
	)
	require.NoError(t, err)

	sessions, err := bot.buildShardSessions()
	require.NoError(t, err)
	require.Len(t, sessions, 3)

	assert.Equal(t, 0, sessions[0].ShardID)
	assert.Equal(t, 1, sessions[1].ShardID)
	assert.Equal(t, 2, sessions[2].ShardID)

	for _, s := range sessions {
		assert.Equal(t, 3, s.ShardCount)
	}

	assert.Same(t, bot.session, sessions[0])
	assert.NotSame(t, sessions[0], sessions[1])
	assert.NotSame(t, sessions[0], sessions[2])

	bot.mu.RLock()
	assert.Equal(t, 3, bot.shardCount)
	assert.Equal(t, 2, bot.maxConcurrency)
	assert.Equal(t, 123*time.Millisecond, bot.identifyDelay)
	bot.mu.RUnlock()
}

func TestBot_buildShardSessions_AutoShardingRequiresGatewayBot(t *testing.T) {
	bot, err := New("test-token", WithAutoSharding(), WithGatewayBot(false))
	require.NoError(t, err)

	_, err = bot.buildShardSessions()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auto sharding requires GatewayBot")
}
