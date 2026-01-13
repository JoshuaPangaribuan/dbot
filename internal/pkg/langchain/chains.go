package langchain

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

// ChainBuilder helps construct LangChain chains
type ChainBuilder struct {
	llm    llms.Model
	logger logger.Logger
}

// NewChainBuilder creates a chain builder from a Service
func NewChainBuilder(svc Service) *ChainBuilder {
	// Extract the internal service structure to access the LLM
	s, ok := svc.(*service)
	if !ok {
		return nil
	}

	// Get the underlying model from the wrapper
	wrapper, ok := s.llm.(*llmWrapper)
	if !ok {
		return nil
	}

	return &ChainBuilder{
		llm:    wrapper.model,
		logger: s.logger,
	}
}

// SimpleChain creates a simple prompt->LLM chain
func (b *ChainBuilder) SimpleChain(template string, inputVars []string) (*Chain, error) {
	prompt := prompts.NewPromptTemplate(template, inputVars)

	// Create the chain using the LLM model
	chain := chains.NewLLMChain(b.llm, prompt)

	return &Chain{
		chain:  chain,
		logger: b.logger,
	}, nil
}

// SequentialChain creates a chain that runs multiple chains in sequence
func (b *ChainBuilder) SequentialChain(chains ...*Chain) (*Chain, error) {
	// Implementation for sequential chains
	// This would use chains.NewSequentialChain or similar
	return nil, fmt.Errorf("sequential chains not yet implemented")
}

// Chain wraps langchaingo chains
type Chain struct {
	chain  chains.Chain
	logger logger.Logger
}

// Run executes the chain with the given input
func (c *Chain) Run(ctx context.Context, input map[string]any) (string, error) {
	// Execute the chain with the provided input
	result, err := chains.Run(ctx, c.chain, input)
	if err != nil {
		c.logger.Error(ctx, "chain execution failed",
			logger.Fields{"error": err})
		return "", fmt.Errorf("chain run: %w", err)
	}

	return result, nil
}

// Apply executes the chain and returns the full output map
func (c *Chain) Apply(ctx context.Context, input map[string]any) (map[string]any, error) {
	// Implementation for Apply would use chains.Call or similar
	result, err := c.Run(ctx, input)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"output": result,
	}, nil
}
