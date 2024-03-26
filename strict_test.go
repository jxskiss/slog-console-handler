package betterslog

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStrictLogger(t *testing.T) {
	var buf bytes.Buffer
	SetDefault(New(NewTextHandler(&buf, &HandlerOptions{Level: LevelDebug})))

	ctx := context.Background()
	printLogs := func(l StrictLogger) {
		l.Debug(ctx, "Debug message", "k1", "v1", "k2", 234)
		l.DebugAttr(ctx, "DebugAttr message", String("k1", "v1"), Int("k2", 234))
		l.Info(ctx, "Info message", String("k1", "v1"), Int("k2", 234))
		l.Warn(ctx, "Warn message", String("k1", "v1"), Int("k2", 234))
		l.Error(ctx, errors.New("test error"), "Error message",
			String("k1", "v1"), Int("k2", 234))
		l.Log(ctx, LevelError, "Log message",
			String("k1", "v1"), Int("k2", 234))
	}

	t.Run("root logger", func(t *testing.T) {
		buf.Reset()
		lg := Strict()
		printLogs(lg)
		got := buf.String()
		t.Logf("got: %s", got)

		for _, want := range []string{
			`level=DEBUG msg="Debug message" k1=v1 k2=234`,
			`level=DEBUG msg="DebugAttr message" k1=v1 k2=234`,
			`level=INFO msg="Info message" k1=v1 k2=234`,
			`level=WARN msg="Warn message" k1=v1 k2=234`,
			`level=ERROR msg="Error message" error="test error" k1=v1 k2=234`,
			`level=ERROR msg="Log message" k1=v1 k2=234`,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("not found but want %q in logger output", want)
			}
		}
	})

	t.Run("with attrs", func(t *testing.T) {
		buf.Reset()
		lg := Strict().With(String("k0", "v0"))
		printLogs(lg)
		got := buf.String()
		t.Logf("got: %s", got)

		for _, want := range []string{
			`level=DEBUG msg="Debug message" k0=v0 k1=v1 k2=234`,
			`level=DEBUG msg="DebugAttr message" k0=v0 k1=v1 k2=234`,
			`level=INFO msg="Info message" k0=v0 k1=v1 k2=234`,
			`level=WARN msg="Warn message" k0=v0 k1=v1 k2=234`,
			`level=ERROR msg="Error message" k0=v0 error="test error" k1=v1 k2=234`,
			`level=ERROR msg="Log message" k0=v0 k1=v1 k2=234`,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("not found but want %q in logger output", want)
			}
		}
	})

	t.Run("with group", func(t *testing.T) {
		buf.Reset()
		lg := Strict().With(String("k0", "v0")).WithGroup("group1")
		printLogs(lg)
		got := buf.String()
		t.Logf("got: %s", got)

		for _, want := range []string{
			`level=DEBUG msg="Debug message" k0=v0 group1.k1=v1 group1.k2=234`,
			`level=DEBUG msg="DebugAttr message" k0=v0 group1.k1=v1 group1.k2=234`,
			`level=INFO msg="Info message" k0=v0 group1.k1=v1 group1.k2=234`,
			`level=WARN msg="Warn message" k0=v0 group1.k1=v1 group1.k2=234`,
			`level=ERROR msg="Error message" k0=v0 group1.error="test error" group1.k1=v1 group1.k2=234`,
			`level=ERROR msg="Log message" k0=v0 group1.k1=v1 group1.k2=234`,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("not found but want %q in logger output", want)
			}
		}
	})
}
