package config

import (
	"fmt"
)

// Provider defines the configuration provider type
type Provider string

const (
	ProviderOS    Provider = "os"
	ProviderViper Provider = "viper"
)

type Config interface {
	GetInt(key string) int
	GetBool(key string) bool
	GetInt64(key string) int64
	GetFloat64(key string) float64
	GetString(key string) string
	GetSlice(key string) []any
	GetMap(key string) map[string]any
	Close() error
}

// configSettings holds configuration options
type configSettings struct {
	files    []string
	watch    bool
	provider Provider
	onError  func(error)
}

// Option serves as the option pattern for configuration
type Option func(*configSettings)

// WithFiles adds files to be loaded
func WithFiles(paths ...string) Option {
	return func(c *configSettings) {
		c.files = append(c.files, paths...)
	}
}

// WithWatcher enables or disables the file watcher
func WithWatcher(enabled bool) Option {
	return func(c *configSettings) {
		c.watch = enabled
	}
}

// WithProvider sets the configuration provider
func WithProvider(p Provider) Option {
	return func(c *configSettings) {
		c.provider = p
	}
}

// WithErrorHandler sets a callback for watcher errors
func WithErrorHandler(fn func(error)) Option {
	return func(c *configSettings) {
		c.onError = fn
	}
}

// New creates a new Config instance
func New(opts ...Option) (Config, error) {
	cfg := &configSettings{
		watch:    true,
		provider: ProviderOS,
		onError:  func(err error) { fmt.Printf("config error: %v\n", err) },
	}

	for _, opt := range opts {
		opt(cfg)
	}

	switch cfg.provider {
	case ProviderViper:
		return newViperConfig(cfg)
	case ProviderOS:
		return newOSConfig(cfg)
	default:
		return newOSConfig(cfg)
	}
}
