# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Development
- `go run .` - Run the bot directly
- `make build` - Build binary to `bin/dbot`
- `make run` - Alias for `go run .`
- `make test` - Run all tests with build caching
- `make lint` - Run golangci-lint (errcheck, govet, ineffassign, staticcheck, unused)
- `make fmt` - Format code with gofmt

### Docker
- `make docker-build` - Build Docker image with BuildKit
- `make docker-test` - Run tests in Docker build stage
- `make docker-run` - Run Docker container (requires DISCORD_TOKEN env var)

### Single Test
- `go test ./internal/pkg/discord -run TestBotOpen -v` - Run specific test
- `go test ./... -v` - Run all tests verbosely

## Environment Setup

Required environment variables:
- `DISCORD_TOKEN` - Discord bot token (overrides config file)
- `DBOT_CONFIG_PATH` - Optional path to config file (default: `opt/config/config.yaml`)

## Architecture Overview

This is a **feature-based Discord bot framework** built on discordgo with an event bus architecture.

### Core Concepts

**Event Bus Pattern**: All Discord events (messages, commands, components, modals, voice, reactions) flow through a central EventBus. Features subscribe to event types with middleware chains and handlers.

**Feature Lifecycle**: Features are self-contained modules implementing:
- `Init(b *Bot) error` - Called on registration
- `Shutdown(ctx context.Context) error` - Called on bot close
- `Subscriptions() []*Subscription` - Returns event subscriptions

**CommandProvider**: Optional interface for features that provide slash commands. Commands are bulk-synced to Discord during `bot.Open()`.

**Sharding**: Single-binary, multi-shard support with automatic GatewayBot integration. Shards run concurrently in goroutines with bucketed identify delays.

### Key Components

**Bot** (`internal/pkg/discord/bot.go`)
- Wraps discordgo Session with EventBus and StateStore
- Manages lifecycle: New → Open → Close
- Handles shard session cloning and concurrent startup
- Atomic state transitions (new → open → closed)
- Registers features before Open(), syncs commands during Open()

**EventBus** (`internal/pkg/discord/event.go`)
- Priority-ordered subscriptions
- Global middleware + subscription-specific middleware
- Thread-safe subscription management

**Context Wrappers** (`internal/pkg/discord/ctx.go`)
- `CommandContext` - Slash command interactions with option parsing
- `MessageContext` - Message events
- `ComponentContext` - Button/select menu interactions
- `ModalContext` - Modal submissions
- `ReactionAddContext` - Reactions

**Middleware** (`internal/pkg/discord/middleware/`)
- `IgnoreBot()` - Skip bot users
- `RequireGuild()` - Block DMs
- `Chain()` - Combine middleware
- `RateLimit(n, window)` - Sliding window rate limiter
- `RequireNSFW()`, content filters, permission checks

**Reply Builder** (`internal/pkg/discord/reply/`)
- Fluent API for responses: `ctx.Reply().Content("...").Ephemeral().Send()`
- Supports embeds, components, attachments, files

### Application Bootstrap

`main.go` → `app.New()` → `app.Start()`
1. Load config (Viper with hot-reload)
2. Initialize logger (slog or zap provider)
3. Create Bot with sharding options
4. Register global middleware
5. Register features
6. Connect to Discord (opens shards, syncs commands)
7. Wait for SIGINT/SIGTERM
8. Shutdown in reverse order

### Feature Registration Pattern

```go
myfeature.Register(myfeature.Deps{
    Bot:    bot,
    Logger: logger,
    Config: myfeature.Config{...},
})
```

Features subscribe to events and optionally provide commands:
```go
func (f *Feature) Subscriptions() []*discord.Subscription {
    return []*discord.Subscription{
        {
            Type:    discord.EventTypeCommand,
            Handler: discord.CommandFilter("hello", f.handleHello),
            Middleware: []discord.MiddlewareFunc{discord.IgnoreBot()},
        },
    }
}
```

### Sharding Configuration

Controlled via config or env vars:
- `discord.sharding.enabled` / `DISCORD_SHARDING` (bool)
- `discord.sharding.auto` / `DISCORD_AUTO_SHARDING` (bool) - Uses GatewayBot recommended count
- `discord.sharding.count` / `DISCORD_SHARD_COUNT` (int) - Manual shard count
- `discord.sharding.use_gateway_bot` / `DISCORD_USE_GATEWAY_BOT` (bool)
- `discord.sharding.identify_delay` / `DISCORD_IDENTIFY_DELAY` (duration)

When sharding is enabled:
- Shards are grouped into concurrency buckets: `bucket = shard_id % max_concurrency`
- Each bucket starts concurrently; shards within a bucket have delayed identifies
- GatewayBot provides `max_concurrency` (session start limit)

### Configuration System

**Viper Provider** (default): Hot-reload with file watching, debounced 100ms
**OS Provider**: Simple key-value without file watching

Config precedence: Environment variables > Config file

### Testing

Tests use:
- `github.com/stretchr/testify` for assertions
- Mockery for generating mocks (check `internal/mocks/`)
- Proper cleanup in `defer` statements

Run specific package tests: `go test ./internal/pkg/discord -run TestName -v`

### Code Organization

- `internal/app/` - Application wiring, lifecycle
- `internal/features/` - Feature implementations (levelling example)
- `internal/pkg/discord/` - Discord framework (bot, event bus, contexts, middleware, reply)
- `internal/pkg/config/` - Configuration abstraction
- `internal/pkg/logger/` - Logging abstraction (slog/zap providers)
- `internal/pkg/langchain/` - LangChain integration (new)
- `opt/config/` - Configuration files

### Important Patterns

**Context Propagation**: Bot has a base context set via `SetBaseContext()`. All events inherit this context.

**Lock Ordering**: Bot uses lock-free reads where possible (atomic state, RLock for reads). Locks are held briefly to avoid blocking during network calls.

**Session Management**: Primary session (`bot.session`) is shard 0 when sharding, or the only session when not sharding. Use `bot.Sessions()` for all shards.

**Command Sync**: Bulk overwrite on `Open()` is idempotent. Commands are stored from `CommandProvider` features and simple `SlashCommand()` calls.
