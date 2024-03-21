package betterslog

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestNewLogLogger(t *testing.T) {
	printLogs := func(l *log.Logger) {
		for _, x := range []string{
			"Trace", "Debug", "Info", "Notice", "Warn", "Warning", "Error", "Fatal",
		} {
			l.Printf("[%s] %s", x, x)
			l.Printf("%s: %s", x, x)
		}
		l.Print("message without level prefix")
	}

	var levelVar LevelVar
	levelVar.Set(LevelDebug)
	buf := bytes.NewBuffer(nil)
	l := NewLogLogger(NewTextHandler(buf, &HandlerOptions{
		Level: &levelVar,
	}), LevelInfo)
	printLogs(l)
	got := buf.String()
	t.Logf("got: %s", got)
	for _, want := range []string{
		`level=DEBUG msg="[Trace] Trace"`,
		`level=DEBUG msg="[Debug] Debug"`,
		`level=INFO msg="[Info] Info"`,
		`level=WARN msg="[Notice] Notice"`,
		`level=WARN msg="[Warn] Warn"`,
		`level=WARN msg="[Warning] Warning"`,
		`level=ERROR msg="[Error] Error"`,
		`level=ERROR msg="[Fatal] Fatal"`,

		`level=DEBUG msg="Trace: Trace"`,
		`level=DEBUG msg="Debug: Debug"`,
		`level=INFO msg="Info: Info"`,
		`level=WARN msg="Notice: Notice"`,
		`level=WARN msg="Warn: Warn"`,
		`level=WARN msg="Warning: Warning"`,
		`level=ERROR msg="Error: Error"`,
		`level=ERROR msg="Fatal: Fatal"`,

		`level=INFO msg="message without level prefix"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("not found but want %q in logger output", want)
		}
	}

	levelVar.Set(LevelWarn)
	buf.Reset()
	printLogs(l)
	got = buf.String()
	t.Logf("got: %s", got)
	for _, want := range []string{
		`level=WARN msg="[Notice] Notice"`,
		`level=WARN msg="[Warn] Warn"`,
		`level=WARN msg="[Warning] Warning"`,
		`level=ERROR msg="[Error] Error"`,
		`level=ERROR msg="[Fatal] Fatal"`,

		`level=WARN msg="Notice: Notice"`,
		`level=WARN msg="Warn: Warn"`,
		`level=WARN msg="Warning: Warning"`,
		`level=ERROR msg="Error: Error"`,
		`level=ERROR msg="Fatal: Fatal"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("not found but want %q in logger output", want)
		}
	}
	for _, notWant := range []string{
		`level=DEBUG msg="[Trace] Trace"`,
		`level=DEBUG msg="[Debug] Debug"`,
		`level=INFO msg="[Info] Info"`,

		`level=DEBUG msg="Trace: Trace"`,
		`level=DEBUG msg="Debug: Debug"`,
		`level=INFO msg="Info: Info"`,

		`level=INFO msg="message without level prefix"`,
	} {
		if strings.Contains(got, notWant) {
			t.Errorf("found but not want %q in logger output", notWant)
		}
	}
}

func TestNamed(t *testing.T) {
	var buf bytes.Buffer
	SetDefault(New(NewTextHandler(&buf, &HandlerOptions{
		AddLogger: true,
	})))
	logger1 := Named(nil, "l1")
	logger2 := Named(logger1, "l2")
	logger3 := Named(logger1, "l3")
	logger4 := Named(logger3, "l4")
	Info("default 1")
	logger1.Info("logger 1")
	logger2.Info("logger 2")
	logger3.Info("logger 3")
	logger4.Info("logger 4")
	got := buf.String()
	t.Logf("got: %s", got)
	for _, want := range []string{
		"level=INFO msg=\"default 1\"\n",
		`level=INFO msg="logger 1" logger=l1`,
		`level=INFO msg="logger 2" logger=l1.l2`,
		`level=INFO msg="logger 3" logger=l1.l3`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("not found but want %q in logger output", want)
		}
	}
}
