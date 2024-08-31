package slogconsolehandler

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var levelVar = &slog.LevelVar{}

func init() { levelVar.Set(slog.LevelDebug) }

// Default is a default handler configured at debug level,
// color is enabled, time and source are formatted in a short form.
// The level can be changed on-the-fly by calling SetLevel.
var Default = New(os.Stderr, &HandlerOptions{
	AddSource: true,
	Level:     levelVar,
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			if a.Value.Kind() == slog.KindTime {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, formatTimeShort(t))
				}
			}
		case slog.SourceKey:
			if a.Value.Kind() == slog.KindAny {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					return slog.String(slog.SourceKey, formatSourceShort(src))
				}
			}
		}
		return a
	},
	DisableColor: false,
})

// SetLevel sets the Default handler's level to l.
func SetLevel(l slog.Level) { levelVar.Set(l) }

func formatTimeShort(t time.Time) string {
	return t.Format("01/02 15:04:05.000")
}

func formatSourceShort(s *slog.Source) string {
	// nb. To make sure we trim the path correctly on Windows too,
	// we counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.
	//
	// Find the last separator.
	//
	file := s.File
	idx := strings.LastIndexByte(file, '/')
	if idx > 0 {
		idx = strings.LastIndexByte(file[:idx], '/')
	}
	if idx >= 0 {
		file = file[idx+1:]
	}
	value := file + ":" + strconv.Itoa(s.Line)
	return value
}
