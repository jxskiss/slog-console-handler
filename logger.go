package betterslog

import (
	"bytes"
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
	return log.New(&handlerWriter{h, level, true}, "", 0)
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
