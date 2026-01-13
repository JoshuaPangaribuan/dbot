package langchain

import (
	"sync"
)

// Memory provides conversation history management
type Memory struct {
	messages []Message
	maxSize  int
	mu       sync.RWMutex
}

// NewMemory creates conversation memory with a maximum size
func NewMemory(maxSize int) *Memory {
	if maxSize <= 0 {
		maxSize = 10 // Default max size
	}
	return &Memory{
		messages: make([]Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add adds a message to memory
func (m *Memory) Add(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, Message{
		Role:    role,
		Content: content,
	})

	// Trim to max size
	if len(m.messages) > m.maxSize {
		m.messages = m.messages[len(m.messages)-m.maxSize:]
	}
}

// Get returns all messages in memory
func (m *Memory) Get() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// Clear clears all messages from memory
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]Message, 0, m.maxSize)
}

// Size returns the current number of messages in memory
func (m *Memory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.messages)
}

// Last returns the last n messages from memory
func (m *Memory) Last(n int) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n <= 0 || n > len(m.messages) {
		n = len(m.messages)
	}

	start := len(m.messages) - n
	result := make([]Message, n)
	copy(result, m.messages[start:])
	return result
}

// AddSystem adds a system message to memory
func (m *Memory) AddSystem(content string) {
	m.Add("system", content)
}

// AddUser adds a user message to memory
func (m *Memory) AddUser(content string) {
	m.Add("user", content)
}

// AddAssistant adds an assistant message to memory
func (m *Memory) AddAssistant(content string) {
	m.Add("assistant", content)
}

// ToMessages returns the messages as a slice suitable for Chat
func (m *Memory) ToMessages() []Message {
	return m.Get()
}
