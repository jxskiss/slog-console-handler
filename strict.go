package betterslog

import (
	"context"
	"log/slog"
)

// StrictLogger contains a limited set of methods from slog.Logger.
// This forces adding context.Context and using Attr.
type StrictLogger interface {

	// Named adds a new path segment to the logger's name.
	// Segments are joined by periods.
	// By default, Loggers are unnamed.
	Named(name string) StrictLogger

	With(attrs ...Attr) StrictLogger
	WithGroup(name string) StrictLogger

	Debug(ctx context.Context, msg string, args ...any)
	DebugAttr(ctx context.Context, msg string, attrs ...Attr)
	Info(ctx context.Context, msg string, attrs ...Attr)
	Warn(ctx context.Context, msg string, attrs ...Attr)
	Error(ctx context.Context, err error, msg string, attrs ...Attr)

	// Log ... for custom levels.
	Log(ctx context.Context, level Level, msg string, attrs ...Attr)

	Handler() Handler
	ToLogger() *Logger
}

// Strict returns the default logger as a StrictLogger.
func Strict() StrictLogger {
	return ToStrict(Default())
}

// ToStrict converts a Logger to a StrictLogger.
func ToStrict(l *Logger) StrictLogger {
	return &strictLogger{l.Handler()}
}

type strictLogger struct {
	handler slog.Handler
}

func (l *strictLogger) Named(name string) StrictLogger {
	switch impl := l.handler.(type) {
	case internalHandler:
		subName := getSubLoggerName(impl.getLoggerName(), name)
		handler := impl.cloneHandler()
		handler.setLoggerName(subName)
		return &strictLogger{handler}
	default:
		return l
	}
}

func (l *strictLogger) With(attrs ...Attr) StrictLogger {
	return &strictLogger{l.handler.WithAttrs(attrs)}
}

func (l *strictLogger) WithGroup(name string) StrictLogger {
	return &strictLogger{l.handler.WithGroup(name)}
}

func (l *strictLogger) Log(ctx context.Context, level Level, msg string, attrs ...Attr) {
	doLogAttrs(l.handler, ctx, level, msg, attrs...)
}

func (l *strictLogger) Debug(ctx context.Context, msg string, args ...any) {
	doLog(l.handler, ctx, LevelDebug, msg, args...)
}

func (l *strictLogger) DebugAttr(ctx context.Context, msg string, attrs ...Attr) {
	doLogAttrs(l.handler, ctx, LevelDebug, msg, attrs...)
}

func (l *strictLogger) Info(ctx context.Context, msg string, attrs ...Attr) {
	doLogAttrs(l.handler, ctx, LevelInfo, msg, attrs...)
}

func (l *strictLogger) Warn(ctx context.Context, msg string, attrs ...Attr) {
	doLogAttrs(l.handler, ctx, LevelWarn, msg, attrs...)
}

func (l *strictLogger) Error(ctx context.Context, err error, msg string, attrs ...Attr) {
	if err != nil {
		asp := attrSlicePool.Get().(*attrSlice)
		defer asp.Free()
		*asp = append(*asp, Err(err))
		*asp = append(*asp, attrs...)
		attrs = *asp
	}
	doLogAttrs(l.handler, ctx, LevelError, msg, attrs...)
}

func (l *strictLogger) Handler() Handler {
	return l.handler
}

func (l *strictLogger) ToLogger() *Logger {
	return slog.New(l.handler)
}
