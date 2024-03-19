package betterslog

import (
	"context"
	"log"
	"log/slog"
)

var logLoggerLevel LevelVar

func SetDefault(l *Logger) {
	addSource := false
	switch impl := l.Handler().(type) {
	case internalHandler:
		addSource = impl.getOptions().AddSource
	default:
		addSource = log.Flags()&(log.Lshortfile|log.Llongfile) != 0
	}
	slog.SetDefault(l)
	if log.Flags() == 0 {
		log.SetOutput(&handlerWriter{l.Handler(), &logLoggerLevel, addSource})
	}
}

func Default() *Logger {
	return slog.Default()
}

func With(attrs ...any) *Logger {
	return Default().With(attrs...)
}

func Debug(msg string, args ...any) {
	ctx := context.Background()
	doLog(Default().Handler(), ctx, LevelDebug, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	doLog(Default().Handler(), ctx, LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	ctx := context.Background()
	doLog(Default().Handler(), ctx, LevelInfo, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	doLog(Default().Handler(), ctx, LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	ctx := context.Background()
	doLog(Default().Handler(), ctx, LevelWarn, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	doLog(Default().Handler(), ctx, LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	ctx := context.Background()
	doLog(Default().Handler(), ctx, LevelError, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	doLog(Default().Handler(), ctx, LevelError, msg, args...)
}

func Log(ctx context.Context, level Level, msg string, args ...any) {
	doLog(Default().Handler(), ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	doLogAttrs(Default().Handler(), ctx, level, msg, attrs...)
}
