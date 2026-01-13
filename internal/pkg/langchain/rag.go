package langchain

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaPangaribuan/dbot/internal/pkg/logger"
)

// RAGService provides retrieval-augmented generation
type RAGService struct {
	llm         LLM
	vectorStore VectorStore
	logger      logger.Logger
}

// VectorStore interface for document storage/retrieval
type VectorStore interface {
	AddDocuments(ctx context.Context, docs []Document) error
	SimilaritySearch(ctx context.Context, query string, k int) ([]Document, error)
}

// Document represents a text document with metadata
type Document struct {
	PageContent string
	Metadata    map[string]any
}

// NewRAGService creates a RAG service
func NewRAGService(svc Service, store VectorStore) *RAGService {
	s, ok := svc.(*service)
	if !ok {
		return nil
	}

	return &RAGService{
		llm:         s.llm,
		vectorStore: store,
		logger:      s.logger,
	}
}

// Query performs RAG query with retrieval and generation
func (r *RAGService) Query(ctx context.Context, question string, k int) (string, error) {
	// 1. Retrieve relevant documents
	docs, err := r.vectorStore.SimilaritySearch(ctx, question, k)
	if err != nil {
		return "", fmt.Errorf("retrieve documents: %w", err)
	}

	r.logger.Debug(ctx, "rag retrieved documents",
		logger.Fields{
			"query":       question,
			"doc_count":   len(docs),
			"retrieval_k": k,
		})

	// 2. Build prompt with context
	contextText := r.buildContext(docs)
	prompt := r.buildRAGPrompt(contextText, question)

	// 3. Generate answer
	result, err := r.llm.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	return result, nil
}

// AddDocuments adds documents to the vector store
func (r *RAGService) AddDocuments(ctx context.Context, texts []string, metadata []map[string]any) error {
	docs := make([]Document, len(texts))
	for i, text := range texts {
		meta := map[string]any{}
		if i < len(metadata) {
			meta = metadata[i]
		}
		docs[i] = Document{
			PageContent: text,
			Metadata:    meta,
		}
	}

	err := r.vectorStore.AddDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("add documents: %w", err)
	}

	r.logger.Debug(ctx, "rag added documents",
		logger.Fields{
			"doc_count": len(docs),
		})

	return nil
}

// buildContext creates a context string from retrieved documents
func (r *RAGService) buildContext(docs []Document) string {
	var parts []string
	for i, doc := range docs {
		parts = append(parts, fmt.Sprintf("[Doc %d]: %s", i+1, doc.PageContent))
	}
	return strings.Join(parts, "\n\n")
}

// buildRAGPrompt creates a RAG prompt with context
func (r *RAGService) buildRAGPrompt(context, question string) string {
	return fmt.Sprintf(`Use the following context to answer the question. If you cannot answer from the context, say so.

Context:
%s

Question: %s

Answer:`, context, question)
}

// SimpleVectorStore is an in-memory vector store for testing
type SimpleVectorStore struct {
	documents []Document
}

// NewSimpleVectorStore creates a new in-memory vector store
func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		documents: make([]Document, 0),
	}
}

// AddDocuments adds documents to the store
func (s *SimpleVectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	s.documents = append(s.documents, docs...)
	return nil
}

// SimilaritySearch performs a simple similarity search
// This is a placeholder implementation that uses basic string matching
// A real implementation would use vector embeddings
func (s *SimpleVectorStore) SimilaritySearch(ctx context.Context, query string, k int) ([]Document, error) {
	if k <= 0 || k > len(s.documents) {
		k = len(s.documents)
	}

	// Simple implementation: return documents that contain the query
	var results []Document
	count := 0
	for _, doc := range s.documents {
		if count >= k {
			break
		}
		if strings.Contains(strings.ToLower(doc.PageContent), strings.ToLower(query)) {
			results = append(results, doc)
			count++
		}
	}

	// If no matches, return the first k documents
	if len(results) == 0 && len(s.documents) > 0 {
		end := k
		if end > len(s.documents) {
			end = len(s.documents)
		}
		results = s.documents[:end]
	}

	return results, nil
}
