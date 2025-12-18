package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type viperConfig struct {
	mu      sync.RWMutex
	v       *viper.Viper
	files   []string
	watcher *fsnotify.Watcher
	onError func(error)
}

func newViperConfig(settings *configSettings) (Config, error) {
	v := viper.New()

	if err := loadViperFiles(v, settings.files); err != nil {
		return nil, err
	}

	c := &viperConfig{
		v:       v,
		files:   settings.files,
		onError: settings.onError,
	}

	if settings.watch && len(settings.files) > 0 {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("failed to create watcher: %w", err)
		}
		c.watcher = w

		watchedDirs := make(map[string]bool)
		for _, f := range settings.files {
			absPath, err := filepath.Abs(f)
			if err != nil {
				continue
			}
			dir := filepath.Dir(absPath)
			if watchedDirs[dir] {
				continue
			}
			if err := w.Add(dir); err != nil {
				if c.onError != nil {
					c.onError(fmt.Errorf("failed to watch directory %s: %w", toRelativePath(dir), err))
				}
				continue
			}
			watchedDirs[dir] = true
		}

		go c.watchLoop()
	}

	return c, nil
}

func (c *viperConfig) Close() error {
	if c.watcher != nil {
		return c.watcher.Close()
	}
	return nil
}

func (c *viperConfig) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetString(key)
}

func (c *viperConfig) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetString(key)
}

func (c *viperConfig) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetInt(key)
}

func (c *viperConfig) GetInt64(key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetInt64(key)
}

func (c *viperConfig) GetFloat64(key string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetFloat64(key)
}

func (c *viperConfig) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetBool(key)
}

func (c *viperConfig) GetSlice(key string) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Viper Get returns interface{}, GetStringSlice returns []string.
	// We need []any.
	val := c.v.Get(key)
	if s, ok := val.([]any); ok {
		return copySlice(s)
	}
	if s, ok := val.([]string); ok {
		res := make([]any, len(s))
		for i, v := range s {
			res[i] = v
		}
		return res
	}
	// Viper might return []interface{} which is []any
	return nil
}

func (c *viperConfig) GetMap(key string) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return copyMap(c.v.GetStringMap(key))
}

func loadViperFiles(v *viper.Viper, files []string) error {
	read := false
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to stat config file %s: %w", toRelativePath(f), err)
		}

		v.SetConfigFile(f)
		var err error
		if !read {
			err = v.ReadInConfig()
			read = true
		} else {
			err = v.MergeInConfig()
		}
		if err != nil {
			return fmt.Errorf("failed to load config file %s: %w", toRelativePath(f), err)
		}
	}
	return nil
}

func (c *viperConfig) watchLoop() {
	debounceDuration := 100 * time.Millisecond
	var timer *time.Timer

	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
				continue
			}

			matches := false
			absEventPath, _ := filepath.Abs(event.Name)
			for _, f := range c.files {
				absConfigPath, _ := filepath.Abs(f)
				if absEventPath == absConfigPath {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounceDuration, func() {
				if err := c.reload(); err != nil && c.onError != nil {
					c.onError(err)
				}
			})
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

func (c *viperConfig) reload() error {
	v := viper.New()
	if err := loadViperFiles(v, c.files); err != nil {
		return err
	}
	c.mu.Lock()
	c.v = v
	c.mu.Unlock()
	return nil
}
