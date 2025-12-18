// Package port defines the adapter interfaces (ports) for the v2 Discord framework.
// Adapters implement these interfaces to connect v2 to specific Discord libraries.
package port

import "errors"

// Common errors returned by adapters.
var (
	// ErrNotConnected is returned when an operation requires a connection but the adapter is not connected.
	ErrNotConnected = errors.New("discord: not connected")

	// ErrAlreadyConnected is returned when trying to connect an already connected adapter.
	ErrAlreadyConnected = errors.New("discord: already connected")

	// ErrClosed is returned when an operation is attempted on a closed adapter.
	ErrClosed = errors.New("discord: adapter closed")

	// ErrRateLimited is returned when Discord rate limits the request.
	ErrRateLimited = errors.New("discord: rate limited")

	// ErrUnauthorized is returned when the token is invalid or lacks permissions.
	ErrUnauthorized = errors.New("discord: unauthorized")

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("discord: not found")

	// ErrForbidden is returned when the bot lacks permission for an operation.
	ErrForbidden = errors.New("discord: forbidden")

	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("discord: bad request")

	// ErrInteractionAlreadyResponded is returned when trying to respond to an interaction that was already responded to.
	ErrInteractionAlreadyResponded = errors.New("discord: interaction already responded")

	// ErrInteractionExpired is returned when an interaction token has expired (15 minutes).
	ErrInteractionExpired = errors.New("discord: interaction expired")
)

// APIError represents a Discord API error with additional context.
type APIError struct {
	Code       int
	Message    string
	StatusCode int
	Err        error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "discord: api error"
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new API error.
func NewAPIError(statusCode, code int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}
