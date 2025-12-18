package logger

import (
	"context"
	"io"
	"os"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = 0
	LogLevelInfo  LogLevel = 1
	LogLevelWarn  LogLevel = 2
	LogLevelError LogLevel = 3
)

// Fields is a type alias for map[string]any
type Fields map[string]any

// Logger is the interface that defines the logging methods
type Logger interface {
	Info(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Debug(ctx context.Context, msg string, fields Fields)
	Close() error
}

type Provider string

const (
	ProviderSlog Provider = "slog"
	ProviderZap  Provider = "zap"
)

// config holds hierarchy of configuration options
type config struct {
	level    LogLevel
	output   io.Writer
	provider Provider
	ctxFn    ContextFieldsFunc
}

// Option serves as the option pattern for configuration
type Option func(*config)

// WithLevel sets the log level
func WithLevel(level LogLevel) Option {
	return func(c *config) {
		switch level {
		case LogLevelDebug:
			c.level = LogLevelDebug
		case LogLevelInfo:
			c.level = LogLevelInfo
		case LogLevelWarn:
			c.level = LogLevelWarn
		case LogLevelError:
			c.level = LogLevelError
		default:
			c.level = LogLevelInfo
		}
	}
}

// WithOutput sets the output destination
func WithOutput(w io.Writer) Option {
	return func(c *config) {
		c.output = w
	}
}

// WithProvider sets the logger provider (slog or zap)
func WithProvider(p Provider) Option {
	return func(c *config) {
		c.provider = p
	}
}

// WithContextFields injects context-derived fields into every log call.
func WithContextFields(fn ContextFieldsFunc) Option {
	return func(c *config) {
		c.ctxFn = fn
	}
}

// New creates a new Logger instance
func New(opts ...Option) Logger {
	cfg := &config{
		level:    LogLevelInfo,
		output:   os.Stdout,
		provider: ProviderSlog,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	switch cfg.provider {
	case ProviderZap:
		return newZapLogger(cfg)
	case ProviderSlog:
		return newSlogLogger(cfg)
	default:
		return newSlogLogger(cfg)
	}
}
