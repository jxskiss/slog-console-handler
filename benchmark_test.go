package slogconsolehandler

import (
	"io"
	"log/slog"
	"testing"

	"github.com/jxskiss/slog-console-handler/examples/dolog"
)

func TestConsoleHandler(t *testing.T) {
	slog.SetDefault(slog.New(Default))
	dolog.DoLogging()
}

func BenchmarkSlogTextHandler(b *testing.B) {
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
