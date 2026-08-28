package testutil

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var updateGolden = flag.Bool("update", false, "regenerate render golden files")

// AssertSnapshot compares the ANSI-stripped content against testdata/<name>.golden, writing the file when -update is set.
func AssertSnapshot(t *testing.T, name string, content string) {
	t.Helper()

	goldenPath := "testdata/" + name + ".golden"
	stripped := StripANSI(content)

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

	assert.Equal(t, string(want), stripped, "snapshot %q mismatch", name)
}
