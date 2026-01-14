package config

import (
	"sync"
)

// baseConfig provides common functionality for both providers
type baseConfig struct {
	mu      sync.RWMutex
	files   []string
	watcher *fileWatcher
	onError func(error)
}

// Close closes the file watcher if active
func (b *baseConfig) Close() error {
	if b.watcher != nil {
		return b.watcher.Close()
	}
	return nil
}

// setupWatcher initializes the file watcher if enabled
func (b *baseConfig) setupWatcher(enabled bool, reloadFunc func() error) error {
	if !enabled || len(b.files) == 0 {
		return nil
	}
	fw, err := newFileWatcher(b.files, reloadFunc, b.onError)
	if err != nil {
		return err
	}
	b.watcher = fw
	return nil
}
