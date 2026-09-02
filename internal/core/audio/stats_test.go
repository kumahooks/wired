package audio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeStats(t *testing.T) {
	t.Parallel()

	t.Run("nil library", func(t *testing.T) {
		t.Parallel()

		stats := ComputeStats(nil)

		assert.Equal(t, 0, stats.FilesCount)
		assert.Equal(t, int64(0), stats.TotalBytes)
		assert.Empty(t, stats.FormatCounts)
	})

	t.Run("empty library", func(t *testing.T) {
		t.Parallel()

		stats := ComputeStats(NewLibrary())

		assert.Equal(t, 0, stats.FilesCount)
		assert.Equal(t, int64(0), stats.TotalBytes)
		assert.Empty(t, stats.FormatCounts)
	})

	t.Run("populated library", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.Add("/music/song1.mp3", 1024)
		library.Add("/music/song2.mp3", 2048)
		library.Add("/music/song3.flac", 4096)

		stats := ComputeStats(library)

		assert.Equal(t, 3, stats.FilesCount)
		assert.Equal(t, int64(7168), stats.TotalBytes)
		assert.Equal(t, map[string]int{".mp3": 2, ".flac": 1}, stats.FormatCounts)
	})
}

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
