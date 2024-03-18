package bslog

import (
	"context"
	"log"
	"log/slog"
	"runtime"
	"time"
)

type (
	Logger = slog.Logger
	Record = slog.Record
	Source = slog.Source
)

func New(h Handler) *Logger {
	return slog.New(h)
}

func NewLogLogger(h Handler, level Level) *log.Logger {
	return slog.NewLogLogger(h, level)
}

func NewRecord(t time.Time, level Level, msg string, pc uintptr) Record {
	return slog.NewRecord(t, level, msg, pc)
}

func Named(l *Logger, name string) *Logger {
	if l == nil {
		l = Default()
	}
	if name == "" {
		return l
	}
	switch impl := l.Handler().(type) {
	case internalHandler:
		subName := getSubLoggerName(impl.getLoggerName(), name)
		handler := impl.cloneHandler()
		handler.setLoggerName(subName)
		return New(handler)
	default:
		return l
	}
}

func getSubLoggerName(parent, sub string) string {
	if parent == "" {
		return sub
	}
	return parent + "." + sub
}

func doLog(handler Handler, ctx context.Context, level Level, msg string, args ...any) {
	if !handler.Enabled(ctx, level) {
		return
	}

	// skip [runtime.Callers, this function, this function's caller]
	const callerSkip = 3
	var pcs [1]uintptr
	runtime.Callers(callerSkip, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = handler.Handle(ctx, r)
}

func doLogAttrs(handler Handler, ctx context.Context, level Level, msg string, attrs ...Attr) {
	if !handler.Enabled(ctx, level) {
		return
	}

	// skip [runtime.Callers, this function, this function's caller]
	const callerSkip = 3
	var pcs [1]uintptr
	runtime.Callers(callerSkip, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = handler.Handle(ctx, r)
}
