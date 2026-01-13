package langchain

import (
	"context"
	"fmt"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Agent represents an LLM agent with tool access
type Agent struct {
	llm    LLM
	tools  []Tool
	logger logger.Logger
}

// Tool represents a function the agent can call
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns what the tool does
	Description() string

	// Call executes the tool with the given input
	Call(ctx context.Context, input string) (string, error)
}

// NewAgent creates an agent with tools
func NewAgent(svc Service, tools ...Tool) *Agent {
	s, ok := svc.(*service)
	if !ok {
		return nil
	}

	return &Agent{
		llm:    s.llm,
		tools:  tools,
		logger: s.logger,
	}
}

// Run executes the agent with the given input
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	// Build the system prompt with tool descriptions
	systemPrompt := a.buildSystemPrompt()

	// Simple agent implementation:
	// 1. LLM decides which tool to use (or answer directly)
	// 2. Call tool if needed
	// 3. Iterate until final answer

	// For now, we'll do a single pass
	// In a full implementation, this would be a loop with proper LLM-based tool selection

	// Check if any tool is relevant
	var toolResult string
	var usedTool bool
	for _, tool := range a.tools {
		if a.shouldUseTool(ctx, input, tool) {
			result, err := tool.Call(ctx, input)
			if err != nil {
				a.logger.Error(ctx, "tool call failed",
					logger.Fields{
						"tool":  tool.Name(),
						"error": err,
					})
				return "", fmt.Errorf("tool %s call: %w", tool.Name(), err)
			}
			toolResult = result
			usedTool = true
			break
		}
	}

	if usedTool {
		// If a tool was used, incorporate its result
		prompt := fmt.Sprintf("%s\n\nTool result: %s\n\nOriginal question: %s\n\nAnswer the question based on the tool result:", systemPrompt, toolResult, input)
		return a.llm.Call(ctx, prompt)
	}

	// No tool used, answer directly with system prompt
	prompt := fmt.Sprintf("%s\n\nUser: %s\n\nAssistant:", systemPrompt, input)
	return a.llm.Call(ctx, prompt)
}

// buildSystemPrompt creates a system prompt with tool descriptions
func (a *Agent) buildSystemPrompt() string {
	prompt := "You are a helpful assistant with access to the following tools:\n\n"

	for _, tool := range a.tools {
		prompt += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
	}

	prompt += "\nWhen a user asks a question, decide if you should use a tool or answer directly."
	prompt += " If you need to use a tool, respond with the tool name and the input for it."

	return prompt
}

// shouldUseTool determines if a tool should be used for the given input
// This is a simple heuristic-based implementation
func (a *Agent) shouldUseTool(ctx context.Context, input string, tool Tool) bool {
	// Simple keyword matching for tool selection
	// In a full implementation, this would use the LLM to decide

	toolName := tool.Name()
	description := tool.Description()

	// Check if the tool name or description is related to the input
	inputLower := input

	return containsWord(inputLower, toolName) || containsWord(inputLower, description)
}

// containsWord checks if a string contains a word
func containsWord(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || indexOfWord(s, substr) >= 0)
}

// indexOfWord finds the index of a word in a string
func indexOfWord(s, word string) int {
	for i := 0; i <= len(s)-len(word); i++ {
		if s[i:i+len(word)] == word {
			return i
		}
	}
	return -1
}

// SimpleTool is a simple function-based tool implementation
type SimpleTool struct {
	name        string
	description string
	fn          func(ctx context.Context, input string) (string, error)
}

// NewSimpleTool creates a new simple tool
func NewSimpleTool(name, description string, fn func(ctx context.Context, input string) (string, error)) Tool {
	return &SimpleTool{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// Name returns the tool name
func (t *SimpleTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *SimpleTool) Description() string {
	return t.description
}

// Call executes the tool function
func (t *SimpleTool) Call(ctx context.Context, input string) (string, error) {
	return t.fn(ctx, input)
}
