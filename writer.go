package slogconsolehandler

import (
	"bytes"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/jxskiss/slog-console-handler/internal/terminal"
)

var (
	byteNewline = []byte("\\n")
	sourceKey   = []byte("source")
)

func checkIsTerminal(w io.Writer) bool {
	return terminal.CheckIsTerminal(w)
}

type consoleWriter struct {
	writer io.Writer
	color  bool

	buf    []byte
	keyBuf []byte
	record bufRecord
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	w.resetBuf()
	w.record.parse(p)
	if len(w.record.Source) > 0 {
		w.record.Fields = append(w.record.Fields, sourceKey, w.record.Source)
		w.record.Source = nil
	}
	if !w.color {
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
		Fields: w.record.Fields[:0],
	}
}

func (w *consoleWriter) formatNoColor() {
	buf := w.buf
	buf = addToBuf(buf, nil, w.record.Time)
	buf = addToBuf(buf, nil, w.record.Level)
	if len(w.record.Level) < 5 {
		buf = append(buf, ' ')
	}
	buf = addToBuf(buf, nil, w.record.Message)
	buf = append(buf, ' ', '\t')
	for i := 0; i < len(w.record.Errors); i += 2 {
		buf = addToBuf(buf, w.record.Errors[i], w.record.Errors[i+1])
	}
	for i := 0; i < len(w.record.Fields); i += 2 {
		buf = addToBuf(buf, w.record.Fields[i], w.record.Fields[i+1])
	}
	for i := 0; i < len(w.record.Stacktrace); i += 2 {
		buf = append(buf, '\n', '\t')
		buf = append(buf, w.record.Stacktrace[i]...)
		buf = append(buf, '=', ' ')
		buf = formatStacktrace(buf, w.record.Stacktrace[i+1])
	}
	w.buf = buf
}

func (w *consoleWriter) formatColorized() {
	color := terminal.NoColor
	level := b2s(w.record.Level)
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
	buf, buf1 := w.buf, w.keyBuf
	buf = addToBuf(buf, nil, w.record.Time)
	buf = addWithColor(buf, w.record.Level, color)
	if len(w.record.Level) < 5 {
		buf = append(buf, ' ')
	}
	buf = addToBuf(buf, nil, w.record.Message)
	buf = append(buf, ' ', '\t')
	for i := 0; i < len(w.record.Errors); i += 2 {
		buf1 = buf1[:0]
		buf1 = append(buf1, w.record.Errors[i]...)
		buf1 = append(buf1, '=', ' ')
		buf = addWithColor(buf, buf1, color)
		buf = terminal.Red.Append(buf, w.record.Errors[i+1])
	}
	for i := 0; i < len(w.record.Fields); i += 2 {
		buf1 = buf1[:0]
		buf1 = append(buf1, w.record.Fields[i]...)
		buf1 = append(buf1, '=', ' ')
		buf = addWithColor(buf, buf1, color)
		buf = append(buf, w.record.Fields[i+1]...)
	}
	for i := 0; i < len(w.record.Stacktrace); i += 2 {
		buf = append(buf, '\n', '\t')
		buf1 = buf1[:0]
		buf1 = append(buf1, w.record.Stacktrace[i]...)
		buf1 = append(buf1, '=', ' ')
		buf = addWithColor(buf, buf1, color)
		buf = formatStacktrace(buf, w.record.Stacktrace[i+1])
	}
	w.buf, w.keyBuf = buf, buf1
}

type bufRecord struct {
	line []byte

	Time       []byte
	Level      []byte
	Source     []byte
	Message    []byte
	Stacktrace [][]byte
	Errors     [][]byte
	Fields     [][]byte
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
	k := b2s(key)
	switch k {
	case slog.TimeKey:
		if len(r.Time) == 0 {
			if value[0] == '"' {
				value = value[1 : len(value)-1]
			}
			r.Time = value
			return
		}
	case slog.LevelKey:
		if len(r.Level) == 0 {
			r.Level = value
			return
		}
	case slog.SourceKey:
		if len(r.Source) == 0 {
			r.Source = value
			return
		}
	case slog.MessageKey:
		if len(r.Message) == 0 {
			if len(value) > 2 && value[0] == '"' {
				value = value[1 : len(value)-1]
			}
			r.Message = value
			return
		}
	}
	if (strings.Contains(k, "stack") || strings.Contains(k, "trace")) && bytes.Contains(value, byteNewline) {
		r.Stacktrace = append(r.Stacktrace, key, value)
	} else if k == "err" || k == "error" ||
		strings.HasSuffix(k, ".err") ||
		strings.HasSuffix(k, ".error") {
		r.Errors = append(r.Errors, key, value)
	} else {
		r.Fields = append(r.Fields, key, value)
	}
}

func addToBuf[T string | []byte](b []byte, k, v T) []byte {
	if len(k) == 0 && len(v) == 0 {
		return b
	}
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ', ' ')
	}
	if len(k) > 0 {
		b = append(b, k...)
		b = append(b, '=', ' ')
	}
	b = append(b, v...)
	return b
}

func addWithColor(b []byte, s []byte, color terminal.Color) []byte {
	if len(s) == 0 {
		return b
	}
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ', ' ')
	}
	b = color.Append(b, s)
	return b
}

func formatStacktrace(b []byte, st []byte) []byte {
	st = bytes.TrimSpace(st)
	s, _ := strconv.Unquote(b2s(st))
	b = append(b, '\n', '\t', '\t')
	i, n := 0, len(s)
	for j, x := range s {
		if x != '\n' && j != n-1 {
			continue
		}
		b = append(b, string([]rune(s)[i:j])...)
		if j < n-1 {
			b = append(b, '\n', '\t', '\t')
		}
		i = j + 1
	}
	return b
}

func b2s(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
