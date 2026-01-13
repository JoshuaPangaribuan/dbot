package langchain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRAGService_Table(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (Service, VectorStore)
		wantNil bool
	}{
		{
			name: "valid service and store",
			setup: func() (Service, VectorStore) {
				svc, err := New(
					WithProvider(ProviderOllama),
				)
				require.NoError(t, err)
				store := NewSimpleVectorStore()
				return svc, store
			},
			wantNil: false,
		},
		{
			name: "nil service",
			setup: func() (Service, VectorStore) {
				return nil, NewSimpleVectorStore()
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := tt.setup()
			ragSvc := NewRAGService(svc, store)

			if tt.wantNil {
				require.Nil(t, ragSvc)
			} else {
				require.NotNil(t, ragSvc)
				_ = svc.Close()
			}
		})
	}
}

func TestRAGService_AddDocuments_Table(t *testing.T) {
	tests := []struct {
		name     string
		texts    []string
		metadata []map[string]any
		wantErr  bool
	}{
		{
			name: "add single document",
			texts: []string{
				"Paris is the capital of France.",
			},
			metadata: nil,
			wantErr:  false,
		},
		{
			name: "add multiple documents",
			texts: []string{
				"Paris is the capital of France.",
				"London is the capital of the UK.",
				"Berlin is the capital of Germany.",
			},
			metadata: nil,
			wantErr:  false,
		},
		{
			name: "add with metadata",
			texts: []string{
				"Document 1",
				"Document 2",
			},
			metadata: []map[string]any{
				{"source": "wiki"},
				{"source": "news"},
			},
			wantErr: false,
		},
		{
			name: "add with partial metadata",
			texts: []string{
				"Document 1",
				"Document 2",
				"Document 3",
			},
			metadata: []map[string]any{
				{"source": "wiki"},
			},
			wantErr: false,
		},
		{
			name:     "empty documents",
			texts:    []string{},
			metadata: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			store := NewSimpleVectorStore()
			ragSvc := NewRAGService(svc, store)
			require.NotNil(t, ragSvc)

			ctx := context.Background()
			err = ragSvc.AddDocuments(ctx, tt.texts, tt.metadata)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRAGService_Query_Table(t *testing.T) {
	tests := []struct {
		name      string
		documents []string
		query     string
		k         int
		wantErr   bool
		skip      bool
		skipReason string
	}{
		{
			name: "query with matching documents",
			documents: []string{
				"Paris is the capital of France.",
				"London is the capital of the UK.",
			},
			query:     "What is the capital of France?",
			k:         2,
			wantErr:   false,
			skip:      true,
			skipReason: "requires LLM API call",
		},
		{
			name: "query with no matches",
			documents: []string{
				"Random text about something.",
			},
			query:     "What is the capital of France?",
			k:         1,
			wantErr:   false,
			skip:      true,
			skipReason: "requires LLM API call",
		},
		{
			name:      "query empty store",
			documents: []string{},
			query:     "What is the capital?",
			k:         1,
			wantErr:   false,
			skip:      true,
			skipReason: "requires LLM API call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			store := NewSimpleVectorStore()
			ragSvc := NewRAGService(svc, store)
			require.NotNil(t, ragSvc)

			ctx := context.Background()
			err = ragSvc.AddDocuments(ctx, tt.documents, nil)
			require.NoError(t, err)

			result, err := ragSvc.Query(ctx, tt.query, tt.k)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
			}
		})
	}
}

func TestSimpleVectorStore_AddDocuments_Table(t *testing.T) {
	tests := []struct {
		name  string
		docs  []Document
		want  int
	}{
		{
			name: "add single document",
			docs: []Document{
				{PageContent: "Test content", Metadata: nil},
			},
			want: 1,
		},
		{
			name: "add multiple documents",
			docs: []Document{
				{PageContent: "Doc 1", Metadata: map[string]any{"id": 1}},
				{PageContent: "Doc 2", Metadata: map[string]any{"id": 2}},
				{PageContent: "Doc 3", Metadata: map[string]any{"id": 3}},
			},
			want: 3,
		},
		{
			name: "add empty documents",
			docs: []Document{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSimpleVectorStore()
			ctx := context.Background()

			err := store.AddDocuments(ctx, tt.docs)
			require.NoError(t, err)

			// Verify documents were added
			result, err := store.SimilaritySearch(ctx, "test", len(tt.docs))
			require.NoError(t, err)
			require.Equal(t, tt.want, len(result))
		})
	}
}

func TestSimpleVectorStore_SimilaritySearch_Table(t *testing.T) {
	tests := []struct {
		name        string
		documents   []string
		query       string
		k           int
		wantMin     int
		wantMax     int
	}{
		{
			name: "search with matches",
			documents: []string{
				"Paris is the capital of France.",
				"London is the capital of the UK.",
				"Berlin is the capital of Germany.",
			},
			query:   "Paris France",
			k:       2,
			wantMin: 1,
			wantMax: 2,
		},
		{
			name: "search with no matches",
			documents: []string{
				"Random text about something.",
				"Another random text.",
			},
			query:   "capital of France",
			k:       1,
			wantMin: 0,
			wantMax: 1,
		},
		{
			name: "search empty store",
			documents: []string{},
			query:     "test",
			k:         1,
			wantMin:   0,
			wantMax:   0,
		},
		{
			name: "k larger than available",
			documents: []string{
				"Document 1",
				"Document 2",
			},
			query:   "Document",
			k:       10,
			wantMin: 2,
			wantMax: 2,
		},
		{
			name: "k is zero",
			documents: []string{
				"Document 1",
				"Document 2",
			},
			query:   "Document",
			k:       0,
			wantMin: 0,
			wantMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSimpleVectorStore()
			ctx := context.Background()

			// Add documents
			for _, doc := range tt.documents {
				docs := []Document{{PageContent: doc}}
				err := store.AddDocuments(ctx, docs)
				require.NoError(t, err)
			}

			// Search
			result, err := store.SimilaritySearch(ctx, tt.query, tt.k)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(result), tt.wantMin)
			require.LessOrEqual(t, len(result), tt.wantMax)
		})
	}
}

func TestDocument_Table(t *testing.T) {
	tests := []struct {
		name  string
		doc   Document
		check func(*testing.T, Document)
	}{
		{
			name: "document with content",
			doc:  Document{PageContent: "Test content", Metadata: nil},
			check: func(t *testing.T, d Document) {
				require.Equal(t, "Test content", d.PageContent)
				require.Nil(t, d.Metadata)
			},
		},
		{
			name: "document with metadata",
			doc: Document{
				PageContent: "Test",
				Metadata: map[string]any{
					"source": "test",
					"id":      1,
				},
			},
			check: func(t *testing.T, d Document) {
				require.Equal(t, "Test", d.PageContent)
				require.Equal(t, "test", d.Metadata["source"])
				require.Equal(t, 1, d.Metadata["id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.doc)
		})
	}
}

func TestRAGService_BuildContext_Table(t *testing.T) {
	tests := []struct {
		name     string
		docs     []Document
		contains []string
	}{
		{
			name: "single document",
			docs: []Document{
				{PageContent: "Test content", Metadata: nil},
			},
			contains: []string{"[Doc 1]", "Test content"},
		},
		{
			name: "multiple documents",
			docs: []Document{
				{PageContent: "Doc 1", Metadata: nil},
				{PageContent: "Doc 2", Metadata: nil},
				{PageContent: "Doc 3", Metadata: nil},
			},
			contains: []string{"[Doc 1]", "[Doc 2]", "[Doc 3]", "Doc 1", "Doc 2", "Doc 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			store := NewSimpleVectorStore()
			ragSvc := NewRAGService(svc, store)
			require.NotNil(t, ragSvc)

			context := ragSvc.buildContext(tt.docs)

			for _, substr := range tt.contains {
				require.Contains(t, context, substr)
			}
		})
	}
}

func TestRAGService_BuildRAGPrompt_Table(t *testing.T) {
	tests := []struct {
		name     string
		context  string
		question string
		contains []string
	}{
		{
			name:     "basic prompt",
			context:  "Paris is the capital.",
			question: "What is the capital?",
			contains: []string{"Context:", "Paris is the capital.", "Question:", "What is the capital?", "Answer:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(WithProvider(ProviderOllama))
			require.NoError(t, err)
			defer svc.Close()

			store := NewSimpleVectorStore()
			ragSvc := NewRAGService(svc, store)
			require.NotNil(t, ragSvc)

			prompt := ragSvc.buildRAGPrompt(tt.context, tt.question)

			for _, substr := range tt.contains {
				require.Contains(t, prompt, substr)
			}
		})
	}
}
