package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.yaml.in/yaml/v3"
)

type osConfig struct {
	mu      sync.RWMutex
	data    map[string]any
	files   []string
	watch   bool
	watcher *fsnotify.Watcher
	onError func(error)
}

// newOSConfig creates a new osConfig with the given settings.
// It loads the initial configuration and starts the watcher.
func newOSConfig(settings *configSettings) (Config, error) {
	c := &osConfig{
		data:    make(map[string]any),
		watch:   settings.watch,
		files:   settings.files,
		onError: settings.onError,
	}

	// Initial load
	if err := c.reload(); err != nil {
		return nil, err
	}

	// Setup watcher
	if c.watch && len(c.files) > 0 {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("failed to create watcher: %w", err)
		}
		c.watcher = w

		// Watch all unique directories containing config files
		watchedDirs := make(map[string]bool)
		for _, f := range c.files {
			absPath, err := filepath.Abs(f)
			if err == nil {
				dir := filepath.Dir(absPath)
				if !watchedDirs[dir] {
					if err := w.Add(dir); err != nil {
						if c.onError != nil {
							c.onError(fmt.Errorf("failed to watch directory %s: %w", toRelativePath(dir), err))
						}
					} else {
						watchedDirs[dir] = true
					}
				}
			}
		}

		go c.watchLoop()
	}

	return c, nil
}

func (c *osConfig) Close() error {
	if c.watcher != nil {
		return c.watcher.Close()
	}
	return nil
}

func unmarshalConfigFile(path string, content []byte) (map[string]any, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var current map[string]any
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &current); err != nil {
			return nil, err
		}
	default:
		if err := json.Unmarshal(content, &current); err != nil {
			return nil, err
		}
	}
	return current, nil
}

func (c *osConfig) reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	merged := make(map[string]any)

	for _, path := range c.files {
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip missing files
			}
			return fmt.Errorf("failed to read config file %s: %w", toRelativePath(path), err)
		}

		current, err := unmarshalConfigFile(path, content)
		if err != nil {
			return fmt.Errorf("failed to parse config file %s: %w", toRelativePath(path), err)
		}

		mergeMaps(merged, current)
	}

	c.data = merged
	return nil
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstVal, ok := dst[k]; ok {
				if dstMap, ok := dstVal.(map[string]any); ok {
					mergeMaps(dstMap, srcMap)
					continue
				}
			}
		}
		dst[k] = v
	}
}

func (c *osConfig) watchLoop() {
	debounceDuration := 100 * time.Millisecond
	var timer *time.Timer

	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				// check if event file matches any of our config files
				matches := false
				absEventPath, _ := filepath.Abs(event.Name)
				for _, f := range c.files {
					absConfigPath, _ := filepath.Abs(f)
					if absEventPath == absConfigPath {
						matches = true
						break
					}
				}

				if matches {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(debounceDuration, func() {
						if err := c.reload(); err != nil {
							if c.onError != nil {
								c.onError(err)
							}
						}
					})
				}
			}
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			if c.onError != nil {
				c.onError(err)
			}
		}
	}
}

// Helper to traverse dot notation
func (c *osConfig) exists(path []string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var current any = c.data
	for _, segment := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, ok := m[segment]
		if !ok {
			return nil, false
		}
		current = val
	}
	return current, true
}

func (c *osConfig) Get(key string) string {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return ""
	}
	// Best effort string conversion
	return fmt.Sprintf("%v", val)
}

func (c *osConfig) GetString(key string) string {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

func (c *osConfig) GetInt(key string) int {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

func (c *osConfig) GetInt64(key string) int64 {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

func (c *osConfig) GetFloat64(key string) float64 {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

func (c *osConfig) GetBool(key string) bool {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

func (c *osConfig) GetSlice(key string) []any {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return nil
	}
	if s, ok := val.([]any); ok {
		return copySlice(s)
	}
	return nil
}

func (c *osConfig) GetMap(key string) map[string]any {
	val, ok := c.exists(strings.Split(key, "."))
	if !ok {
		return nil
	}
	if m, ok := val.(map[string]any); ok {
		return copyMap(m)
	}
	return nil
}
