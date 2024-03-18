package bslog

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"runtime"
	"time"
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

type handlerWriter struct {
	handler   Handler
	level     Leveler
	addSource bool
}

func (w *handlerWriter) Write(buf []byte) (int, error) {
	level, ok := w.detectLevel(buf)
	if !ok {
		level = w.level.Level()
	}
	if !w.handler.Enabled(context.Background(), level) {
		return 0, nil
	}

	var pc uintptr
	if w.addSource {
		// skip [runtime.Callers, w.Write, Logger.Output, log.Print]
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		pc = pcs[0]
	}

	// Remove final newline.
	origLen := len(buf) // Report that the entire buf was written.
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	r := NewRecord(time.Now(), level, string(buf), pc)
	return origLen, w.handler.Handle(context.Background(), r)
}

var levelNames = []byte("trace debug info notice warn warning error")

func (w *handlerWriter) detectLevel(msg []byte) (Level, bool) {
	const levelPrefixMinLen = 5
	const tIdx, dIdx, iIdx, nIdx, wIdx1, wIdx2, eIdx = 0, 6, 12, 17, 24, 29, 39
	if len(msg) < levelPrefixMinLen {
		return 0, false
	}
	end := uint8(':')
	if msg[0] == '[' {
		end, msg = ']', msg[1:]
	}
	switch msg[0] {
	//case 'T', 't':
	//	if len(msg) > 5 && msg[5] == end &&
	//		bytes.Equal(levelNames[tIdx:tIdx+5], msg[:5]) {
	//		return LevelTrace, true
	//	}
	case 'D', 'd':
		if len(msg) > 5 && msg[5] == end &&
			bytes.EqualFold(levelNames[dIdx:dIdx+5], msg[:5]) {
			return LevelDebug, true
		}
	case 'I', 'i':
		if len(msg) > 4 && msg[4] == end &&
			bytes.EqualFold(levelNames[iIdx:iIdx+4], msg[:4]) {
			return LevelInfo, true
		}
	//case 'N', 'n':
	//	if len(msg) > 6 && msg[6] == end &&
	//		bytes.EqualFold(levelNames[nIdx:nIdx+6], msg[:6]) {
	//		return LevelNotice, true
	//	}
	case 'W', 'w':
		if len(msg) > 4 && msg[4] == end &&
			bytes.EqualFold(levelNames[wIdx1:wIdx1+4], msg[:4]) {
			return LevelWarn, true
		}
		if len(msg) > 7 && msg[7] == end &&
			bytes.EqualFold(levelNames[wIdx2:wIdx2+7], msg[:7]) {
			return LevelWarn, true
		}
	case 'E', 'e':
		if len(msg) > 5 && msg[5] == end &&
			bytes.EqualFold(levelNames[eIdx:eIdx+5], msg[:5]) {
			return LevelError, true
		}
	}
	return 0, false
}
