package v1

import "sync"

// StateStore provides thread-safe key-value storage for features.
// Each feature can use this to store state that persists across event handlers.
type StateStore struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewStateStore creates a new StateStore.
func NewStateStore() *StateStore {
	return &StateStore{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value by key. Returns the value and true if found,
// or nil and false if not found.
func (s *StateStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set stores a value with the given key.
func (s *StateStore) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a key from the store.
func (s *StateStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// GetOrSet retrieves a value by key, or sets and returns the result
// of calling factory if the key doesn't exist. This is atomic.
func (s *StateStore) GetOrSet(key string, factory func() interface{}) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.data[key]; ok {
		return v
	}
	v := factory()
	s.data[key] = v
	return v
}

// Keys returns all keys in the store.
func (s *StateStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Clear removes all keys from the store.
func (s *StateStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
}

// Len returns the number of items in the store.
func (s *StateStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// ScopedState provides a namespaced view of StateStore.
// Keys are automatically prefixed with the scope name.
type ScopedState struct {
	store  *StateStore
	prefix string
}

// NewScopedState creates a scoped view of a StateStore.
// All keys will be prefixed with "scope:".
func NewScopedState(store *StateStore, scope string) *ScopedState {
	return &ScopedState{
		store:  store,
		prefix: scope + ":",
	}
}

// Get retrieves a value by key (prefixed with scope).
func (s *ScopedState) Get(key string) (interface{}, bool) {
	return s.store.Get(s.prefix + key)
}

// Set stores a value with the given key (prefixed with scope).
func (s *ScopedState) Set(key string, value interface{}) {
	s.store.Set(s.prefix+key, value)
}

// Delete removes a key from the store (prefixed with scope).
func (s *ScopedState) Delete(key string) {
	s.store.Delete(s.prefix + key)
}

// GetOrSet retrieves or creates a value atomically (prefixed with scope).
func (s *ScopedState) GetOrSet(key string, factory func() interface{}) interface{} {
	return s.store.GetOrSet(s.prefix+key, factory)
}

