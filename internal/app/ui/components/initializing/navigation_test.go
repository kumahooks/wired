package initializing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoveCursor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mode       initMode
		start      int
		delta      int
		wantCursor int
	}{
		{
			name:       "config error advances from reload to proceed",
			mode:       modeConfigError,
			start:      0,
			delta:      1,
			wantCursor: 1,
		},
		{
			name:       "config error wraps from proceed to reload",
			mode:       modeConfigError,
			start:      1,
			delta:      1,
			wantCursor: 0,
		},
		{
			name:       "config error wraps from reload backwards to proceed",
			mode:       modeConfigError,
			start:      0,
			delta:      -1,
			wantCursor: 1,
		},
		{
			name:       "config error delta zero stays put",
			mode:       modeConfigError,
			start:      0,
			delta:      0,
			wantCursor: 0,
		},
		{
			name:       "config error large positive delta wraps",
			mode:       modeConfigError,
			start:      0,
			delta:      5,
			wantCursor: 1,
		},
		{
			name:       "config error large negative delta wraps",
			mode:       modeConfigError,
			start:      1,
			delta:      -4,
			wantCursor: 1,
		},
		{
			name:       "loading single visible button is a no-op",
			mode:       modeLoading,
			start:      1,
			delta:      1,
			wantCursor: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t))
			model.mode = test.mode
			model.cursorPosition = test.start
			model.moveCursor(test.delta)

			assert.Equal(t, test.wantCursor, model.cursorPosition)
		})
	}
}

func TestMoveCursorEmptyButtonsIsNoOp(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))
	model.buttons = nil
	model.cursorPosition = 0
	model.moveCursor(1)

	assert.Zero(t, model.cursorPosition)
}
