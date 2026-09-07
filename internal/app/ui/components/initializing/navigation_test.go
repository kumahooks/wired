package initializing

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"wired/internal/core/testutil"
)

func appendLogs(model *Model, count int) {
	for range count {
		model.AppendLog("line", LogNormal)
	}
}

func TestMoveLogScrollClampsToTailAndHead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		logCount    int
		startOffset int
		deltas      []int
		wantOffset  int
	}{
		{name: "fewer lines than viewport clamps at zero", logCount: 2, deltas: []int{-1}, wantOffset: 0},
		{name: "empty buffer clamps at zero", logCount: 0, deltas: []int{-1, 1}, wantOffset: 0},
		{
			name:        "scroll down below tail clamps at zero",
			logCount:    maxVisibleLogLines + 3,
			startOffset: 0,
			deltas:      []int{-1},
			wantOffset:  0,
		},
		{
			name:        "scroll up past head clamps at max offset",
			logCount:    maxVisibleLogLines + 3,
			startOffset: 3,
			deltas:      []int{2, 5},
			wantOffset:  3,
		},
		{
			name:        "scroll down from scrolled position returns to tail",
			logCount:    maxVisibleLogLines + 3,
			startOffset: 3,
			deltas:      []int{-3, -1},
			wantOffset:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(testutil.DefaultKeyMap(t))
			appendLogs(model, test.logCount)
			model.scrollOffset = test.startOffset

			for _, delta := range test.deltas {
				model.moveLogScroll(delta)
			}

			assert.Equal(t, test.wantOffset, model.scrollOffset)
		})
	}
}

func TestHandleMessageScrollKeysMoveViewport(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	appendLogs(model, maxVisibleLogLines+2)

	model.HandleMessage(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 1, model.scrollOffset, "scroll up should lift the viewport from the tail")

	model.HandleMessage(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 0, model.scrollOffset, "scroll down should return to the tail")
}

func TestMoveLogScrollJumps(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	appendLogs(model, maxVisibleLogLines+4)

	model.moveLogScrollToHead()
	assert.Equal(t, 4, model.scrollOffset, "head jump should offset by the whole buffer minus the viewport")

	model.moveLogScrollToTail()
	assert.Equal(t, 0, model.scrollOffset, "tail jump should return to zero offset")

	model.moveLogScrollToHead()
	assert.Equal(t, 4, model.scrollOffset, "head jump from any position should land at max offset")
}

func TestHandleMessageGgSequenceAndTailJump(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	appendLogs(model, maxVisibleLogLines+4)

	model.HandleMessage(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 0, model.scrollOffset, "G should jump to the tail")

	model.HandleMessage(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, model.scrollOffset, "single g should only arm the sequence")

	model.HandleMessage(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 1, model.scrollOffset)
	model.HandleMessage(tea.KeyPressMsg{Code: 'g', Text: "g"})
	model.HandleMessage(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 0, model.scrollOffset, "other keys between g presses should disarm the sequence")

	model.HandleMessage(tea.KeyPressMsg{Code: 'g', Text: "g"})
	model.HandleMessage(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 4, model.scrollOffset, "consecutive gg should jump to the head")
}

func TestVisibleLogRowsRespectsScrollOffset(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))
	appendLogs(model, maxVisibleLogLines+2)
	model.scrollOffset = 2

	rows := model.visibleLogRows()

	assert.Len(t, rows, maxVisibleLogLines)
	logTexts := model.LogLines()
	wantFirst := logTexts[0]
	assert.Contains(t, rows[0], wantFirst, "viewport should start at the oldest line when fully scrolled up")
}
