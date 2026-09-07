package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"wired/internal/core/testutil"
)

func TestRenderContainsQuestionAndButtons(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)

	rendered := testutil.StripANSI(model.Render())

	assert.Contains(t, rendered, "proceed?")
	assert.Contains(t, rendered, "yes")
	assert.Contains(t, rendered, "no")
}

func TestRenderButtonStatesDiffer(t *testing.T) {
	t.Parallel()

	// Snapshot goldens are ANSI-stripped, so the focused/blurred styling difference is invisible to them.
	focused := openedTestModel(t).renderButton("scan", true)
	blurred := openedTestModel(t).renderButton("scan", false)

	assert.NotEqual(t, focused, blurred, "focused and blurred buttons must render differently")
}

func TestRenderWrapsLongQuestions(t *testing.T) {
	t.Parallel()

	// A question longer than the card's fixed width must wrap into more lines, not overflow the card horizontally.
	shortCard := openedTestModel(t).Render()

	longQuestionModel := openedTestModel(t)
	longQuestionModel.text = strings.Repeat("word ", 40)
	longCard := longQuestionModel.Render()

	assert.Greater(t, lipgloss.Height(longCard), lipgloss.Height(shortCard), "a long question must wrap, not overflow")
	assert.Equal(t, lipgloss.Width(shortCard), lipgloss.Width(longCard), "the card's width must stay fixed")
}
