package ui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var ansiEscSeqUI = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

func stripANSIUI(value string) string {
	return ansiEscSeqUI.ReplaceAllString(value, "")
}

func TestViewContentTerminalTooSmallWidth(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowWidth = 10

	if got := model.viewContent(); got != "terminal size is too small" {
		t.Errorf("viewContent() = %q, want the too-small message", got)
	}
}

func TestViewContentTerminalTooSmallHeight(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowHeight = 10

	if got := model.viewContent(); got != "terminal size is too small" {
		t.Errorf("viewContent() = %q, want the too-small message", got)
	}
}

func TestViewContentTerminalTooSmallBoth(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.windowWidth = 10
	model.windowHeight = 10

	if got := model.viewContent(); got != "terminal size is too small" {
		t.Errorf("viewContent() = %q, want the too-small message", got)
	}
}

func TestViewContentInitializing(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	if model.state != uiInitializing {
		t.Fatalf("state = %v, want uiInitializing as the newTestUI default", model.state)
	}

	rendered := model.viewContent()
	if !strings.Contains(stripANSIUI(rendered), "wire(d) is starting...") {
		t.Errorf("viewContent() missing panel header substring:\n%s", rendered)
	}
}

func TestViewContentIdle(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.state = uiIdle

	rendered := model.viewContent()
	if !strings.Contains(stripANSIUI(rendered), "program loaded successfully. idle~") {
		t.Errorf("viewContent() missing idle substring:\n%s", rendered)
	}
}

func TestViewContentUnknownState(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.state = uiState(99)

	if got := model.viewContent(); got != "" {
		t.Errorf("viewContent() = %q, want empty string for unknown state", got)
	}
}

func TestView(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.config.Title = "test-title"

	view := model.View()

	if !view.AltScreen {
		t.Error("view.AltScreen = false, want true")
	}
	if view.MouseMode != tea.MouseModeNone {
		t.Errorf("view.MouseMode = %v, want tea.MouseModeNone", view.MouseMode)
	}
	if view.WindowTitle != "test-title" {
		t.Errorf("view.WindowTitle = %q, want %q", view.WindowTitle, "test-title")
	}
	if view.Content != model.viewContent() {
		t.Errorf("view.Content does not match viewContent()")
	}
}
