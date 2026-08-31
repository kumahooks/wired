package librarystats

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"wired/internal/core/audio"
	"wired/internal/core/testutil"
)

func TestRenderEmptyLibrary(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())

	testutil.AssertSnapshot(t, "render_empty_library", model.Render(80, 24))
}

func TestRenderManyPathsTruncatesAndCountsRemainder(t *testing.T) {
	t.Parallel()

	paths := make([]string, 0, maxVisiblePathRows+2)
	for index := range maxVisiblePathRows + 2 {
		paths = append(paths, strings.Repeat("/very/deep/library/path", 3)+"/dir"+string(rune('a'+index)))
	}

	model := New(testutil.DefaultKeyMap(t), nil)
	model.SetLibraryPaths(paths)

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(t, strings.Contains(rendered, "...and 2 more"), "render output missing remainder line:\n%s", rendered)
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
	assert.True(t, strings.Contains(rendered, "no library paths in config"), "render output:\n%s", rendered)
	assert.True(t, strings.Contains(rendered, dashPlaceholder), "render output:\n%s", rendered)
}

func TestApplyThemeAndSetLibraryPaths(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t), nil)
	model.ApplyTheme(testutil.DefaultTheme())
	model.SetLibraryPaths([]string{"/new/path"})

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(t, strings.Contains(rendered, "/new/path"), "render output:\n%s", rendered)
}
