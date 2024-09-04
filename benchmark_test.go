package slogconsolehandler

import (
	"io"
	"log/slog"
	"testing"

	"github.com/jxskiss/slog-console-handler/examples/dolog"
)

func BenchmarkStdSlog(b *testing.B) {
	w := io.Discard
	slog.SetDefault(slog.New(slog.NewTextHandler(w, nil)))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dolog.DoLogging()
		}
	})
}

func BenchmarkConsoleHandler(b *testing.B) {
	w := io.Discard
	slog.SetDefault(slog.New(New(w, nil)))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dolog.DoLogging()
		}
	})
}
