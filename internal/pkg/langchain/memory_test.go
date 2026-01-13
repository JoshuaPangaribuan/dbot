package langchain

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMemory_Table(t *testing.T) {
	tests := []struct {
		name         string
		maxSize      int
		wantCapacity int
	}{
		{
			name:         "default size",
			maxSize:      0,
			wantCapacity: 10,
		},
		{
			name:         "custom size",
			maxSize:      5,
			wantCapacity: 5,
		},
		{
			name:         "negative size uses default",
			maxSize:      -1,
			wantCapacity: 10,
		},
		{
			name:         "large size",
			maxSize:      100,
			wantCapacity: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(tt.maxSize)
			require.NotNil(t, mem)
			require.Equal(t, 0, mem.Size())
		})
	}
}

func TestMemory_Add_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		maxSize  int
		wantSize int
	}{
		{
			name: "add single message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			maxSize:  10,
			wantSize: 1,
		},
		{
			name: "add multiple messages",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
				{Role: "user", Content: "How are you?"},
			},
			maxSize:  10,
			wantSize: 3,
		},
		{
			name: "trim to max size",
			messages: []Message{
				{Role: "user", Content: "1"},
				{Role: "user", Content: "2"},
				{Role: "user", Content: "3"},
				{Role: "user", Content: "4"},
				{Role: "user", Content: "5"},
			},
			maxSize:  3,
			wantSize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(tt.maxSize)
			for _, msg := range tt.messages {
				mem.Add(msg.Role, msg.Content)
			}
			require.Equal(t, tt.wantSize, mem.Size())
		})
	}
}

func TestMemory_Get_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		wantLen  int
	}{
		{
			name:     "empty memory",
			messages: nil,
			wantLen:  0,
		},
		{
			name: "single message",
			messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			wantLen: 1,
		},
		{
			name: "multiple messages",
			messages: []Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			for _, msg := range tt.messages {
				mem.Add(msg.Role, msg.Content)
			}

			got := mem.Get()
			require.Equal(t, tt.wantLen, len(got))

			// Verify returned messages match added messages
			if tt.wantLen > 0 {
				for i := range tt.messages {
					if i < len(got) {
						require.Equal(t, tt.messages[i].Role, got[i].Role)
						require.Equal(t, tt.messages[i].Content, got[i].Content)
					}
				}
			}
		})
	}
}

func TestMemory_Clear_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
	}{
		{
			name:     "clear empty memory",
			messages: nil,
		},
		{
			name: "clear with messages",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			for _, msg := range tt.messages {
				mem.Add(msg.Role, msg.Content)
			}
			require.GreaterOrEqual(t, mem.Size(), 0)

			mem.Clear()
			require.Equal(t, 0, mem.Size())

			got := mem.Get()
			require.Empty(t, got)
		})
	}
}

func TestMemory_Last_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		n        int
		wantLen  int
	}{
		{
			name:     "last 0 from empty",
			messages: nil,
			n:        0,
			wantLen:  0,
		},
		{
			name: "last 2 from 5",
			messages: []Message{
				{Role: "user", Content: "1"},
				{Role: "user", Content: "2"},
				{Role: "user", Content: "3"},
				{Role: "user", Content: "4"},
				{Role: "user", Content: "5"},
			},
			n:       2,
			wantLen: 2,
		},
		{
			name: "last more than available",
			messages: []Message{
				{Role: "user", Content: "1"},
				{Role: "user", Content: "2"},
			},
			n:       5,
			wantLen: 2,
		},
		{
			name: "last all",
			messages: []Message{
				{Role: "user", Content: "1"},
				{Role: "user", Content: "2"},
				{Role: "user", Content: "3"},
			},
			n:       3,
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			for _, msg := range tt.messages {
				mem.Add(msg.Role, msg.Content)
			}

			got := mem.Last(tt.n)
			require.Equal(t, tt.wantLen, len(got))

			// Verify the content is correct
			if tt.wantLen > 0 && len(tt.messages) > 0 {
				startIdx := len(tt.messages) - tt.wantLen
				for i, msg := range got {
					require.Equal(t, tt.messages[startIdx+i].Role, msg.Role)
					require.Equal(t, tt.messages[startIdx+i].Content, msg.Content)
				}
			}
		})
	}
}

func TestMemory_ConvenienceMethods_Table(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*Memory)
		wantSize int
	}{
		{
			name: "AddSystem",
			method: func(m *Memory) {
				m.AddSystem("You are a helpful assistant")
			},
			wantSize: 1,
		},
		{
			name: "AddUser",
			method: func(m *Memory) {
				m.AddUser("Hello")
			},
			wantSize: 1,
		},
		{
			name: "AddAssistant",
			method: func(m *Memory) {
				m.AddAssistant("Hi there!")
			},
			wantSize: 1,
		},
		{
			name: "multiple convenience methods",
			method: func(m *Memory) {
				m.AddSystem("System message")
				m.AddUser("User message")
				m.AddAssistant("Assistant message")
			},
			wantSize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			tt.method(mem)
			require.Equal(t, tt.wantSize, mem.Size())

			messages := mem.Get()
			require.Equal(t, tt.wantSize, len(messages))
		})
	}
}

func TestMemory_ToMessages_Table(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
	}{
		{
			name:     "empty memory",
			messages: nil,
		},
		{
			name: "with messages",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			for _, msg := range tt.messages {
				mem.Add(msg.Role, msg.Content)
			}

			got := mem.ToMessages()
			require.Equal(t, len(tt.messages), len(got))

			for i := range tt.messages {
				require.Equal(t, tt.messages[i].Role, got[i].Role)
				require.Equal(t, tt.messages[i].Content, got[i].Content)
			}
		})
	}
}

func TestMemory_ConcurrentAccess_Table(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		operations int
	}{
		{
			name:       "single goroutine",
			goroutines: 1,
			operations: 100,
		},
		{
			name:       "multiple goroutines",
			goroutines: 10,
			operations: 100,
		},
		{
			name:       "many goroutines",
			goroutines: 50,
			operations: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(1000)
			var wg sync.WaitGroup

			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.operations; j++ {
						mem.Add("user", "message")
						_ = mem.Get()
						_ = mem.Size()
					}
				}(i)
			}

			wg.Wait()
			// Should not panic and size should be within bounds
			require.LessOrEqual(t, mem.Size(), 1000)
		})
	}
}

func TestMemory_ReturnsCopy_Table(t *testing.T) {
	tests := []struct {
		name string
		act  func(*Memory)
	}{
		{
			name: "modify Get result",
			act: func(m *Memory) {
				m.Add("user", "original")
				got := m.Get()
				got[0] = Message{Role: "user", Content: "modified"}

				// Original should be unchanged
				original := m.Get()
				require.Equal(t, "original", original[0].Content)
			},
		},
		{
			name: "modify Last result",
			act: func(m *Memory) {
				m.Add("user", "original")
				got := m.Last(1)
				got[0] = Message{Role: "user", Content: "modified"}

				// Original should be unchanged
				original := m.Last(1)
				require.Equal(t, "original", original[0].Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMemory(10)
			tt.act(mem)
		})
	}
}
