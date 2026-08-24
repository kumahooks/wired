package initializing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"wired/internal/core/testutil"
)

func TestRenderDefaultModel(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	assertSnapshot(t, "render_default", model.Render(80, 24))
}

func TestRenderWithNormalAndErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("config loaded successfully", LogNormal)
	model.AppendLog("theme.surface must be a hex color", LogError)
	model.AppendLog("library scan starting", LogNormal)

	assertSnapshot(t, "render_logs_normal_and_error", model.Render(80, 24))
}

func TestRenderCountInProgress(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("counting total library files", LogNormal)
	model.SetCountFilesProgress(42)

	assertSnapshot(t, "render_count_in_progress", model.Render(80, 24))
}

func TestRenderCountDone(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("counting total library files", LogNormal)
	model.AppendLog("counted a total of 137 audio files successfully", LogNormal)
	model.SetCountFilesProgress(-1)

	assertSnapshot(t, "render_count_done", model.Render(80, 24))
}

func TestRenderAllErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("config parse failed", LogError)
	model.AppendLog("keybinds.select must have at least one binding", LogError)
	model.AppendLog("theme.track must be a #RRGGBB hex color", LogError)

	assertSnapshot(t, "render_all_errors", model.Render(80, 24))
}

func TestRenderCursorOnFirstButton(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.cursorPosition = 0

	assertSnapshot(t, "render_cursor_first", model.Render(80, 24))
}

func TestRenderCursorOnSecondButton(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.cursorPosition = 1

	assertSnapshot(t, "render_cursor_second", model.Render(80, 24))
}

func TestRenderContainsPanelHeader(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	rendered := testutil.StripANSI(model.Render(80, 24))
	assert.True(
		t,
		strings.Contains(rendered, "wire(d) is starting..."),
		"render output missing panel header substring:\n%s",
		rendered,
	)
}
