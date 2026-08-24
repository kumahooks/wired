package initializing

import (
	"strings"
	"testing"

	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

func defaultKeyMapForRender(t *testing.T) keymap.KeyMap {
	t.Helper()

	keyMap, err := keymap.New(config.Defaults().Keybinds)
	if err != nil {
		t.Fatalf("keymap.New(defaults) error: %v", err)
	}

	return keyMap
}

func TestRenderDefaultModel(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))

	assertSnapshot(t, "render_default", model.Render(80, 24))
}

func TestRenderWithNormalAndErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.AppendLog("config loaded successfully", LogNormal)
	model.AppendLog("theme.surface must be a hex color", LogError)
	model.AppendLog("library scan starting", LogNormal)

	assertSnapshot(t, "render_logs_normal_and_error", model.Render(80, 24))
}

func TestRenderCountInProgress(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.AppendLog("counting total library files", LogNormal)
	model.SetCountFilesProgress(42)

	assertSnapshot(t, "render_count_in_progress", model.Render(80, 24))
}

func TestRenderCountDone(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.AppendLog("counting total library files", LogNormal)
	model.AppendLog("counted a total of 137 audio files successfully", LogNormal)
	model.SetCountFilesProgress(-1)

	assertSnapshot(t, "render_count_done", model.Render(80, 24))
}

func TestRenderAllErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.AppendLog("config parse failed", LogError)
	model.AppendLog("keybinds.select must have at least one binding", LogError)
	model.AppendLog("theme.track must be a #RRGGBB hex color", LogError)

	assertSnapshot(t, "render_all_errors", model.Render(80, 24))
}

func TestRenderCursorOnFirstButton(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.cursorPosition = 0

	assertSnapshot(t, "render_cursor_first", model.Render(80, 24))
}

func TestRenderCursorOnSecondButton(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))
	model.cursorPosition = 1

	assertSnapshot(t, "render_cursor_second", model.Render(80, 24))
}

func TestRenderContainsPanelHeader(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMapForRender(t))

	rendered := stripANSI(model.Render(80, 24))
	if !strings.Contains(rendered, "wire(d) is starting...") {
		t.Errorf("render output missing panel header substring:\n%s", rendered)
	}
}
