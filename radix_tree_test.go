package betterslog

import (
	"bytes"
	"strings"
	"testing"
)

func TestRadixTree(t *testing.T) {
	root := &radixNode[Level]{}
	root.insert("some.module_2.pkg_1", LevelInfo)
	root.insert("some.module_2.pkg_2", LevelDebug)
	root.insert("betterslog.filtertest", LevelWarn)
	root.insert("some.module_1", LevelError)
	root.insert("some.module_1.pkg_1", LevelWarn)

	wantDump := `betterslog.filtertest=WARN
some.module_1=ERROR
some.module_1.pkg_1=WARN
some.module_2.pkg_1=INFO
some.module_2.pkg_2=DEBUG
`
	if got := root.dumpTree(""); got != wantDump {
		t.Fatalf("radixTree dump not euqal, want %v, got %v", wantDump, got)
	}

	testcases := []struct {
		Name  string
		Level Level
		Found bool
	}{
		{"some.module_1", LevelError, true},
		{"some.module_1.pkg_0", LevelError, true},
		{"some.module_1.pkg_1", LevelWarn, true},
		{"some.module_2", LevelInfo, false},
		{"some.module_2.pkg_0", LevelInfo, false},
		{"some.module_2.pkg_1", LevelInfo, true},
		{"some.module_2.pkg_2", LevelDebug, true},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			level, found := root.search(tc.Name)
			if tc.Found != found {
				t.Errorf("radixTree.search result found, want %v but got %v", tc.Found, found)
			}
			if tc.Level != level {
				t.Errorf("radixTree.search result level, want %v bot got %v", tc.Level, level)
			}
		})
	}
}

func TestPerLoggerLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	filters := []string{
		"betterslog.testfilters=debug",
		"some.pkg2=error",
	}
	ppl, err := NewPerLoggerLevel(LevelInfo, filters)
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
