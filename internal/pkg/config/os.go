package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type osConfig struct {
	*baseConfig // Embed base for Close(), watcher, mutex, files
	data        map[string]any
}

// newOSConfig creates a new osConfig with the given settings.
// It loads the initial configuration and starts the watcher.
func newOSConfig(settings *configSettings) (Config, error) {
	c := &osConfig{
		baseConfig: &baseConfig{
			files:   settings.files,
			onError: settings.onError,
		},
		data: make(map[string]any),
	}

	// Initial load
	if err := c.reload(); err != nil {
		return nil, err
	}

	// Setup watcher using base's method
	if err := c.baseConfig.setupWatcher(settings.watch, c.reload); err != nil {
		return nil, err
	}

	return c, nil
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
	c.baseConfig.mu.Lock()
	defer c.baseConfig.mu.Unlock()

	merged := make(map[string]any)

	for _, path := range c.baseConfig.files {
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

// Helper to traverse dot notation
func (c *osConfig) exists(path []string) (any, bool) {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()

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
