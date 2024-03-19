//go:build go1.22

package betterslog

import "log/slog"

func SetLogLoggerLevel(level Level) (oldLevel Level) {
	logLoggerLevel.Set(level)
	return slog.SetLogLoggerLevel(level)
}
