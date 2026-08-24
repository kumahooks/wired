package initializing

import (
	"flag"
	"os"
	"regexp"
	"testing"
)

var updateGolden = flag.Bool("update", false, "regenerate render golden files")

// ansiEscSeq matches CSI sequences (the only escape sequences lipgloss emits for styling).
var ansiEscSeq = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

// stripANSI removes CSI escape sequences so golden files compare against the plain text layout.
func stripANSI(value string) string {
	return ansiEscSeq.ReplaceAllString(value, "")
}

// assertSnapshot compares the stripped content against testdata/<name>.golden, writing the file when -update is set.
func assertSnapshot(t *testing.T, name string, content string) {
	t.Helper()

	goldenPath := "testdata/" + name + ".golden"
	stripped := stripANSI(content)

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}

		if err := os.WriteFile(goldenPath, []byte(stripped), 0o644); err != nil {
			t.Fatalf("write golden %q: %v", goldenPath, err)
		}

		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %q: %v (run with -update to generate)", goldenPath, err)
	}

	if stripped != string(want) {
		t.Errorf("snapshot %q mismatch:\nwant:\n%s\ngot:\n%s", name, want, stripped)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
