package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LoggerSuite struct {
	suite.Suite
	buf bytes.Buffer
}

type testCtxKey string

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (s *LoggerSuite) SetupTest() {
	s.buf.Reset()
}

func decodeLogLine(t *testing.T, b []byte) map[string]any {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var m map[string]any
	require.NoError(t, dec.Decode(&m))
	return m
}

func (s *LoggerSuite) TestNew_SelectsProvider_Table() {
	tests := []struct {
		name     string
		provider Provider
		assert   func(t *testing.T, l Logger)
	}{
		{
			name:     "slog provider",
			provider: ProviderSlog,
			assert: func(t *testing.T, l Logger) {
				_, ok := l.(*slogLogger)
				require.True(t, ok)
			},
		},
		{
			name:     "zap provider",
			provider: ProviderZap,
			assert: func(t *testing.T, l Logger) {
				_, ok := l.(*zapLogger)
				require.True(t, ok)
			},
		},
		{
			name:     "unknown provider defaults to slog",
			provider: Provider("unknown"),
			assert: func(t *testing.T, l Logger) {
				_, ok := l.(*slogLogger)
				require.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			l := New(WithProvider(tt.provider))
			tt.assert(s.T(), l)
		})
	}
}

func (s *LoggerSuite) TestMergeFields_Table() {
	tests := []struct {
		name     string
		base     Fields
		override Fields
		want     Fields
	}{
		{
			name:     "nil base uses override",
			base:     nil,
			override: Fields{"a": 1},
			want:     Fields{"a": 1},
		},
		{
			name:     "nil override uses base",
			base:     Fields{"a": 1},
			override: nil,
			want:     Fields{"a": 1},
		},
		{
			name:     "merge disjoint",
			base:     Fields{"a": 1},
			override: Fields{"b": 2},
			want:     Fields{"a": 1, "b": 2},
		},
		{
			name:     "override wins on conflict",
			base:     Fields{"a": 1},
			override: Fields{"a": 2},
			want:     Fields{"a": 2},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := mergeFields(tt.base, tt.override)
			require.Equal(s.T(), tt.want, got)
		})
	}
}

func (s *LoggerSuite) TestTrimmedCaller_Table() {
	tests := []struct {
		name string
		file string
		line int
		want string
	}{
		{
			name: "empty file",
			file: "",
			line: 1,
			want: "",
		},
		{
			name: "simple path",
			file: "a/b/c.go",
			line: 42,
			want: "b/c.go:42",
		},
		{
			name: "no slash path",
			file: "main.go",
			line: 7,
			want: "main.go:7",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			require.Equal(s.T(), tt.want, trimmedCaller(tt.file, tt.line))
		})
	}
}

func (s *LoggerSuite) TestLog_WritesJSON_WithContextFields_Table() {
	tests := []struct {
		name      string
		provider  Provider
		wantLevel string
	}{
		{name: "slog", provider: ProviderSlog, wantLevel: "INFO"},
		{name: "zap", provider: ProviderZap, wantLevel: "info"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.buf.Reset()

			ctx := context.WithValue(context.Background(), testCtxKey("ctxKey"), "ctxVal")

			l := New(
				WithProvider(tt.provider),
				WithOutput(&s.buf),
				WithLevel(LogLevelDebug),
				WithContextFields(func(ctx context.Context) Fields {
					v, _ := ctx.Value(testCtxKey("ctxKey")).(string)
					return Fields{"ctxKey": v}
				}),
			)

			l.Info(ctx, "hello", Fields{"a": "b"})

			out := s.buf.String()
			require.NotEmpty(s.T(), out)

			m := decodeLogLine(s.T(), []byte(out))
			require.Equal(s.T(), "hello", m["msg"])
			require.Equal(s.T(), "b", m["a"])
			require.Equal(s.T(), "ctxVal", m["ctxKey"])
			require.Equal(s.T(), tt.wantLevel, m["level"])

			caller, ok := m["caller"].(string)
			require.True(s.T(), ok)
			require.True(s.T(), strings.Contains(caller, ".go:"))
		})
	}
}

func (s *LoggerSuite) TestLog_LevelFilters_Table() {
	tests := []struct {
		name     string
		provider Provider
		level    LogLevel
		logFn    func(l Logger)
		wantAny  bool
	}{
		{
			name:     "info suppressed at error level",
			provider: ProviderSlog,
			level:    LogLevelError,
			logFn: func(l Logger) {
				l.Info(context.Background(), "info", nil)
			},
			wantAny: false,
		},
		{
			name:     "error allowed at error level",
			provider: ProviderSlog,
			level:    LogLevelError,
			logFn: func(l Logger) {
				l.Error(context.Background(), "err", nil)
			},
			wantAny: true,
		},
		{
			name:     "info suppressed at error level (zap)",
			provider: ProviderZap,
			level:    LogLevelError,
			logFn: func(l Logger) {
				l.Info(context.Background(), "info", nil)
			},
			wantAny: false,
		},
		{
			name:     "error allowed at error level (zap)",
			provider: ProviderZap,
			level:    LogLevelError,
			logFn: func(l Logger) {
				l.Error(context.Background(), "err", nil)
			},
			wantAny: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.buf.Reset()

			l := New(WithProvider(tt.provider), WithOutput(&s.buf), WithLevel(tt.level))
			tt.logFn(l)

			gotAny := strings.TrimSpace(s.buf.String()) != ""
			require.Equal(s.T(), tt.wantAny, gotAny)
		})
	}
}
