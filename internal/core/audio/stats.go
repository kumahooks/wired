package audio

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Stats struct {
	FilesCount   int            // the total number of audio files.
	FormatCounts map[string]int // a map of each file extension and its total amount.
	TotalBytes   int64          // the sum of the size of the whole library in bytes.
}

// ComputeStats derives library-wide statistics from the files in the library.
func ComputeStats(library *Library) Stats {
	stats := Stats{FormatCounts: make(map[string]int)}

	if library == nil {
		return stats
	}

	for _, file := range library.File {
		stats.FilesCount++
		stats.TotalBytes += file.SizeBytes

		format := strings.ToLower(filepath.Ext(file.Path))
		stats.FormatCounts[format]++
	}

	return stats
}

var byteSizeUnits = []struct {
	name  string
	bytes int64
}{
	{"GiB", 1 << 30},
	{"MiB", 1 << 20},
	{"KiB", 1 << 10},
}

// GetReadableByteSize renders a byte count as a readable string (e.g. "3.4 MiB").
func GetReadableByteSize(bytes int64) string {
	if bytes < 0 {
		bytes = 0
	}

	for _, unit := range byteSizeUnits {
		if bytes >= unit.bytes {
			return fmt.Sprintf("%.1f %s", float64(bytes)/float64(unit.bytes), unit.name)
		}
	}

	return fmt.Sprintf("%d B", bytes)
}
