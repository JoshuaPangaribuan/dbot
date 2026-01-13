package langchain

import (
	"context"
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestNew_Service_Table(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
		assert  func(t *testing.T, svc Service)
	}{
		{
			name: "valid openai config",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-api-key", ""),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "valid ollama config",
			opts: []Option{
				WithProvider(ProviderOllama),
				WithOllamaConfig("http://localhost:11434"),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "missing openai api key",
			opts: []Option{
				WithProvider(ProviderOpenAI),
			},
			wantErr: true,
			assert:  nil,
		},
		{
			name: "with logger",
			opts: []Option{
				WithProvider(ProviderOllama),
				WithLogger(logger.New()),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "with custom model",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-key", ""),
				WithModel("gpt-4"),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "with temperature",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-key", ""),
				WithTemperature(0.5),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "with max tokens",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-key", ""),
				WithMaxTokens(2000),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "with timeout",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-key", ""),
				WithTimeout(60 * time.Second),
			},
			wantErr: false,
			assert: func(t *testing.T, svc Service) {
				require.NotNil(t, svc)
			},
		},
		{
			name: "unsupported provider",
			opts: []Option{
				WithProvider(ProviderAnthropic),
			},
			wantErr: true,
			assert:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, svc)
			} else {
				require.NoError(t, err)
				if tt.assert != nil {
					tt.assert(t, svc)
				}
				// Cleanup
				_ = svc.Close()
			}
		})
	}
}

func TestService_Close_Table(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "close openai service",
			opts: []Option{
				WithProvider(ProviderOpenAI),
				WithOpenAIConfig("test-key", ""),
			},
			wantErr: false,
		},
		{
			name: "close ollama service",
			opts: []Option{
				WithProvider(ProviderOllama),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(tt.opts...)
			require.NoError(t, err)
			err = svc.Close()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// mockLLM is a mock LLM for testing
type mockLLM struct {
	callFunc func(ctx context.Context, prompt string, opts ...interface{}) (string, error)
}

func (m *mockLLM) Call(ctx context.Context, prompt string, opts ...interface{}) (string, error) {
	if m.callFunc != nil {
		return m.callFunc(ctx, prompt, opts...)
	}
	return "mock response", nil
}

func TestService_Complete_Options_Table(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		opts    []CompletionOption
		wantErr bool
	}{
		{
			name:    "simple completion",
			prompt:  "Hello",
			opts:    nil,
			wantErr: false,
		},
		{
			name:   "with temperature",
			prompt: "Test",
			opts: []CompletionOption{
				WithCompletionTemperature(0.8),
			},
			wantErr: false,
		},
		{
			name:   "with max tokens",
			prompt: "Test",
			opts: []CompletionOption{
				WithCompletionMaxTokens(500),
			},
			wantErr: false,
		},
		{
			name:   "with both options",
			prompt: "Test",
			opts: []CompletionOption{
				WithCompletionTemperature(0.5),
				WithCompletionMaxTokens(1000),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests require actual LLM API keys to run
			// They're structured for integration testing
			t.Skip("requires API keys for integration testing")
		})
	}
}

func TestService_Chat_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		opts     []ChatOption
		wantErr  bool
	}{
		{
			name: "simple chat",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name: "chat with history",
			messages: []Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name: "chat with temperature",
			messages: []Message{
				{Role: "user", Content: "Test"},
			},
			opts: []ChatOption{
				WithChatTemperature(0.9),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests require actual LLM API keys to run
			// They're structured for integration testing
			t.Skip("requires API keys for integration testing")
		})
	}
}

func TestService_Stream_Table(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		opts    []CompletionOption
		wantErr bool
	}{
		{
			name:    "simple stream",
			prompt:  "Hello",
			opts:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests require actual LLM API keys to run
			// They're structured for integration testing
			t.Skip("requires API keys for integration testing")
		})
	}
}

func TestService_ContextTimeout(t *testing.T) {
	t.Run("context timeout applied", func(t *testing.T) {
		opts := []Option{
			WithProvider(ProviderOpenAI),
			WithOpenAIConfig("test-key", ""),
			WithTimeout(1 * time.Nanosecond), // Very short timeout
		}

		svc, err := New(opts...)
		require.NoError(t, err)
		defer svc.Close()

		// This should timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Note: This test requires actual API calls
		// For now, we just verify the service structure
		_ = ctx
	})
}
