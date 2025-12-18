package logger

import (
	"context"
	"sort"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger is the zap implementation of Logger
type zapLogger struct {
	logger *zap.Logger
	ctxFn  ContextFieldsFunc
}

func newZapLogger(cfg *config) *zapLogger {
	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = ""
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create core
	// We map slog.Level to zapcore.Level
	var zapLevel zapcore.Level
	switch cfg.level {
	case LogLevelDebug:
		zapLevel = zapcore.DebugLevel
	case LogLevelInfo:
		zapLevel = zapcore.InfoLevel
	case LogLevelWarn:
		zapLevel = zapcore.WarnLevel
	default:
		zapLevel = zapcore.ErrorLevel
	}

	atom := zap.NewAtomicLevel()
	atom.SetLevel(zapLevel)

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// WriteSyncer
	ws := zapcore.AddSync(cfg.output)

	core := zapcore.NewCore(encoder, ws, atom)

	// Create logger with AddCaller and AddCallerSkip
	// Skip 1 frame: zapLogger.<Level>() wrapper -> zap implementation
	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{logger: z, ctxFn: cfg.ctxFn}
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields Fields) {
	if l.ctxFn != nil {
		fields = mergeFields(l.ctxFn(ctx), fields)
	}
	l.logger.Info(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields Fields) {
	if l.ctxFn != nil {
		fields = mergeFields(l.ctxFn(ctx), fields)
	}
	l.logger.Error(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields Fields) {
	if l.ctxFn != nil {
		fields = mergeFields(l.ctxFn(ctx), fields)
	}
	l.logger.Warn(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields Fields) {
	if l.ctxFn != nil {
		fields = mergeFields(l.ctxFn(ctx), fields)
	}
	l.logger.Debug(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) toZapFields(fields Fields) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	zf := make([]zap.Field, 0, len(keys))
	for _, k := range keys {
		zf = append(zf, zap.Any(k, fields[k]))
	}
	return zf
}

func (l *zapLogger) Close() error {
	return l.logger.Sync()
}
