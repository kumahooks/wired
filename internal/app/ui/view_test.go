package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"wired/internal/core/testutil"
)

func TestViewContentTerminalTooSmallWidth(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowWidth = 10

	assert.Equal(t, "terminal size is too small", model.viewContent())
}

func TestViewContentTerminalTooSmallHeight(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowHeight = 10

	assert.Equal(t, "terminal size is too small", model.viewContent())
}

func TestViewContentTerminalTooSmallBoth(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowWidth = 10
	model.windowHeight = 10

	assert.Equal(t, "terminal size is too small", model.viewContent())
}

func TestViewContentBootstrappingRendersNothing(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	assert.Equal(t, uiBootstrapping, model.state)
	assert.Empty(t, model.viewContent())
}

func TestViewContentInitializing(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.setState(uiInitializing)
	assert.Equal(t, uiInitializing, model.state)

	rendered := model.viewContent()
	assert.True(
		t,
		strings.Contains(testutil.StripANSI(rendered), "wire(d) is starting..."),
		"viewContent() missing card header substring:\n%s",
		rendered,
	)
}

func TestViewContentPlaylist(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.state = uiPlaylist

	rendered := model.viewContent()
	assert.True(
		t,
		strings.Contains(testutil.StripANSI(rendered), "this is the playlist view"),
		"viewContent() missing playlist substring:\n%s",
		rendered,
	)

	model.state = uiLibraryStats

	rendered = model.viewContent()
	assert.True(
		t,
		strings.Contains(testutil.StripANSI(rendered), "library stats"),
		"viewContent() missing library stats substring:\n%s",
		rendered,
	)
}

func TestViewContentUnknownState(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.state = uiState(99)

	assert.Empty(t, model.viewContent())
}

func TestView(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.config.Title = "test-title"

	view := model.View()

	assert.True(t, view.AltScreen, "view.AltScreen = false, want true")
	assert.Equal(t, tea.MouseModeNone, view.MouseMode)
	assert.Equal(t, "test-title", view.WindowTitle)
	assert.Equal(t, model.viewContent(), view.Content)
}
