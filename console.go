package betterslog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/jxskiss/better-slog/internal/terminal"
)

type ConsoleHandler struct {
	inner *TextHandler
}

func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *ConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.inner.Handle(ctx, record)
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ConsoleHandler{
		inner: h.inner.WithAttrs(attrs).(*TextHandler),
	}
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		inner: h.inner.WithGroup(name).(*TextHandler),
	}
}

func (h *ConsoleHandler) cloneHandler() internalHandler {
	return &ConsoleHandler{
		inner: h.inner.clone(),
	}
}
func (h *ConsoleHandler) getOptions() HandlerOptions { return h.inner.getOptions() }
func (h *ConsoleHandler) getLoggerName() string      { return h.inner.getLoggerName() }
func (h *ConsoleHandler) setLoggerName(name string)  { h.inner.setLoggerName(name) }

func NewConsoleHandler(w io.Writer, opts *HandlerOptions) *ConsoleHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	opts.timeFormatter = func(t time.Time) string {
		return t.Format("01/02 15:04:05.000")
	}
	cw := &consoleWriter{
		writer:      w,
		enableColor: !opts.NoColor && terminal.CheckIsTerminal(w),
		buf:         make([]byte, 1024), // 1KB buffer
	}
	inner := NewTextHandler(cw, opts)
	return &ConsoleHandler{inner}
}

type consoleWriter struct {
	writer      io.Writer
	enableColor bool

	buf    []byte
	keyBuf []byte
	record bufRecord
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	w.resetBuf()
	w.record.parse(p)
	if !w.enableColor {
		w.formatNoColor()
	} else {
		w.formatColorized()
	}
	w.buf = bytes.TrimRightFunc(w.buf, unicode.IsSpace)
	w.buf = append(w.buf, '\n')
	_, err = w.writer.Write(w.buf)
	return len(p), err
}

func (w *consoleWriter) resetBuf() {
	w.buf = w.buf[:0]
	w.keyBuf = w.keyBuf[:0]
	w.record = bufRecord{
		Errors: w.record.Errors[:0],
		Others: w.record.Others[:0],
	}
}

func (w *consoleWriter) formatNoColor() {
	buf := w.buf
	buf = addToBuf(buf, nil, w.record.Time)
	buf = addToBuf(buf, nil, w.record.Level)
	buf = addToBuf(buf, nil, w.record.Source)
	buf = append(buf, '\t')
	buf = addToBuf(buf, nil, w.record.Message)
	buf = append(buf, ' ', '\t')
	for i := 0; i < len(w.record.Errors); i += 2 {
		buf = addToBuf(buf, w.record.Errors[i], w.record.Errors[i+1])
	}
	for i := 0; i < len(w.record.Others); i += 2 {
		buf = addToBuf(buf, w.record.Others[i], w.record.Others[i+1])
	}
	w.buf = buf
}

func (w *consoleWriter) formatColorized() {
	color := terminal.NoColor
	level := unsafe.String(unsafe.SliceData(w.record.Level), len(w.record.Level))
	switch {
	case strings.HasPrefix(level, "DEBUG"):
		color = terminal.Magenta
	case strings.HasPrefix(level, "INFO"):
		color = terminal.Cyan
	case strings.HasPrefix(level, "WARN"):
		color = terminal.Yellow
	case strings.HasPrefix(level, "ERROR"):
		color = terminal.Red
	}
	buf, keyBuf := w.buf, w.keyBuf
	buf = addToBuf(buf, nil, w.record.Time)
	buf = addWithColor(buf, w.record.Level, color)
	buf = addToBuf(buf, nil, w.record.Source)
	buf = append(buf, '\t')
	buf = addToBuf(buf, nil, w.record.Message)
	buf = append(buf, ' ', '\t')
	for i := 0; i < len(w.record.Errors); i += 2 {
		keyBuf = keyBuf[:0]
		keyBuf = append(keyBuf, w.record.Errors[i]...)
		keyBuf = append(keyBuf, '=', ' ')
		buf = addWithColor(buf, keyBuf, color)
		buf = terminal.Red.Append(buf, w.record.Errors[i+1])
	}
	for i := 0; i < len(w.record.Others); i += 2 {
		keyBuf = keyBuf[:0]
		keyBuf = append(keyBuf, w.record.Others[i]...)
		keyBuf = append(keyBuf, '=', ' ')
		buf = addWithColor(buf, keyBuf, color)
		buf = append(buf, w.record.Others[i+1]...)
	}
	w.buf, w.keyBuf = buf, keyBuf
}

type bufRecord struct {
	line []byte

	Time    []byte
	Level   []byte
	Source  []byte
	Message []byte
	Errors  [][]byte
	Others  [][]byte
}

func (r *bufRecord) parse(line []byte) {
	line = bytes.TrimRight(line, "\n")
	if len(line) == 0 {
		return
	}

	r.line = line

	var key, value []byte
	for len(r.line) > 0 {
		key = r.getKey()
		if len(key) == 0 {
			return
		}
		value = r.getValue()
		r.addKeyValue(key, value)
	}
}

func (r *bufRecord) getKey() (key []byte) {
	line := r.line
	sepIdx := bytes.IndexByte(line, '=')
	if sepIdx <= 0 {
		key = line
		line = nil
	} else {
		key = line[:sepIdx]
		line = line[sepIdx+1:]
	}
	r.line = line
	return key
}

var emptyStr = []byte(`""`)

func (r *bufRecord) getValue() (value []byte) {
	line := r.line
	if len(line) == 0 {
		return emptyStr
	}
	endIdx := len(line)
	if line[0] != '"' {
		for i := 0; i < len(line); i++ {
			if line[i] == ' ' {
				endIdx = i
				break
			}
		}
		value = line[:endIdx]
		line = line[endIdx:]
	} else {
		for i := 1; i < len(line); i++ {
			if line[i] == '"' && line[i-1] != '\\' {
				endIdx = i
				break
			}
		}
		value = line[:endIdx+1]
		line = line[endIdx+1:]
	}
	if len(line) > 0 && line[0] == ' ' {
		line = line[1:]
	}
	r.line = line
	return value
}

func (r *bufRecord) addKeyValue(key, value []byte) {
	k := unsafe.String(unsafe.SliceData(key), len(key))
	switch k {
	case TimeKey:
		if len(r.Time) == 0 {
			if value[0] == '"' {
				value = value[1 : len(value)-1]
			}
			r.Time = value
			return
		}
	case LevelKey:
		if len(r.Level) == 0 {
			r.Level = value
			return
		}
	case SourceKey:
		if len(r.Source) == 0 {
			r.Source = value
			return
		}
	case MessageKey:
		if len(r.Message) == 0 {
			r.Message = value
			return
		}
	}
	if k == "err" || k == "error" ||
		strings.HasSuffix(k, ".err") ||
		strings.HasSuffix(k, ".error") {
		r.Errors = append(r.Errors, key, value)
	} else {
		r.Others = append(r.Others, key, value)
	}
}

func addToBuf[T string | []byte](b []byte, k, v T) []byte {
	if len(k) == 0 && len(v) == 0 {
		return b
	}
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ')
	}
	if len(k) > 0 {
		b = append(b, k...)
		b = append(b, '=')
	}
	b = append(b, v...)
	return b
}

func addWithColor(b []byte, s []byte, color terminal.Color) []byte {
	if len(s) == 0 {
		return b
	}
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ')
	}
	b = color.Append(b, s)
	return b
}
