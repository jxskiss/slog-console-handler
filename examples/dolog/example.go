package dolog

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"strings"
)

func DoLogging() {
	log.Printf("log message without level prefix")
	log.Printf("Debug: log debug message")
	log.Println("[Info] log info message")

	ctx := context.Background()
	err := errors.New("test error")

	slog.Debug("debug message 1")
	slog.DebugContext(ctx, "debug message 2")

	slog.Info(strings.Repeat("info message 1 ", 13))
	slog.InfoContext(ctx, strings.Repeat("info message 2 ", 12))

	slog.Warn("warn message 1")
	slog.WarnContext(ctx, "warn message 2")
	slog.Warn("warn message / error 1", "error", err)
	slog.Warn("warn message / error 2", "err", err)
	slog.Warn("warn message / error 3", slog.Any("error", err),
		slog.String("stacktrace", "dummy\n    stacktrace:xxx\ndummy2\n    some/file.go:883"),
	)

	slog.Error("error message 1")
	slog.ErrorContext(ctx, `error message "xxx" 2`)
	slog.Error(strings.Repeat("error message / error 1 ", 10), "error", err)
	slog.Error("error message / error 2", "err", err)
	slog.Error("error message / error 3", slog.Any("error", err),
		slog.String("stacktrace", "dummy\n    stacktrace:xxx\ndummy2\n    some/file.go:883"),
		slog.String("err2.stacktrace", "dummy\n    stacktrace:xxx\ndummy2\n    some/file.go:883"),
	)

	slog.With(slog.Any("error", err)).Error("error message / with error")

	slog.Info("")
	slog.InfoContext(ctx, "")

	slog.Info("Starting listener", "listen", ":8080", "pid", 37556)
	slog.Debug("Access", "database", "myapp", "host", "localhost:4962", "pid", 37556)
	slog.Info("Access", "method", "GET", "path", "/users", "pid", 37556, "resp_time", 23)
	slog.Info("Access", "method", "POST", "path", "/posts", "pid", 37556, "resp_time", 532)
	slog.Warn("Slow request", "method", "POST", "path", "/posts", "pid", 37556, "resp_time", 532)
	slog.Info("Access", "method", "GET", "path", "/users", "pid", 37556, "resp_time", 10)
	slog.Error("Database connection lost", "error", "connection reset by peer", "database", "myapp", "pid", 37556)

	logger := slog.With("_tag", "subLogger")
	logger.Info("A group of walrus emerges from the ocean", "animal", "walrus", "size", 10)
	logger.Warn("The group's number increased tremendously!", "number", 123, "omg", true)
	logger.Info("A giant walrus appears!", "animal", "walrus", "size", 10)
	logger.Info("Tremendously sized cow enters the ocean.", "animal", "walrus", "size", 9)
	logger.Info("log message with newline\nthe second line log message", "animal", "strawberry", "size", 9, "omg", true)
	logger.Error("The ice breaks!", "number", 100, "omg", true)
}
