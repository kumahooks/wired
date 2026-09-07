package librarystats

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"wired/internal/core/audio"
	"wired/internal/core/testutil"
)

func TestRenderEmptyLibrary(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	testutil.AssertSnapshot(t, "render_empty_library", model.Render(80, 24))
}

func TestRenderEmptyLibraryWideWindow(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	testutil.AssertSnapshot(t, "render_empty_library_wide_window", model.Render(120, 24))
}

func TestRenderEmptyLibraryWithLeaderboards(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	testutil.AssertSnapshot(t, "render_empty_library_with_leaderboards", model.Render(120, 40))
}

func TestRenderCompactGridDropsArtistsAndLengths(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	rendered := testutil.StripANSI(model.Render(80, 40))
	assert.Contains(t, rendered, "LIBRARY SIZE", "render output:\n%s", rendered)
	assert.Contains(t, rendered, "METADATA HEALTH", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "TOP ARTISTS", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "PLACEHOLDER", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "TRACK LENGTHS", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "ALBUM LENGTHS", "render output:\n%s", rendered)
}

func TestRenderWideShortWindowKeepsArtistsDropsLengths(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	rendered := testutil.StripANSI(model.Render(120, 24))
	assert.Contains(t, rendered, "TOP ARTISTS", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "TRACK LENGTHS", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "ALBUM LENGTHS", "render output:\n%s", rendered)
}

func TestRenderGridFitsWindowHeight(t *testing.T) {
	t.Parallel()

	for _, windowSize := range []struct {
		width  int
		height int
	}{
		{120, fullGridLines},       // exact full-grid fit.
		{120, fullGridLines - 1},   // lengths row must drop, grid must still fit.
		{80, compactGridLines},     // exact compact-grid fit.
		{80, compactGridLines - 1}, // must degrade to the small-window message.
		{40, 10},                   // tiny window, message path.
	} {
		model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

		rendered := model.Render(windowSize.width, windowSize.height)
		assert.True(
			t,
			lipgloss.Height(rendered) <= windowSize.height && lipgloss.Width(rendered) <= windowSize.width,
			"render at %dx%d overflows the window:\n%s",
			windowSize.width,
			windowSize.height,
			testutil.StripANSI(rendered),
		)
	}
}

func TestRenderBoundaryWidthsSwitchTiers(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	// at exactly fullGridWidth the artists column fits; one column narrower it drops.
	withArtists := testutil.StripANSI(model.Render(fullGridWidth, 40))
	assert.Contains(t, withArtists, "TOP ARTISTS", "render output:\n%s", withArtists)

	withoutArtists := testutil.StripANSI(model.Render(fullGridWidth-1, 40))
	assert.NotContains(t, withoutArtists, "TOP ARTISTS", "render output:\n%s", withoutArtists)

	// at exactly lengthsRowWidth (== fullGridWidth) the lengths row fits; one narrower it drops.
	withLengths := testutil.StripANSI(model.Render(lengthsRowWidth, 40))
	assert.Contains(t, withLengths, "TRACK LENGTHS", "render output:\n%s", withLengths)

	withoutLengths := testutil.StripANSI(model.Render(lengthsRowWidth-1, 40))
	assert.NotContains(t, withoutLengths, "TRACK LENGTHS", "render output:\n%s", withoutLengths)
}

func TestRenderBoundaryHeightsSwitchTiers(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	// at exactly fullGridLines the lengths row fits, one shorter it drops (artists still fit).
	withLengths := testutil.StripANSI(model.Render(120, fullGridLines))
	assert.Contains(t, withLengths, "TRACK LENGTHS", "render output:\n%s", withLengths)

	withoutLengths := testutil.StripANSI(model.Render(120, fullGridLines-1))
	assert.NotContains(t, withoutLengths, "TRACK LENGTHS", "render output:\n%s", withoutLengths)

	// at exactly compactGridLines the compact grid fits, one shorter it degrades to the message.
	compact := testutil.StripANSI(model.Render(80, compactGridLines))
	assert.Contains(t, compact, "LIBRARY SIZE", "render output:\n%s", compact)

	degenerate := testutil.StripANSI(model.Render(80, compactGridLines-1))
	assert.Contains(t, degenerate, smallWindowText, "render output:\n%s", degenerate)
}

func TestRenderTinyWindowShowsSizeMessage(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	rendered := testutil.StripANSI(model.Render(40, 10))
	assert.Contains(t, rendered, "the window is too small", "render output:\n%s", rendered)
	assert.Contains(t, rendered, "library stats.", "render output:\n%s", rendered)
	assert.NotContains(t, rendered, "LIBRARY SIZE", "render output:\n%s", rendered)

	testutil.AssertSnapshot(t, "render_empty_library_tiny_window", model.Render(40, 10))
}

func TestRenderPopulatedLibraryWithLeaderboards(t *testing.T) {
	t.Parallel()

	library := audio.NewLibrary()
	files := map[string]*audio.AudioFile{
		"/music/artist_a/album_x/01 long album name that goes on and on forever.flac": {
			Path:      "/music/artist_a/album_x/01 long album name that goes on and on forever.flac",
			Title:     "one",
			Artist:    "artist a",
			Album:     "album x",
			Year:      "2001",
			Length:    61,
			SizeBytes: 1024,
		},
		"/music/artist_a/album_x/02 two.flac": {
			Path:      "/music/artist_a/album_x/02 two.flac",
			Title:     "two",
			Artist:    "artist a",
			Album:     "album x",
			Year:      "2001",
			Length:    122,
			SizeBytes: 2048,
		},
		"/music/artist_b/album_y/01 three.flac": {
			Path:      "/music/artist_b/album_y/01 three.flac",
			Title:     "three",
			Artist:    "artist b",
			Album:     "album y",
			Year:      "2002",
			Length:    183,
			SizeBytes: 4096,
		},
		"/music/artist_b/album_y/02 four.flac": {
			Path:      "/music/artist_b/album_y/02 four.flac",
			Title:     "four",
			Artist:    "artist b",
			Album:     "album y",
			Year:      "2002",
			Length:    244,
			SizeBytes: 8192,
		},
		"/music/artist_c/album_z/01 five.flac": {
			Path:      "/music/artist_c/album_z/01 five.flac",
			Title:     "five",
			Artist:    "artist c",
			Album:     "album z",
			Year:      "2003",
			Length:    305,
			SizeBytes: 16384,
		},
	}
	for path, file := range files {
		library.File[path] = file
	}

	model := New(testutil.DefaultKeyMap(t), library)
	model.ComputeStats()
	model.SetLibraryPaths([]string{"/music"})

	testutil.AssertSnapshot(t, "render_populated_library_with_leaderboards", model.Render(120, 40))
}

func TestRenderCJKArtistNamesDoNotWrap(t *testing.T) {
	t.Parallel()

	library := audio.NewLibrary()
	paths := []string{
		"/music/天気予報/01 track one.flac",
		"/music/天気予報/02 track two.flac",
		"/music/air/album/01 track three.flac",
	}
	for index, path := range paths {
		artist := "天気予報"
		if index == 2 {
			artist = "air"
		}

		library.File[path] = &audio.AudioFile{
			Path:      path,
			Title:     fmt.Sprintf("track %d", index+1),
			Artist:    artist,
			Album:     "album",
			Length:    60 * (index + 1),
			SizeBytes: 1024,
		}
	}

	model := New(testutil.DefaultKeyMap(t), library)
	model.ComputeStats()

	rendered := testutil.StripANSI(model.Render(120, 60))
	lines := strings.Split(rendered, "\n")

	for _, line := range lines {
		if !strings.Contains(line, "TOP ARTISTS") && !strings.Contains(line, "artist") {
			continue
		}

		assert.True(
			t,
			strings.HasSuffix(strings.TrimRight(line, " "), "│"),
			"leaderboard row overflowed its card: %q", line,
		)
	}
	assert.Contains(t, rendered, "天気予報", "render output:\n%s", rendered)
}

func TestRenderManyPathsTruncatesAndCountsRemainder(t *testing.T) {
	t.Parallel()

	paths := make([]string, 0, 12)
	for index := range 12 {
		paths = append(paths, strings.Repeat("/very/deep/library/path", 3)+"/dir"+string(rune('a'+index)))
	}

	model := New(testutil.DefaultKeyMap(t), nil)
	model.SetLibraryPaths(paths)

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(
		t,
		strings.Contains(rendered, "...and 4 more paths"),
		"render output missing remainder line:\n%s",
		rendered,
	)
}

func TestRenderTruncatesLongPaths(t *testing.T) {
	t.Parallel()

	longPath := strings.Repeat("/verylongdirectoryname", 10)
	model := New(testutil.DefaultKeyMap(t), nil)
	model.SetLibraryPaths([]string{longPath})

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.False(t, strings.Contains(rendered, longPath), "render output should truncate long paths:\n%s", rendered)
}

func TestRenderTruncatesMultiBytePathsWithoutCorruption(t *testing.T) {
	t.Parallel()

	longPath := strings.Repeat("あいうえおかきくけこ", 10)
	model := New(testutil.DefaultKeyMap(t), nil)
	model.SetLibraryPaths([]string{longPath})

	rendered := model.Render(80, 24)
	assert.False(t, strings.Contains(rendered, longPath), "render output should truncate long paths")

	assert.True(t, utf8.ValidString(rendered), "render output contains invalid UTF-8 after truncation")
}

func TestRenderNilLibraryPointer(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), nil)

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(t, strings.Contains(rendered, "no library paths"), "render output:\n%s", rendered)
	assert.True(t, strings.Contains(rendered, emptyPlaceholder), "render output:\n%s", rendered)
}

func TestApplyThemeAndSetLibraryPaths(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), nil)
	model.ApplyTheme(testutil.DefaultTheme())
	model.SetLibraryPaths([]string{"/new/path"})

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(t, strings.Contains(rendered, "/new/path"), "render output:\n%s", rendered)
}

func TestRenderHidesScanStatusWhenNotDiscovering(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.False(t, strings.Contains(rendered, "audio files"), "render output:\n%s", rendered)
}

func TestRenderShowsDiscoveryProgressWhileDiscovering(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
	model.StartDiscovery()
	model.SetProgress(testutil.NewDiscoveryProgress(42, 0, false))

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(
		t,
		strings.Contains(rendered, "found 42 audio files..."),
		"render output:\n%s",
		rendered,
	)
}

func TestRenderHidesScanStatusAfterFinish(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
	model.StartDiscovery()
	model.SetProgress(testutil.NewDiscoveryProgress(7, 0, false))
	model.FinishDiscovery()

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.False(t, strings.Contains(rendered, "audio files"), "render output:\n%s", rendered)
}

func TestRenderMetatagPhaseLabels(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
	model.StartDiscovery()
	model.SetProgress(testutil.NewDiscoveryProgress(12, 5, true))

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.Contains(t, rendered, "12 audio files have been found", "render output:\n%s", rendered)
	assert.Contains(t, rendered, "parsing 5/12 files", "render output:\n%s", rendered)
	assert.False(
		t,
		strings.Contains(rendered, "at library paths"),
		"discovery label should not render during metatag phase:\n%s",
		rendered,
	)
}

func TestFinishDiscoveryClearsMetatagState(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
	model.StartDiscovery()
	model.SetProgress(testutil.NewDiscoveryProgress(4, 4, true))
	model.FinishDiscovery()
	model.StartDiscovery()

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.Contains(
		t,
		rendered,
		"found 0 audio files...",
		"fresh discovery should show the discovery label at zero:\n%s",
		rendered,
	)
}
