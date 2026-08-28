package notification

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/testutil"
)

func TestRenderSnapshot(t *testing.T) {
	t.Parallel()

	model := New()
	model.PushNotification("config reloaded")
	model.PushNotification("scan finished: 42 files")

	assertSnapshot(t, "render_default", model.Render(80, 24))
}

func TestRenderEmptyIsBlank(t *testing.T) {
	t.Parallel()

	model := New()

	assert.Empty(t, model.Render(80, 24))
}

func TestRenderClipsLongMessage(t *testing.T) {
	t.Parallel()

	model := New()
	model.PushNotification(
		"this is a very long notification message that will surely exceed the maximum card content lines and must be clipped!!!!!! right??????????",
	)

	windowWidth := 80
	rendered := testutil.StripANSI(model.Render(windowWidth, 24))
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")

	// A card has a top and bottom border line around the clipped body.
	assert.LessOrEqual(
		t,
		len(lines),
		maxContentLines+2,
		"card body should clip to maxContentLines, got:\n%s",
		rendered,
	)
}

func TestRenderCapsAtMaxRenderedNotifications(t *testing.T) {
	t.Parallel()

	model := New()
	for range 6 {
		model.PushNotification("message")
	}

	rendered := testutil.StripANSI(model.Render(80, 24))

	cardCount := strings.Count(rendered, "╭")
	assert.Equal(t, maxRenderedNotifications, cardCount, "rendered cards:\n%s", rendered)
}

func TestRenderRightAlignsStackedCards(t *testing.T) {
	t.Parallel()

	model := New()
	model.PushNotification("tiny")
	model.PushNotification("a considerably longer notification message for alignment")

	windowWidth := 80
	rendered := testutil.StripANSI(model.Render(windowWidth, 24))
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")

	// Right-aligned means no trailing padding: every non-empty line ends at the same column.
	require.Greater(t, len(lines), 0)

	rightEdge := len([]rune(lines[0]))
	for index, line := range lines {
		runes := []rune(line)

		assert.NotEmpty(t, strings.TrimRight(line, " "), "line %d is blank, unexpected padding", index)
		assert.Len(
			t,
			runes,
			rightEdge,
			"line %d ends at a different column, cards are not right-aligned:\n%s",
			index,
			rendered,
		)
	}
}
