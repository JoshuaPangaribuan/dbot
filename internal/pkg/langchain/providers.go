package langchain

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// LLM interface wraps langchaingo llms.Model for internal use
type LLM interface {
	Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)
	Stream(ctx context.Context, prompt string, opts ...llms.CallOption) (<-chan string, error)
}

// llmWrapper wraps langchaingo LLM implementations to provide the Call method
type llmWrapper struct {
	model        llms.Model
	systemPrompt string
}

// newLLM creates an LLM instance based on provider
func newLLM(cfg *config) (LLM, error) {
	switch cfg.provider {
	case ProviderOpenAI:
		return newOpenAILLM(cfg)
	case ProviderOllama:
		return newOllamaLLM(cfg)
	case ProviderAnthropic:
		return newAnthropicLLM(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.provider)
	}
}

// newOpenAILLM creates OpenAI LLM client
func newOpenAILLM(cfg *config) (LLM, error) {
	if cfg.apiKey == "" {
		return nil, errors.New("openai: api key required")
	}

	opts := []openai.Option{
		openai.WithToken(cfg.apiKey),
	}

	if cfg.model != "" {
		opts = append(opts, openai.WithModel(cfg.model))
	}

	if cfg.baseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.baseURL))
	}

	model, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("openai init: %w", err)
	}

	return &llmWrapper{model: model, systemPrompt: cfg.systemPrompt}, nil
}

// newOllamaLLM creates Ollama LLM client
func newOllamaLLM(cfg *config) (LLM, error) {
	baseURL := cfg.baseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	opts := []ollama.Option{}

	if cfg.model != "" {
		opts = append(opts, ollama.WithModel(cfg.model))
	}

	model, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("ollama init: %w", err)
	}

	return &llmWrapper{model: model, systemPrompt: cfg.systemPrompt}, nil
}

// newAnthropicLLM creates Anthropic LLM client
func newAnthropicLLM(cfg *config) (LLM, error) {
	if cfg.apiKey == "" {
		return nil, errors.New("anthropic: api key required")
	}

	opts := []anthropic.Option{
		anthropic.WithToken(cfg.apiKey),
	}

	if cfg.model != "" {
		opts = append(opts, anthropic.WithModel(cfg.model))
	}

	if cfg.baseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(cfg.baseURL))
	}

	model, err := anthropic.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("anthropic init: %w", err)
	}

	return &llmWrapper{model: model, systemPrompt: cfg.systemPrompt}, nil
}

// Call executes the LLM with the given prompt
func (w *llmWrapper) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	// Build message content with optional system prompt
	messages := []llms.MessageContent{}

	if w.systemPrompt != "" {
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, w.systemPrompt))
	}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	// Generate completion using chat-style messages
	resp, err := w.model.GenerateContent(ctx, messages, opts...)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Content, nil
}

// Stream executes the LLM with streaming response
func (w *llmWrapper) Stream(ctx context.Context, prompt string, opts ...llms.CallOption) (<-chan string, error) {
	ch := make(chan string, 10)

	go func() {
		defer close(ch)

		// Build message content with optional system prompt
		messages := []llms.MessageContent{}

		if w.systemPrompt != "" {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, w.systemPrompt))
		}

		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

		// Build options with streaming callback
		allOpts := append([]llms.CallOption{
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				select {
				case ch <- string(chunk):
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			}),
		}, opts...)

		// Use GenerateContent with streaming callback
		_, err := w.model.GenerateContent(ctx, messages, allOpts...)

		if err != nil {
			return
		}
	}()

	return ch, nil
}
