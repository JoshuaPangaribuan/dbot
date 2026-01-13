package langchain

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// LLM interface wraps langchaingo llms.Model for internal use
type LLM interface {
	Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)
}

// llmWrapper wraps langchaingo LLM implementations to provide the Call method
type llmWrapper struct {
	model llms.Model
}

// newLLM creates an LLM instance based on provider
func newLLM(cfg *config) (LLM, error) {
	switch cfg.provider {
	case ProviderOpenAI:
		return newOpenAILLM(cfg)
	case ProviderOllama:
		return newOllamaLLM(cfg)
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

	return &llmWrapper{model: model}, nil
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

	return &llmWrapper{model: model}, nil
}

// Call executes the LLM with the given prompt
func (w *llmWrapper) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	// Generate completion using the model
	return llms.GenerateFromSinglePrompt(ctx, w.model, prompt, opts...)
}
