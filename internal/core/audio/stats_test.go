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
		assert.Equal(t, int64(4096), stats.BiggestFileBytes)
		assert.Equal(t, map[string]int64{".mp3": 3072, ".flac": 4096}, stats.BytesPerFormat)
	})

	t.Run("library of zero-byte files has no heaviest file", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.Add("/music/song1.mp3", 0)

		stats := ComputeStats(library)

		assert.Equal(t, int64(0), stats.BiggestFileBytes)
	})

	t.Run("nil file entries are skipped", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/music/song1.mp3"] = nil

		stats := ComputeStats(library)

		assert.Equal(t, 0, stats.FilesCount)
	})
}

func TestComputeStatsDuplicates(t *testing.T) {
	t.Parallel()

	t.Run("identical filename artist album and length are duplicates", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01 song.flac"] = &AudioFile{
			Path:   "/a/01 song.flac",
			Title:  "song",
			Artist: "artist",
			Album:  "album",
			Length: 100,
		}
		library.File["/b/01 song.flac"] = &AudioFile{
			Path:   "/b/01 song.flac",
			Title:  "song",
			Artist: "artist",
			Album:  "album",
			Length: 100,
		}

		stats := ComputeStats(library)

		assert.Equal(t, 1, stats.DuplicatedTrackCount)
	})

	t.Run("same track under different filenames is not a duplicate", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01 song.flac"] = &AudioFile{
			Path:   "/a/01 song.flac",
			Title:  "song",
			Artist: "artist",
			Album:  "album",
			Length: 100,
		}
		library.File["/a/song.flac"] = &AudioFile{
			Path:   "/a/song.flac",
			Title:  "song",
			Artist: "artist",
			Album:  "album",
			Length: 100,
		}

		stats := ComputeStats(library)

		assert.Equal(t, 0, stats.DuplicatedTrackCount)
	})

	t.Run("distinct lengths are not duplicates", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01 song.flac"] = &AudioFile{
			Path:   "/a/01 song.flac",
			Artist: "artist",
			Album:  "album",
			Length: 100,
		}
		library.File["/b/01 song.flac"] = &AudioFile{
			Path:   "/b/01 song.flac",
			Artist: "artist",
			Album:  "album",
			Length: 200,
		}

		stats := ComputeStats(library)

		assert.Equal(t, 0, stats.DuplicatedTrackCount)
	})

	t.Run("files without artist participate in duplicate detection", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01 song.flac"] = &AudioFile{Path: "/a/01 song.flac", Length: 100}
		library.File["/b/01 song.flac"] = &AudioFile{Path: "/b/01 song.flac", Length: 100}

		stats := ComputeStats(library)

		assert.Equal(t, 1, stats.DuplicatedTrackCount)
	})

	t.Run("three copies count as two duplicates", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		for _, dir := range []string{"/a", "/b", "/c"} {
			path := dir + "/01 song.flac"
			library.File[path] = &AudioFile{Path: path, Artist: "artist", Album: "album", Length: 100}
		}

		stats := ComputeStats(library)

		assert.Equal(t, 2, stats.DuplicatedTrackCount)
	})
}

func TestComputeStatsTrackLengths(t *testing.T) {
	t.Parallel()

	t.Run("longest shortest and average with titles", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac", Title: "one", Length: 100}
		library.File["/a/02.flac"] = &AudioFile{Path: "/a/02.flac", Title: "two", Length: 300}
		library.File["/a/03.flac"] = &AudioFile{Path: "/a/03.flac", Title: "three", Length: 200}

		stats := ComputeStats(library)

		assert.True(t, stats.HasTrackLengths)
		assert.Equal(t, "two", stats.LongestTrack.Name)
		assert.Equal(t, 300, stats.LongestTrack.Value)
		assert.Equal(t, "one", stats.ShortestTrack.Name)
		assert.Equal(t, 100, stats.ShortestTrack.Value)
		assert.Equal(t, 200, stats.AverageLengthTrack.Value)
	})

	t.Run("missing title falls back to filename base", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01 fallback.flac"] = &AudioFile{Path: "/a/01 fallback.flac", Length: 100}

		stats := ComputeStats(library)

		assert.Equal(t, "01 fallback", stats.LongestTrack.Name)
	})

	t.Run("no length data", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.Add("/a/01.flac", 1024)

		stats := ComputeStats(library)

		assert.False(t, stats.HasTrackLengths)
		assert.Equal(t, "", stats.LongestTrack.Name)
		assert.Equal(t, 0, stats.AverageLengthTrack.Value)
	})

	t.Run("tied lengths break ties by name deterministically", func(t *testing.T) {
		t.Parallel()

		for range 20 {
			library := NewLibrary()
			library.File["/a/01 zeta.flac"] = &AudioFile{Path: "/a/01 zeta.flac", Title: "zeta", Length: 100}
			library.File["/a/02 alpha.flac"] = &AudioFile{Path: "/a/02 alpha.flac", Title: "alpha", Length: 100}

			stats := ComputeStats(library)

			assert.Equal(t, "zeta", stats.LongestTrack.Name)
			assert.Equal(t, "alpha", stats.ShortestTrack.Name)
		}
	})
}

func TestComputeStatsAlbumLengths(t *testing.T) {
	t.Parallel()

	t.Run("longest shortest and average across albums", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac", Artist: "artist a", Album: "album a", Length: 100}
		library.File["/a/02.flac"] = &AudioFile{Path: "/a/02.flac", Artist: "artist a", Album: "album a", Length: 200}
		library.File["/b/01.flac"] = &AudioFile{Path: "/b/01.flac", Artist: "artist b", Album: "album b", Length: 300}

		stats := ComputeStats(library)

		assert.True(t, stats.HasAlbumLengths)
		assert.Equal(t, "album b", stats.LongestAlbum.Name)
		assert.Equal(t, 300, stats.LongestAlbum.Value)
		assert.Equal(t, "album a", stats.ShortestAlbum.Name)
		assert.Equal(t, 300, stats.ShortestAlbum.Value)
		assert.Equal(t, 300, stats.AverageLengthAlbum.Value)
	})

	t.Run("album display name keeps everything after the first colon", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac", Artist: "a:b", Album: "c:d", Length: 100}

		stats := ComputeStats(library)

		assert.Equal(t, "c:d", stats.LongestAlbum.Name)
	})

	t.Run("albums without length data", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac", Artist: "artist", Album: "album", Length: 0}

		stats := ComputeStats(library)

		assert.False(t, stats.HasAlbumLengths)
	})

	t.Run("tied album lengths break ties by name deterministically", func(t *testing.T) {
		t.Parallel()

		for range 20 {
			library := NewLibrary()
			library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac", Artist: "artist a", Album: "zeta", Length: 100}
			library.File["/b/01.flac"] = &AudioFile{Path: "/b/01.flac", Artist: "artist b", Album: "alpha", Length: 100}

			stats := ComputeStats(library)

			assert.Equal(t, "zeta", stats.LongestAlbum.Name)
			assert.Equal(t, "alpha", stats.ShortestAlbum.Name)
		}
	})
}

func TestComputeStatsMetadataHealth(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	library.File["/a/01.flac"] = &AudioFile{Path: "/a/01.flac"}
	library.File["/a/02.flac"] = &AudioFile{Path: "/a/02.flac", Title: "two", Album: "album"}

	stats := ComputeStats(library)

	assert.Equal(t, 1, stats.MissingTitleCount)
	assert.Equal(t, 2, stats.MissingArtistCount)
	assert.Equal(t, 1, stats.MissingAlbumCount)
}

func TestComputeStatsTopArtists(t *testing.T) {
	t.Parallel()

	t.Run("top three by files albums and playtime", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/a/1.flac"] = &AudioFile{Path: "/a/1.flac", Artist: "artist a", Album: "album a1", Length: 10}
		library.File["/a/2.flac"] = &AudioFile{Path: "/a/2.flac", Artist: "artist a", Album: "album a1", Length: 20}
		library.File["/a/3.flac"] = &AudioFile{Path: "/a/3.flac", Artist: "artist a", Album: "album a2", Length: 30}
		library.File["/b/1.flac"] = &AudioFile{Path: "/b/1.flac", Artist: "artist b", Album: "album b", Length: 400}
		library.File["/c/1.flac"] = &AudioFile{Path: "/c/1.flac", Artist: "artist c", Album: "album c", Length: 500}
		library.File["/d/1.flac"] = &AudioFile{Path: "/d/1.flac", Artist: "artist d", Album: "album d", Length: 600}

		stats := ComputeStats(library)

		assert.Equal(t, []NamedCount{
			{Name: "artist a", Value: 3},
			{Name: "artist b", Value: 1},
			{Name: "artist c", Value: 1},
		}, stats.TopArtistsByFiles)

		assert.Equal(t, []NamedCount{
			{Name: "artist a", Value: 2},
			{Name: "artist b", Value: 1},
			{Name: "artist c", Value: 1},
		}, stats.TopArtistsByAlbums)

		assert.Equal(t, []NamedCount{
			{Name: "artist d", Value: 600},
			{Name: "artist c", Value: 500},
			{Name: "artist b", Value: 400},
		}, stats.TopArtistsByDuration)
	})

	t.Run("value ties break ties by name", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.File["/b/1.flac"] = &AudioFile{Path: "/b/1.flac", Artist: "artist b", Length: 10}
		library.File["/a/1.flac"] = &AudioFile{Path: "/a/1.flac", Artist: "artist a", Length: 10}

		stats := ComputeStats(library)

		assert.Equal(
			t,
			[]NamedCount{{Name: "artist a", Value: 10}, {Name: "artist b", Value: 10}},
			stats.TopArtistsByDuration,
		)
	})

	t.Run("no artists", func(t *testing.T) {
		t.Parallel()

		library := NewLibrary()
		library.Add("/a/1.flac", 1024)

		stats := ComputeStats(library)

		assert.Empty(t, stats.TopArtistsByFiles)
		assert.Empty(t, stats.TopArtistsByDuration)
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

func TestReadableDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{"zero", 0, "0s"},
		{"negative", -5, "0s"},
		{"seconds boundary", 59, "59s"},
		{"minute", 60, "1m 00s"},
		{"minute boundary", 3599, "59m 59s"},
		{"hour", 3600, "1h 00m 00s"},
		{"hours", 10006, "2h 46m 46s"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, GetReadableDuration(test.seconds))
		})
	}
}
