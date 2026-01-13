// Package langchain provides integration with LangChainGo for LLM operations.
//
// This package offers a standalone service for Large Language Model (LLM) operations
// with support for multiple providers, chains, RAG (Retrieval Augmented Generation),
// and agent workflows.
//
// # Basic Usage
//
//	svc, err := langchain.New(
//	    langchain.WithProvider(langchain.ProviderOpenAI),
//	    langchain.WithOpenAIConfig(apiKey, ""),
//	    langchain.WithLogger(logger),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer svc.Close()
//
//	response, err := svc.Complete(ctx, "What is the capital of France?")
//
// # Providers
//
// The package supports multiple LLM providers:
//   - ProviderOpenAI: OpenAI API (GPT-3.5, GPT-4, etc.)
//   - ProviderOllama: Local Ollama instance
//
// # Advanced Features
//
//   - Chains: Build complex LLM workflows with ChainBuilder
//   - RAG: Retrieval-augmented generation with vector stores
//   - Agents: LLM agents with tool access
//   - Memory: Conversation history management
//
// # Configuration
//
// The service uses the Options pattern for flexible configuration:
//   - WithProvider: Sets the LLM provider
//   - WithModel: Sets the model name
//   - WithTemperature: Controls response randomness (0.0 - 1.0)
//   - WithMaxTokens: Limits response length
//   - WithLogger: Sets the logger instance
//   - WithTimeout: Sets operation timeout
package langchain
