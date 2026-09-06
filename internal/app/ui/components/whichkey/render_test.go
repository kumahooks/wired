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

	model := newTestModel(t)

	rendered := model.Render(40, 10)
	lines := strings.Split(testutil.StripANSI(rendered), "\n")

	assert.Len(t, lines, model.cardHeight(2), "card must render padding, both action rows, gap, and hint")
	for index, line := range lines {
		assert.Len(t, []rune(line), 40, "card line %d must span the full window width", index)
	}
}

func TestRenderContainsCommandEntry(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)
	rendered := testutil.StripANSI(model.Render(40, 10))

	assert.Contains(t, rendered, "p -> playlist")
	assert.Contains(t, rendered, "l -> library stats")
}

func TestMappedActions(t *testing.T) {
	t.Parallel()

	bindings := testBindings(t)
	model := newTestModel(t)

	actions := model.mappedActions()

	require.Len(t, actions, len(bindings))

	for index, binding := range bindings {
		help := binding.Keys.Help()
		assert.Equal(t, model.renderEntry(help.Key, help.Desc), actions[index].text)
		assert.Equal(t, lipgloss.Width(actions[index].text), actions[index].width)
	}
}

func TestRenderSnapshot(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)

	testutil.AssertSnapshot(t, "render_default", model.Render(40, 10))
}
