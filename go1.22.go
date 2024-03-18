//go:build go1.22

package bslog

import "log/slog"

func SetLogLoggerLevel(level Level) (oldLevel Level) {
	logLoggerLevel.Set(level)
	return slog.SetLogLoggerLevel(level)
}
