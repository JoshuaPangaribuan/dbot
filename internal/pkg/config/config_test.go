package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
	baseDir string
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) SetupTest() {
	root := filepath.Join(".", ".cache", "test")
	require.NoError(s.T(), os.MkdirAll(root, 0o755))

	dir, err := os.MkdirTemp(root, "config-")
	require.NoError(s.T(), err)
	s.baseDir = dir
}

func (s *ConfigSuite) TearDownTest() {
	require.NoError(s.T(), os.RemoveAll(s.baseDir))
}

func (s *ConfigSuite) writeFile(name, contents string) string {
	s.T().Helper()
	path := filepath.Join(s.baseDir, name)
	require.NoError(s.T(), os.WriteFile(path, []byte(contents), 0o644))
	return path
}

func (s *ConfigSuite) TestNew_SelectsProvider_Table() {
	tests := []struct {
		name     string
		provider Provider
		assert   func(t *testing.T, cfg Config)
	}{
		{
			name:     "os provider",
			provider: ProviderOS,
			assert: func(t *testing.T, cfg Config) {
				_, ok := cfg.(*osConfig)
				require.True(t, ok)
			},
		},
		{
			name:     "viper provider",
			provider: ProviderViper,
			assert: func(t *testing.T, cfg Config) {
				_, ok := cfg.(*viperConfig)
				require.True(t, ok)
			},
		},
		{
			name:     "unknown provider defaults to os",
			provider: Provider("unknown"),
			assert: func(t *testing.T, cfg Config) {
				_, ok := cfg.(*osConfig)
				require.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg, err := New(WithProvider(tt.provider), WithWatcher(false))
			require.NoError(s.T(), err)
			tt.assert(s.T(), cfg)
			require.NoError(s.T(), cfg.Close())
		})
	}
}

func (s *ConfigSuite) TestOSProvider_LoadsAndMergesFiles_Table() {
	base := s.writeFile("base.yaml", `
app:
  name: "dbot"
  num: 123
  enabled: true
  ratio: 1.5
  tags: ["a", "b"]
  nested:
    key: "val"
`)
	override := s.writeFile("override.yaml", `
app:
  name: "override"
  nested:
    other: 2
`)

	cfg, err := New(
		WithProvider(ProviderOS),
		WithFiles(base, override),
		WithWatcher(false),
	)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { require.NoError(s.T(), cfg.Close()) })

	tests := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "string getter",
			assert: func(t *testing.T) {
				require.Equal(t, "override", cfg.GetString("app.name"))
			},
		},
		{
			name: "int getter",
			assert: func(t *testing.T) {
				require.Equal(t, 123, cfg.GetInt("app.num"))
			},
		},
		{
			name: "bool getter",
			assert: func(t *testing.T) {
				require.True(t, cfg.GetBool("app.enabled"))
			},
		},
		{
			name: "float getter",
			assert: func(t *testing.T) {
				require.InDelta(t, 1.5, cfg.GetFloat64("app.ratio"), 0.0001)
			},
		},
		{
			name: "dot traversal preserves nested",
			assert: func(t *testing.T) {
				require.Equal(t, "val", cfg.GetString("app.nested.key"))
				require.Equal(t, 2, cfg.GetInt("app.nested.other"))
			},
		},
		{
			name: "slice getter returns copy",
			assert: func(t *testing.T) {
				got := cfg.GetSlice("app.tags")
				require.Equal(t, []any{"a", "b"}, got)
				got[0] = "changed"
				require.Equal(t, []any{"a", "b"}, cfg.GetSlice("app.tags"))
			},
		},
		{
			name: "map getter returns deep copy",
			assert: func(t *testing.T) {
				got := cfg.GetMap("app.nested")
				require.Equal(t, map[string]any{"key": "val", "other": 2}, got)
				got["key"] = "changed"
				require.Equal(t, "val", cfg.GetString("app.nested.key"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.assert(s.T())
		})
	}
}

func (s *ConfigSuite) TestOSProvider_ParsesJSON() {
	path := s.writeFile("cfg.json", `{"a":{"b":1}}`)

	cfg, err := New(WithProvider(ProviderOS), WithFiles(path), WithWatcher(false))
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { require.NoError(s.T(), cfg.Close()) })

	require.Equal(s.T(), 1, cfg.GetInt("a.b"))
}

func (s *ConfigSuite) TestViperProvider_LoadsValues() {
	path := s.writeFile("cfg.yaml", `
app:
  name: "dbot"
  num: 123
  enabled: true
  tags: ["a", "b"]
`)

	cfg, err := New(WithProvider(ProviderViper), WithFiles(path), WithWatcher(false))
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { require.NoError(s.T(), cfg.Close()) })

	require.Equal(s.T(), "dbot", cfg.GetString("app.name"))
	require.Equal(s.T(), 123, cfg.GetInt("app.num"))
	require.True(s.T(), cfg.GetBool("app.enabled"))
	require.Equal(s.T(), []any{"a", "b"}, cfg.GetSlice("app.tags"))
}

func (s *ConfigSuite) TestClose_WithWatcherEnabled_DoesNotError() {
	path := s.writeFile("cfg.yaml", "a: 1\n")

	cfg, err := New(WithProvider(ProviderOS), WithFiles(path), WithWatcher(true))
	require.NoError(s.T(), err)
	require.NoError(s.T(), cfg.Close())
}
