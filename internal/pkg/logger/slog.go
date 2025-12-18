package logger

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// slogLogger is the slog implementation of Logger
type slogLogger struct {
	handler slog.Handler
	ctxFn   ContextFieldsFunc
}

func newSlogLogger(cfg *config) *slogLogger {
	handler := slog.NewJSONHandler(cfg.output, &slog.HandlerOptions{
		Level: func(cfg *config) slog.Level {
			switch cfg.level {
			case LogLevelDebug:
				return slog.LevelDebug
			case LogLevelInfo:
				return slog.LevelInfo
			case LogLevelWarn:
				return slog.LevelWarn
			case LogLevelError:
				return slog.LevelError
			default:
				return slog.LevelInfo
			}
		}(cfg),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove default keys to control order manually
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
				return slog.Attr{}
			}
			// Rename custom keys back to standard names
			switch a.Key {
			case "time_custom":
				a.Key = "time"
			case "caller_custom":
				a.Key = "caller"
			case "level_custom":
				a.Key = "level"
			case "msg_custom":
				a.Key = "msg"
			}
			return a
		},
	})
	return &slogLogger{handler: handler, ctxFn: cfg.ctxFn}
}

func (l *slogLogger) log(ctx context.Context, level slog.Level, msg string, fields Fields) {
	if !l.handler.Enabled(ctx, level) {
		return
	}

	frame := l.findCaller()

	// Create record with empty standard fields to be removed by ReplaceAttr
	r := slog.NewRecord(time.Time{}, level, "", 0)

	// Add fields in desired order: time, level, caller, msg, <fields...>
	r.Add("time_custom", time.Now().Format(time.RFC3339))
	r.Add("level_custom", level.String())

	if frame.File != "" {
		r.Add("caller_custom", trimmedCaller(frame.File, frame.Line))
	}

	r.Add("msg_custom", msg)

	if l.ctxFn != nil {
		fields = mergeFields(l.ctxFn(ctx), fields)
	}

	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			r.Add(k, fields[k])
		}
	}

	_ = l.handler.Handle(ctx, r)
}

func (l *slogLogger) findCaller() runtime.Frame {
	var pcs [16]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	for {
		f, more := frames.Next()
		if f.File != "" {
			s := filepath.ToSlash(f.File)
			if !strings.Contains(s, "/internal/pkg/logger/") {
				return f
			}
		}
		if !more {
			return runtime.Frame{}
		}
	}
}

func (l *slogLogger) Info(ctx context.Context, msg string, fields Fields) {
	l.log(ctx, slog.LevelInfo, msg, fields)
}

func (l *slogLogger) Error(ctx context.Context, msg string, fields Fields) {
	l.log(ctx, slog.LevelError, msg, fields)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, fields Fields) {
	l.log(ctx, slog.LevelWarn, msg, fields)
}

func (l *slogLogger) Debug(ctx context.Context, msg string, fields Fields) {
	l.log(ctx, slog.LevelDebug, msg, fields)
}

func (l *slogLogger) Close() error {
	return nil
}
