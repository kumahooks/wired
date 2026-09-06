package audio

import (
	"cmp"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// topEntryCount is how many entries the leaderboard-style stats hold (top 3).
const topEntryCount = 3

// NamedCount is a "name + raw count" pair. the count is a plain number (file count, album count, or seconds of playtime).
type NamedCount struct {
	Name  string
	Value int
}

type Stats struct {
	FilesCount       int            // the total number of audio files.
	FormatCounts     map[string]int // a map of each file extension and its total amount.
	TotalBytes       int64          // the sum of the size of the whole library in bytes.
	BiggestFileBytes int64          // the size of the biggest file in the library.

	BytesPerFormat map[string]int64 // a map of the sum of file sizes for each file extension.

	TopArtistsByFiles    []NamedCount // holds the artists with most files.
	TopArtistsByAlbums   []NamedCount // holds the artists with most distinct albums.
	TopArtistsByDuration []NamedCount // holds the artists with most total playtime, in seconds.

	HasTrackLengths    bool       // whether the lengths of the tracks have been computed.
	LongestTrack       NamedCount // the longest track in seconds.
	ShortestTrack      NamedCount // the shortest track in seconds.
	AverageLengthTrack NamedCount // the average length among all the tracks, in seconds.

	MissingTitleCount  int // how many files are missing title metadata.
	MissingArtistCount int // how many files are missing artist metadata.
	MissingAlbumCount  int // how many files are missing album metadata.

	HasAlbumLengths    bool       // whether the lengths of the albums have been computed.
	LongestAlbum       NamedCount // the longest album in seconds.
	ShortestAlbum      NamedCount // the shortest album in seconds.
	AverageLengthAlbum NamedCount // the average length among all the albums, in seconds.

	DuplicatedTrackCount int // how many files are duplicated by filename, artist, album, and length.
}

// statsAccumulator is the per-file folding state ComputeStats aggregates over.
type statsAccumulator struct {
	artistFiles       map[string]int    // artist -> file count.
	artistAlbums      map[string]int    // artist -> distinct album count.
	artistSeconds     map[string]int    // artist -> total playtime in seconds.
	albumFiles        map[string]int    // "artist:album" -> file count.
	albumLengths      map[string]int    // "artist:album" -> total playtime in seconds.
	albumNames        map[string]string // "artist:album" -> album display name.
	trackKeys         map[string]int    // "filename:artist:album:length" -> duplicate detection.
	totalTrackSeconds int
	tracksWithLength  int
}

func newStatsAccumulator() *statsAccumulator {
	return &statsAccumulator{
		artistFiles:   make(map[string]int),
		artistAlbums:  make(map[string]int),
		artistSeconds: make(map[string]int),
		albumFiles:    make(map[string]int),
		albumLengths:  make(map[string]int),
		albumNames:    make(map[string]string),
		trackKeys:     make(map[string]int),
	}
}

// ComputeStats derives library-wide statistics from the files in the library.
func ComputeStats(library *Library) Stats {
	stats := Stats{
		FormatCounts:   make(map[string]int),
		BytesPerFormat: make(map[string]int64),
	}

	if library == nil {
		return stats
	}

	accumulator := newStatsAccumulator()
	for _, file := range library.File {
		if file == nil {
			continue
		}

		accumulateTrack(&stats, accumulator, file)
	}

	stats.DuplicatedTrackCount = countDuplicates(accumulator.trackKeys)

	computeAverageTrackLength(&stats, accumulator)
	computeAlbumLengths(&stats, accumulator)

	stats.TopArtistsByFiles = topValues(accumulator.artistFiles)
	stats.TopArtistsByAlbums = topValues(accumulator.artistAlbums)
	stats.TopArtistsByDuration = topValues(accumulator.artistSeconds)

	return stats
}

// accumulateTrack computes stats and accumulates data based on the given track file.
func accumulateTrack(stats *Stats, accumulator *statsAccumulator, file *AudioFile) {
	stats.FilesCount++
	stats.TotalBytes += file.SizeBytes

	format := strings.ToLower(filepath.Ext(file.Path))
	stats.FormatCounts[format]++
	stats.BytesPerFormat[format] += file.SizeBytes

	if file.SizeBytes > stats.BiggestFileBytes {
		stats.BiggestFileBytes = file.SizeBytes
	}

	// required metadata health: Title, Artist, Album.
	if file.Title == "" {
		stats.MissingTitleCount++
	}
	if file.Artist == "" {
		stats.MissingArtistCount++
	}
	if file.Album == "" {
		stats.MissingAlbumCount++
	}

	// duplicate detection participates even without tags.
	fileName := filepath.Base(file.Path)
	trackKey := fileName + ":" + file.Artist + ":" + file.Album + ":" + strconv.Itoa(file.Length)
	accumulator.trackKeys[trackKey]++

	if file.Length > 0 {
		// track lengths feed the average and the longest/shortest extremes.
		accumulator.totalTrackSeconds += file.Length
		accumulator.tracksWithLength++

		// tracks without a title tag display by their filename base instead.
		trackName := file.Title
		if trackName == "" {
			trackName = strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
		}

		if !stats.HasTrackLengths {
			stats.HasTrackLengths = true
			stats.LongestTrack = NamedCount{Name: trackName, Value: file.Length}
			stats.ShortestTrack = NamedCount{Name: trackName, Value: file.Length}
		} else {
			track := NamedCount{Name: trackName, Value: file.Length}

			if compareNamedCounts(track, stats.LongestTrack) > 0 {
				stats.LongestTrack = track
			}

			if compareNamedCounts(track, stats.ShortestTrack) < 0 {
				stats.ShortestTrack = track
			}
		}
	}

	if file.Artist == "" {
		return
	}

	accumulator.artistFiles[file.Artist]++
	accumulator.artistSeconds[file.Artist] += file.Length

	if file.Album == "" {
		return
	}

	// it's assumed an artist has only one album of the same title, so it's keyed by "{artist}:{album}".
	albumKey := file.Artist + ":" + file.Album
	if accumulator.albumFiles[albumKey] == 0 {
		accumulator.artistAlbums[file.Artist]++
		accumulator.albumNames[albumKey] = file.Album
	}
	accumulator.albumFiles[albumKey]++
	accumulator.albumLengths[albumKey] += file.Length
}

// countDuplicates counts every extra copy except the first of each duplicate key.
func countDuplicates(trackKeys map[string]int) int {
	duplicates := 0
	for _, count := range trackKeys {
		if count > 1 {
			duplicates += count - 1
		}
	}

	return duplicates
}

// computeAverageTrackLength computes the per-track average from the accumulated totals.
func computeAverageTrackLength(stats *Stats, acc *statsAccumulator) {
	if acc.tracksWithLength == 0 {
		return
	}

	stats.AverageLengthTrack = NamedCount{Value: acc.totalTrackSeconds / acc.tracksWithLength}
}

// compareNamedCounts orders NamedCounts by value ascending, ties broken by name ascending.
func compareNamedCounts(left NamedCount, right NamedCount) int {
	return cmp.Or(
		cmp.Compare(left.Value, right.Value),
		strings.Compare(left.Name, right.Name),
	)
}

// computeAlbumLengths derives longest/shortest/average album durations from the accumulated album playtimes.
func computeAlbumLengths(stats *Stats, acc *statsAccumulator) {
	albumLengths := make([]NamedCount, 0, len(acc.albumFiles))
	for albumKey := range acc.albumFiles {
		// an album without any length is useless data.
		durationSeconds := acc.albumLengths[albumKey]
		if durationSeconds <= 0 {
			continue
		}

		albumLengths = append(albumLengths, NamedCount{Name: acc.albumNames[albumKey], Value: durationSeconds})
	}
	if len(albumLengths) == 0 {
		return
	}

	slices.SortStableFunc(albumLengths, func(left NamedCount, right NamedCount) int {
		return cmp.Or(
			cmp.Compare(left.Value, right.Value),
			strings.Compare(left.Name, right.Name),
		)
	})

	stats.HasAlbumLengths = true
	stats.LongestAlbum = albumLengths[len(albumLengths)-1]
	stats.ShortestAlbum = albumLengths[0]

	albumTotal := 0
	for _, album := range albumLengths {
		albumTotal += album.Value
	}

	stats.AverageLengthAlbum = NamedCount{Value: albumTotal / len(albumLengths)}
}

// topValues returns the top topEntryCount entries with the highest value, tie-broken by name.
func topValues(counts map[string]int) []NamedCount {
	entries := make([]NamedCount, 0, len(counts))
	for name, value := range counts {
		entries = append(entries, NamedCount{Name: name, Value: value})
	}

	slices.SortStableFunc(entries, func(left NamedCount, right NamedCount) int {
		return cmp.Or(
			cmp.Compare(right.Value, left.Value),
			strings.Compare(left.Name, right.Name),
		)
	})

	if len(entries) > topEntryCount {
		entries = entries[:topEntryCount]
	}

	return entries
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

// GetReadableDuration renders an int representation of seconds as a readable duration (e.g. "50h 53m 26s").
func GetReadableDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	if seconds < 60*60 {
		return fmt.Sprintf("%dm %02ds", seconds/60, seconds%60)
	}

	return fmt.Sprintf("%dh %02dm %02ds", seconds/3600, (seconds%3600)/60, seconds%60)
}
