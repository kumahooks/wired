package audio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeStatsEmpty(t *testing.T) {
	t.Parallel()

	stats := ComputeStats(nil)

	assert.Equal(t, 0, stats.FilesCount)
	assert.Empty(t, stats.FormatCounts)
	assert.Equal(t, int64(0), stats.TotalBytes)
}

func TestComputeStatsMixedFormats(t *testing.T) {
	t.Parallel()

	files := []File{
		{FileName: "uwu.mp3", SizeBytes: 100},
		{FileName: "owo.mp3", SizeBytes: 200},
		{FileName: "lain.flac", SizeBytes: 1000},
		{FileName: "wired.MP3", SizeBytes: 1},
		{FileName: "uwu?", SizeBytes: 10},
		{FileName: "yep.ogg", SizeBytes: 5},
	}

	stats := ComputeStats(files)

	assert.Equal(t, 6, stats.FilesCount)
	assert.Equal(t, int64(1316), stats.TotalBytes)
	assert.Equal(t, map[string]int{".mp3": 3, ".flac": 1, ".ogg": 1, "": 1}, stats.FormatCounts)
}

func TestComputeStatsDoesNotMutateCallerCopy(t *testing.T) {
	t.Parallel()

	files := []File{{FileName: "a.mp3", SizeBytes: 10}}
	stats := ComputeStats(files)
	stats.FormatCounts[".wav"] = 99

	assert.Equal(t, map[string]int{".mp3": 1}, ComputeStats(files).FormatCounts)
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
