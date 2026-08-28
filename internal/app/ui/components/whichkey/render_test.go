package whichkey

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/testutil"
)

func TestRenderRectangleAtExpectedCardHeight(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	rendered := model.Render(40, 10)
	lines := strings.Split(testutil.StripANSI(rendered), "\n")

	assert.Len(t, lines, model.cardHeight(1), "card must render padding, one action row, gap, and hint")
	for index, line := range lines {
		assert.Len(t, []rune(line), 40, "card line %d must span the full window width", index)
	}
}

func TestRenderContainsCommandEntry(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	rendered := testutil.StripANSI(model.Render(40, 10))

	assert.Contains(t, rendered, "P -> playlist")
	assert.Contains(t, rendered, "L -> library stats")
}

func TestMappedActions(t *testing.T) {
	t.Parallel()

	keyMap := testutil.DefaultKeyMap(t)
	model := New(keyMap)

	actions := model.mappedActions()

	require.Len(t, actions, 2)

	playlistHelp := keyMap.Actions.Playlist.Help()
	assert.Equal(t, model.renderEntry(playlistHelp.Key, playlistHelp.Desc), actions[0].text)
	assert.Equal(t, lipgloss.Width(actions[0].text), actions[0].width)

	libraryStatsHelp := keyMap.Actions.LibraryStats.Help()
	assert.Equal(t, model.renderEntry(libraryStatsHelp.Key, libraryStatsHelp.Desc), actions[1].text)
	assert.Equal(t, lipgloss.Width(actions[1].text), actions[1].width)
}

func TestRenderSnapshot(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	testutil.AssertSnapshot(t, "render_default", model.Render(40, 10))
}
