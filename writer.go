package slogconsolehandler

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/jxskiss/slog-console-handler/internal/terminal"
)

var (
	byteNewline = []byte("\\n")
	emptyStr    = []byte(`""`)
	equalSpace  = []byte("= ")
	newlineTab2 = "\n\t\t"
	messageKey  = []byte(slog.MessageKey)
	sourceKey   = []byte(slog.SourceKey)
)

func checkIsTerminal(w io.Writer) bool {
	return terminal.CheckIsTerminal(w)
}

type consoleWriter struct {
	writer io.Writer
	color  bool

	buf    buffer
	record bufRecord
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	w.resetBuf()
	w.record.parse(p)
	if len(w.record.Source) > 0 {
		w.record.Fields = append(w.record.Fields, sourceKey, w.record.Source)
		w.record.Source = nil
	}
	w.formatBuf()
	w.buf = bytes.TrimRightFunc(w.buf, unicode.IsSpace)
	w.buf = append(w.buf, '\n')
	_, err = w.writer.Write(w.buf)
	return len(p), err
}

func (w *consoleWriter) resetBuf() {
	w.buf = w.buf[:0]
	w.record = bufRecord{
		Stacktrace: w.record.Stacktrace[:0],
		Errors:     w.record.Errors[:0],
		Fields:     w.record.Fields[:0],
	}
}

func (w *consoleWriter) formatBuf() {
	color := terminal.NoColor
	errColor := terminal.NoColor
	if w.color {
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
		errColor = terminal.Red
	}
	buf := w.buf
	buf.addKeyValue(nil, w.record.Time)
	buf.addWithColor(color, w.record.Level)
	if len(w.record.Level) < 5 {
		buf = append(buf, ' ')
	}
	buf = append(buf, ' ', ' ')
	buf.appendUnquote(messageKey, w.record.Message, "\n")
	if bytes.Contains(w.record.Message, byteNewline) {
		buf = append(buf, '\n', '\t')
	} else {
		buf = append(buf, ' ', '\t')
	}
	for i := 0; i < len(w.record.Errors); i += 2 {
		buf.addWithColor(color, w.record.Errors[i], equalSpace)
		buf.addWithColor(errColor, w.record.Errors[i+1])
	}
	for i := 0; i < len(w.record.Fields); i += 2 {
		buf.addWithColor(color, w.record.Fields[i], equalSpace)
		buf = append(buf, w.record.Fields[i+1]...)
	}
	for i := 0; i < len(w.record.Stacktrace); i += 2 {
		buf = append(buf, '\n', '\t')
		buf.addWithColor(color, w.record.Stacktrace[i], equalSpace)
		buf = append(buf, newlineTab2...)
		buf.appendUnquote(w.record.Stacktrace[i], w.record.Stacktrace[i+1], newlineTab2)
	}
	w.buf = buf
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
		r.addField(key, value)
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
	for len(line) > 0 && line[0] == ' ' {
		line = line[1:]
	}
	r.line = line
	return value
}

func (r *bufRecord) addField(key, value []byte) {
	k := b2s(key)
	switch k {
	case slog.TimeKey:
		if len(r.Time) == 0 {
			if len(value) > 2 && value[0] == '"' {
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

type buffer []byte

func (buf *buffer) addKeyValue(k, v []byte) {
	if len(k) == 0 && len(v) == 0 {
		return
	}
	b := *buf
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ', ' ')
	}
	if len(k) > 0 {
		b = append(b, k...)
		b = append(b, equalSpace...)
	}
	b = append(b, v...)
	*buf = b
}

func (buf *buffer) addWithColor(color terminal.Color, ss ...[]byte) {
	if len(ss) == 0 {
		return
	}
	b := *buf
	if len(b) > 0 && b[len(b)-1] != '\t' {
		b = append(b, ' ', ' ')
	}
	b = color.Append(b, ss...)
	*buf = b
}

func (buf *buffer) appendUnquote(k, v []byte, newlineRepl string) {
	b := *buf
	defer func() { *buf = b }()
	if bytes.Equal(v, emptyStr) {
		b = append(b, emptyStr...)
		return
	}
	const quote = '"'
	s := b2s(v)
	if s[0] != quote || s[len(s)-1] != quote {
		b = append(b, s...)
		return
	}
	s = s[1 : len(s)-1]
	// Handle quoted strings without any escape sequences.
	if !containsByte(s, '\\') && !containsByte(s, '\n') {
		b = append(b, s...)
		return
	}
	// Handle quoted strings with escape sequences.
	for len(s) > 0 {
		r, multibyte, rem, err := strconv.UnquoteChar(s, quote)
		if err != nil {
			bugMsg := fmt.Sprintf("!!![slog-console-handler] bug: value of field %s is not a valid quoted string", k)
			if len(b) > 0 && b[len(b)-1] != ' ' {
				b = append(b, ' ')
			}
			b = append(b, bugMsg...)
			return
		}
		if r < utf8.RuneSelf || !multibyte {
			if r == '\n' {
				b = append(b, newlineRepl...)
			} else {
				b = append(b, byte(r))
			}
		} else {
			b = utf8.AppendRune(b, r)
		}
		s = rem
	}
}

// containsByte reports whether the string containsByte the byte c.
func containsByte(s string, c byte) bool {
	return bytes.IndexByte(s2b(s), c) != -1
}

func b2s(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func s2b(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
