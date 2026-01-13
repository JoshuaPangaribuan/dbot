package langchain

import (
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Provider defines the LLM provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderOllama    Provider = "ollama"
	ProviderAnthropic Provider = "anthropic" // Future support
)

// Message represents a chat message
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// config holds service configuration
type config struct {
	provider     Provider
	logger       logger.Logger
	apiKey       string
	baseURL      string
	model        string
	temperature  float32
	maxTokens    int
	timeout      time.Duration
	systemPrompt string
}

// Option configures the service
type Option func(*config)

// WithProvider sets the LLM provider
func WithProvider(p Provider) Option {
	return func(c *config) {
		c.provider = p
	}
}

// WithModel sets the model name
func WithModel(model string) Option {
	return func(c *config) {
		c.model = model
	}
}

// WithTemperature sets the temperature (0.0 - 1.0)
func WithTemperature(temp float32) Option {
	return func(c *config) {
		c.temperature = temp
	}
}

// WithMaxTokens sets maximum tokens in response
func WithMaxTokens(n int) Option {
	return func(c *config) {
		c.maxTokens = n
	}
}

// WithLogger sets the logger
func WithLogger(log logger.Logger) Option {
	return func(c *config) {
		c.logger = log
	}
}

// WithTimeout sets operation timeout
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithOpenAIConfig sets OpenAI-specific config
func WithOpenAIConfig(apiKey, baseURL string) Option {
	return func(c *config) {
		c.apiKey = apiKey
		c.baseURL = baseURL
		c.provider = ProviderOpenAI
	}
}

// WithOllamaConfig sets Ollama-specific config
func WithOllamaConfig(baseURL string) Option {
	return func(c *config) {
		c.baseURL = baseURL
		c.provider = ProviderOllama
	}
}

// WithAnthropicConfig sets Anthropic-specific config
func WithAnthropicConfig(apiKey, baseURL string) Option {
	return func(c *config) {
		c.apiKey = apiKey
		c.baseURL = baseURL
		c.provider = ProviderAnthropic
	}
}

// WithSystemPrompt sets the system prompt for the LLM
func WithSystemPrompt(prompt string) Option {
	return func(c *config) {
		c.systemPrompt = prompt
	}
}

// completionConfig holds options for a single completion request
type completionConfig struct {
	temperature float32
	maxTokens   int
}

// CompletionOption configures a single completion request
type CompletionOption func(*completionConfig)

// WithCompletionTemperature sets temperature for a single completion
func WithCompletionTemperature(temp float32) CompletionOption {
	return func(cfg *completionConfig) {
		cfg.temperature = temp
	}
}

// WithCompletionMaxTokens sets max tokens for a single completion
func WithCompletionMaxTokens(n int) CompletionOption {
	return func(cfg *completionConfig) {
		cfg.maxTokens = n
	}
}

// chatConfig holds options for a chat request
type chatConfig struct {
	temperature float32
	maxTokens   int
}

// ChatOption configures a chat request
type ChatOption func(*chatConfig)

// WithChatTemperature sets temperature for a chat request
func WithChatTemperature(temp float32) ChatOption {
	return func(cfg *chatConfig) {
		cfg.temperature = temp
	}
}

// WithChatMaxTokens sets max tokens for a chat request
func WithChatMaxTokens(n int) ChatOption {
	return func(cfg *chatConfig) {
		cfg.maxTokens = n
	}
}
