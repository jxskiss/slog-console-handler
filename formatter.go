package betterslog

import (
	"strconv"
	"strings"
	"time"
)

func SourceShortFormatter(source *Source) Attr {
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
	file := source.File
	idx := strings.LastIndexByte(file, '/')
	if idx > 0 {
		idx = strings.LastIndexByte(file[:idx], '/')
	}
	if idx >= 0 {
		file = file[idx+1:]
	}
	value := file + ":" + strconv.Itoa(source.Line)
	return Attr{Key: SourceKey, Value: StringValue(value)}
}

func SourceFullPathFormatter(source *Source) Attr {
	value := source.File + ":" + strconv.Itoa(source.Line)
	return Attr{Key: SourceKey, Value: StringValue(value)}
}

func SourceGroupFormatter(source *Source) Attr {
	var as []Attr
	if source.Function != "" {
		as = append(as, String("function", source.Function))
	}
	if source.File != "" {
		as = append(as, String("file", source.File))
	}
	if source.Line != 0 {
		as = append(as, Int("line", source.Line))
	}
	return Attr{Key: SourceKey, Value: GroupValue(as...)}
}

func TimeShortFormatter(t time.Time) string {
	return t.Format("01/02 15:04:05.000")
}

func TimeLayoutFormatter(layout string) func(time.Time) string {
	return func(t time.Time) string {
		return t.Format(layout)
	}
}
