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

func TestViewContentInitializing(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	assert.Equal(t, uiInitializing, model.state)

	rendered := model.viewContent()
	assert.True(
		t,
		strings.Contains(testutil.StripANSI(rendered), "wire(d) is starting..."),
		"viewContent() missing panel header substring:\n%s",
		rendered,
	)
}

func TestViewContentIdle(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.state = uiIdle

	rendered := model.viewContent()
	assert.True(
		t,
		strings.Contains(testutil.StripANSI(rendered), "program loaded successfully. idle~"),
		"viewContent() missing idle substring:\n%s",
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
