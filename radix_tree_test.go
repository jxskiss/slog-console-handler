package betterslog

import (
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
