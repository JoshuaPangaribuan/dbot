package langchain

import (
	"context"
	"fmt"
	"time"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/tmc/langchaingo/llms"
)

// Service provides LLM operations
type Service interface {
	// Complete - simple text completion
	Complete(ctx context.Context, prompt string, opts ...CompletionOption) (string, error)

	// Chat - conversational completion with history
	Chat(ctx context.Context, messages []Message, opts ...ChatOption) (string, error)

	// Stream - streaming completion
	Stream(ctx context.Context, prompt string, opts ...CompletionOption) (<-chan string, error)

	// Close - cleanup resources
	Close() error
}

// service implements Service
type service struct {
	llm     LLM
	logger  logger.Logger
	config  *config
	timeout time.Duration
}

// New creates a new LangChain service
func New(opts ...Option) (Service, error) {
	cfg := &config{
		provider:    ProviderOpenAI,
		model:       "gpt-3.5-turbo",
		temperature: 0.7,
		maxTokens:   1000,
		timeout:     30 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.logger == nil {
		// Create a default logger if none provided
		cfg.logger = logger.New()
	}

	llm, err := newLLM(cfg)
	if err != nil {
		return nil, fmt.Errorf("create llm: %w", err)
	}

	return &service{
		llm:     llm,
		logger:  cfg.logger,
		config:  cfg,
		timeout: cfg.timeout,
	}, nil
}

// Complete generates a text completion for the given prompt
func (s *service) Complete(ctx context.Context, prompt string, opts ...CompletionOption) (string, error) {
	// Apply completion options
	cfg := &completionConfig{
		temperature: s.config.temperature,
		maxTokens:   s.config.maxTokens,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Add timeout if not already present
	if _, ok := ctx.Deadline(); !ok && s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	s.logger.Debug(ctx, "llm completion requested",
		logger.Fields{
			"provider":      s.config.provider,
			"model":         s.config.model,
			"prompt_length": len(prompt),
		})

	result, err := s.llm.Call(ctx, prompt,
		llms.WithTemperature(float64(cfg.temperature)),
		llms.WithMaxTokens(cfg.maxTokens),
	)
	if err != nil {
		s.logger.Error(ctx, "llm completion failed",
			logger.Fields{"error": err})
		return "", fmt.Errorf("llm call: %w", err)
	}

	s.logger.Debug(ctx, "llm completion succeeded",
		logger.Fields{
			"response_length": len(result),
		})

	return result, nil
}

// Chat generates a chat completion with message history
func (s *service) Chat(ctx context.Context, messages []Message, opts ...ChatOption) (string, error) {
	// Apply chat options
	cfg := &chatConfig{
		temperature: s.config.temperature,
		maxTokens:   s.config.maxTokens,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Add timeout if not already present
	if _, ok := ctx.Deadline(); !ok && s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	// Convert messages to a single prompt for simple completion
	// In a full implementation, this would use the chat API
	prompt := s.buildChatPrompt(messages)

	s.logger.Debug(ctx, "llm chat requested",
		logger.Fields{
			"provider":        s.config.provider,
			"model":           s.config.model,
			"message_count":   len(messages),
			"prompt_length":   len(prompt),
		})

	result, err := s.llm.Call(ctx, prompt,
		llms.WithTemperature(float64(cfg.temperature)),
		llms.WithMaxTokens(cfg.maxTokens),
	)
	if err != nil {
		s.logger.Error(ctx, "llm chat failed",
			logger.Fields{"error": err})
		return "", fmt.Errorf("llm chat: %w", err)
	}

	return result, nil
}

// Stream generates a streaming text completion
func (s *service) Stream(ctx context.Context, prompt string, opts ...CompletionOption) (<-chan string, error) {
	// For now, return a simple implementation
	// Full streaming implementation would use langchaingo's streaming support
	ch := make(chan string, 1)

	go func() {
		defer close(ch)
		result, err := s.Complete(ctx, prompt, opts...)
		if err != nil {
			s.logger.Error(ctx, "stream completion failed",
				logger.Fields{"error": err})
			return
		}
		ch <- result
	}()

	return ch, nil
}

// Close cleans up resources
func (s *service) Close() error {
	// No resources to clean up for now
	return nil
}

// buildChatPrompt converts messages to a single prompt
func (s *service) buildChatPrompt(messages []Message) string {
	prompt := ""
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt += fmt.Sprintf("System: %s\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("User: %s\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("Assistant: %s\n", msg.Content)
		}
	}
	prompt += "Assistant:"
	return prompt
}
