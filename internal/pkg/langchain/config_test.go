package langchain

import (
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestProvider_Table(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		want     string
	}{
		{
			name:     "openai provider",
			provider: ProviderOpenAI,
			want:     "openai",
		},
		{
			name:     "ollama provider",
			provider: ProviderOllama,
			want:     "ollama",
		},
		{
			name:     "anthropic provider",
			provider: ProviderAnthropic,
			want:     "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, string(tt.provider))
		})
	}
}

func TestOptions_WithProvider_Table(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		{
			name:     "openai",
			provider: ProviderOpenAI,
		},
		{
			name:     "ollama",
			provider: ProviderOllama,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithProvider(tt.provider)(cfg)
			require.Equal(t, tt.provider, cfg.provider)
		})
	}
}

func TestOptions_WithModel_Table(t *testing.T) {
	tests := []struct {
		name  string
		model string
	}{
		{
			name:  "gpt-3.5-turbo",
			model: "gpt-3.5-turbo",
		},
		{
			name:  "gpt-4",
			model: "gpt-4",
		},
		{
			name:  "llama2",
			model: "llama2",
		},
		{
			name:  "empty model",
			model: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithModel(tt.model)(cfg)
			require.Equal(t, tt.model, cfg.model)
		})
	}
}

func TestOptions_WithTemperature_Table(t *testing.T) {
	tests := []struct {
		name        string
		temperature float32
	}{
		{
			name:        "zero temperature",
			temperature: 0.0,
		},
		{
			name:        "low temperature",
			temperature: 0.3,
		},
		{
			name:        "medium temperature",
			temperature: 0.7,
		},
		{
			name:        "high temperature",
			temperature: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithTemperature(tt.temperature)(cfg)
			require.Equal(t, tt.temperature, cfg.temperature)
		})
	}
}

func TestOptions_WithMaxTokens_Table(t *testing.T) {
	tests := []struct {
		name   string
		tokens int
	}{
		{
			name:   "small tokens",
			tokens: 100,
		},
		{
			name:   "medium tokens",
			tokens: 1000,
		},
		{
			name:   "large tokens",
			tokens: 4000,
		},
		{
			name:   "zero tokens",
			tokens: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithMaxTokens(tt.tokens)(cfg)
			require.Equal(t, tt.tokens, cfg.maxTokens)
		})
	}
}

func TestOptions_WithLogger_Table(t *testing.T) {
	tests := []struct {
		name   string
		logger logger.Logger
	}{
		{
			name:   "with logger",
			logger: logger.New(),
		},
		{
			name:   "nil logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithLogger(tt.logger)(cfg)
			require.Equal(t, tt.logger, cfg.logger)
		})
	}
}

func TestOptions_WithTimeout_Table(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "1 second",
			timeout: 1 * time.Second,
		},
		{
			name:    "30 seconds",
			timeout: 30 * time.Second,
		},
		{
			name:    "1 minute",
			timeout: 1 * time.Minute,
		},
		{
			name:    "zero timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithTimeout(tt.timeout)(cfg)
			require.Equal(t, tt.timeout, cfg.timeout)
		})
	}
}

func TestOptions_WithOpenAIConfig_Table(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
	}{
		{
			name:    "with api key only",
			apiKey:  "sk-test-key",
			baseURL: "",
		},
		{
			name:    "with api key and base url",
			apiKey:  "sk-test-key",
			baseURL: "https://api.openai.com/v1",
		},
		{
			name:    "with custom base url",
			apiKey:  "sk-test-key",
			baseURL: "https://custom-endpoint.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithOpenAIConfig(tt.apiKey, tt.baseURL)(cfg)
			require.Equal(t, tt.apiKey, cfg.apiKey)
			require.Equal(t, tt.baseURL, cfg.baseURL)
			require.Equal(t, ProviderOpenAI, cfg.provider)
		})
	}
}

func TestOptions_WithOllamaConfig_Table(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "default base url",
			baseURL: "",
		},
		{
			name:    "custom base url",
			baseURL: "http://localhost:11434",
		},
		{
			name:    "remote ollama",
			baseURL: "http://ollama.example.com:11434",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			WithOllamaConfig(tt.baseURL)(cfg)
			require.Equal(t, tt.baseURL, cfg.baseURL)
			require.Equal(t, ProviderOllama, cfg.provider)
		})
	}
}

func TestCompletionOptions_Table(t *testing.T) {
	tests := []struct {
		name  string
		opt   CompletionOption
		check func(*testing.T, *completionConfig)
	}{
		{
			name: "with temperature",
			opt:  WithCompletionTemperature(0.5),
			check: func(t *testing.T, cfg *completionConfig) {
				require.Equal(t, float32(0.5), cfg.temperature)
			},
		},
		{
			name: "with max tokens",
			opt:  WithCompletionMaxTokens(500),
			check: func(t *testing.T, cfg *completionConfig) {
				require.Equal(t, 500, cfg.maxTokens)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &completionConfig{}
			tt.opt(cfg)
			tt.check(t, cfg)
		})
	}
}

func TestChatOptions_Table(t *testing.T) {
	tests := []struct {
		name  string
		opt   ChatOption
		check func(*testing.T, *chatConfig)
	}{
		{
			name: "with temperature",
			opt:  WithChatTemperature(0.8),
			check: func(t *testing.T, cfg *chatConfig) {
				require.Equal(t, float32(0.8), cfg.temperature)
			},
		},
		{
			name: "with max tokens",
			opt:  WithChatMaxTokens(1500),
			check: func(t *testing.T, cfg *chatConfig) {
				require.Equal(t, 1500, cfg.maxTokens)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &chatConfig{}
			tt.opt(cfg)
			tt.check(t, cfg)
		})
	}
}

func TestMessage_Table(t *testing.T) {
	tests := []struct {
		name  string
		msg   Message
		check func(*testing.T, Message)
	}{
		{
			name: "system message",
			msg:  Message{Role: "system", Content: "You are helpful"},
			check: func(t *testing.T, m Message) {
				require.Equal(t, "system", m.Role)
				require.Equal(t, "You are helpful", m.Content)
			},
		},
		{
			name: "user message",
			msg:  Message{Role: "user", Content: "Hello"},
			check: func(t *testing.T, m Message) {
				require.Equal(t, "user", m.Role)
				require.Equal(t, "Hello", m.Content)
			},
		},
		{
			name: "assistant message",
			msg:  Message{Role: "assistant", Content: "Hi there!"},
			check: func(t *testing.T, m Message) {
				require.Equal(t, "assistant", m.Role)
				require.Equal(t, "Hi there!", m.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.msg)
		})
	}
}
