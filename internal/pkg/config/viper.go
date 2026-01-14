package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type viperConfig struct {
	*baseConfig // Embed base
	v           *viper.Viper
}

func newViperConfig(settings *configSettings) (Config, error) {
	v := viper.New()

	if err := loadViperFiles(v, settings.files); err != nil {
		return nil, err
	}

	c := &viperConfig{
		baseConfig: &baseConfig{
			files:   settings.files,
			onError: settings.onError,
		},
		v: v,
	}

	if err := c.baseConfig.setupWatcher(settings.watch, c.reload); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *viperConfig) GetString(key string) string {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
	return c.v.GetString(key)
}

func (c *viperConfig) GetInt(key string) int {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
	return c.v.GetInt(key)
}

func (c *viperConfig) GetInt64(key string) int64 {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
	return c.v.GetInt64(key)
}

func (c *viperConfig) GetFloat64(key string) float64 {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
	return c.v.GetFloat64(key)
}

func (c *viperConfig) GetBool(key string) bool {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
	return c.v.GetBool(key)
}

func (c *viperConfig) GetSlice(key string) []any {
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
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
	c.baseConfig.mu.RLock()
	defer c.baseConfig.mu.RUnlock()
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

func (c *viperConfig) reload() error {
	v := viper.New()
	if err := loadViperFiles(v, c.baseConfig.files); err != nil {
		return err
	}
	c.baseConfig.mu.Lock()
	c.v = v
	c.baseConfig.mu.Unlock()
	return nil
}
