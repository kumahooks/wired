package whichkey

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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

	assert.Contains(t, rendered, "L -> library stats")
}

func TestRenderSnapshot(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	testutil.AssertSnapshot(t, "render_default", model.Render(40, 10))
}
