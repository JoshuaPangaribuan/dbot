// Package ask provides AI-powered question answering.
package ask

import (
	"context"

	"github.com/JoshuaPangaribuan/dbot/internal/features/ask/handlers"
	"github.com/JoshuaPangaribuan/dbot/internal/features/ask/persistence"
	"github.com/JoshuaPangaribuan/dbot/internal/features/ask/service"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/discord"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/langchain"
	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// Config configures the ask feature.
type Config struct {
	Provider     langchain.Provider
	Model        string
	APIKey       string
	BaseURL      string
	Temperature  float32
	MaxTokens    int
	SystemPrompt string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Provider:     langchain.ProviderOpenAI,
		Model:        "gpt-3.5-turbo",
		Temperature:  0.7,
		MaxTokens:    1000,
		SystemPrompt: "You are a helpful AI assistant. Always respond in English.",
	}
}

// Deps holds dependencies for the ask feature.
type Deps struct {
	Bot    *discord.Bot
	Logger logger.Logger
	Config Config
}

// Feature wires together the ask feature components.
type Feature struct {
	svc     service.Service
	persist *persistence.Persist
	handler *handlers.Handler
}

// Register sets up the ask feature with the bot.
func Register(deps Deps) (*Feature, error) {
	// Apply defaults if not set
	if deps.Config.Provider == "" {
		deps.Config.Provider = DefaultConfig().Provider
	}
	if deps.Config.Model == "" {
		deps.Config.Model = DefaultConfig().Model
	}
	if deps.Config.Temperature == 0 {
		deps.Config.Temperature = DefaultConfig().Temperature
	}
	if deps.Config.MaxTokens == 0 {
		deps.Config.MaxTokens = DefaultConfig().MaxTokens
	}
	if deps.Logger == nil {
		deps.Logger = logger.New()
	}

	// Build langchain options based on provider
	opts := []langchain.Option{
		langchain.WithModel(deps.Config.Model),
		langchain.WithTemperature(deps.Config.Temperature),
		langchain.WithMaxTokens(deps.Config.MaxTokens),
		langchain.WithLogger(deps.Logger),
		langchain.WithSystemPrompt(deps.Config.SystemPrompt),
	}

	// Add provider-specific config
	switch deps.Config.Provider {
	case langchain.ProviderOpenAI:
		opts = append(opts, langchain.WithOpenAIConfig(deps.Config.APIKey, deps.Config.BaseURL))
	case langchain.ProviderAnthropic:
		opts = append(opts, langchain.WithAnthropicConfig(deps.Config.APIKey, deps.Config.BaseURL))
	case langchain.ProviderOllama:
		opts = append(opts, langchain.WithOllamaConfig(deps.Config.BaseURL))
	default:
		opts = append(opts, langchain.WithProvider(deps.Config.Provider))
	}

	// Create LangChain service
	langchainSvc, err := langchain.New(opts...)
	if err != nil {
		return nil, err
	}

	// Create completion service
	svcConfig := service.Config{
		MaxResponseLength: 1900, // Discord message limit with safety margin
	}
	svc := service.New(langchainSvc, svcConfig, deps.Logger)

	// Create persistence handler
	persist := persistence.New(svc, deps.Logger)

	// Create handler with service dependency
	handler := handlers.New(svc, deps.Logger)

	f := &Feature{
		svc:     svc,
		persist: persist,
		handler: handler,
	}

	// Register handlers
	deps.Bot.SlashCommand("ask", "Ask the AI a question",
		handler.HandleAsk,
		discord.StringOption("prompt", "Your question for the AI", true),
	)
	deps.Bot.OnShutdown(persist.Save)

	deps.Logger.Info(context.Background(), "ask feature registered", logger.Fields{
		"component": "ask",
		"provider":  deps.Config.Provider,
		"model":     deps.Config.Model,
	})

	return f, nil
}

// Service returns the underlying service for external access if needed.
func (f *Feature) Service() service.Service {
	return f.svc
}
