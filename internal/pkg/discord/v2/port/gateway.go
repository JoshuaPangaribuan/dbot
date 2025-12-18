package port

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord/v2/domain"
)

// Gateway defines the interface for Discord gateway connections.
// Adapters implement this to connect to Discord's gateway and receive events.
type Gateway interface {
	// Open establishes the connection to Discord's gateway.
	// Returns an error if already connected or if connection fails.
	Open(ctx context.Context) error

	// Close disconnects from the gateway gracefully and marks it as permanently closed.
	Close() error

	// Disconnect disconnects from the gateway without marking it as permanently closed.
	// This allows the gateway to be reopened with Open().
	Disconnect() error

	// SetEventHandler sets the callback function for received events.
	// This must be called before Open().
	SetEventHandler(handler EventHandler)

	// BotUser returns the bot's user after connection is established.
	// Returns nil if not connected.
	BotUser() *domain.User

	// AppID returns the application ID after connection is established.
	// Returns empty if not connected.
	AppID() domain.ApplicationID

	// Latency returns the gateway latency (heartbeat RTT).
	Latency() int64
}

// EventHandler is called by the gateway when events are received.
type EventHandler func(event *domain.Event)

// GatewayStatus represents the connection status of the gateway.
type GatewayStatus int

const (
	GatewayStatusDisconnected GatewayStatus = iota
	GatewayStatusConnecting
	GatewayStatusConnected
	GatewayStatusReconnecting
	GatewayStatusClosing
	GatewayStatusClosed
)

// String returns the string representation of the gateway status.
func (s GatewayStatus) String() string {
	switch s {
	case GatewayStatusDisconnected:
		return "disconnected"
	case GatewayStatusConnecting:
		return "connecting"
	case GatewayStatusConnected:
		return "connected"
	case GatewayStatusReconnecting:
		return "reconnecting"
	case GatewayStatusClosing:
		return "closing"
	case GatewayStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}
