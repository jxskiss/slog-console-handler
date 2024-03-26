package betterslog

import (
	"bytes"
	"strings"
	"testing"
)

func TestPerLoggerLeveler(t *testing.T) {
	buf := &bytes.Buffer{}
	filters := []string{
		"betterslog.testfilters=debug",
		"some.pkg2=error",
	}
	ppl, err := NewPerLoggerLeveler(LevelInfo, filters)
	if err != nil {
		t.Fatal(err)
	}

	logger := New(NewTextHandler(buf, &HandlerOptions{
		AddSource: true,
		Level:     ppl,
	}))

	lg1 := Named(logger, "betterslog")
	lg1.Debug("debug message 1") // no
	lg1.Info("info message 2")   // yes

	lg1 = Named(lg1, "testfilters")
	lg1.Debug("debug message 3") // yes
	lg1.Info("info message 4")   // yes

	lg2 := Named(logger, "some.pkg2")
	lg2.Info("info message 5")   // no
	lg2.Warn("warn message 6")   // no
	lg2.Error("error message 7") // yes

	got := buf.String()
	for _, str := range []string{
		"info message 2",
		"debug message 3",
		"info message 4",
		"error message 7",
	} {
		if !strings.Contains(got, str) {
			t.Errorf("should but not conatains msg %q", str)
		}
	}
	for _, str := range []string{
		"debug message 1",
		"info message 5",
		"warn message 6",
	} {
		if strings.Contains(got, str) {
			t.Errorf("should not but contains msg %q", str)
		}
	}
}
