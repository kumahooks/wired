package audio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"negative", -5, "0 B"},
		{"bytes", 512, "512 B"},
		{"kib boundary", 1 << 10, "1.0 KiB"},
		{"kib", 1536, "1.5 KiB"},
		{"mib", 3<<20 + 400<<10, "3.4 MiB"},
		{"gib", 5 << 30, "5.0 GiB"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, GetReadableByteSize(test.bytes))
		})
	}
}
