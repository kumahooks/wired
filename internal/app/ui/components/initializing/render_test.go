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
	model.SetConfigError()

	assertSnapshot(t, "render_default", model.Render(80, 24))
}

func TestRenderWithNormalAndErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("config loaded successfully", LogNormal)
	model.AppendLog("theme.surface must be a hex color", LogError)
	model.AppendLog("library scan starting", LogNormal)
	model.SetConfigError()

	assertSnapshot(t, "render_logs_normal_and_error", model.Render(80, 24))
}

func TestRenderAllErrorLogs(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	model.AppendLog("config parse failed", LogError)
	model.AppendLog("keybinds.select must have at least one binding", LogError)
	model.AppendLog("theme.track must be a #RRGGBB hex color", LogError)
	model.SetConfigError()

	assertSnapshot(t, "render_all_errors", model.Render(80, 24))
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
