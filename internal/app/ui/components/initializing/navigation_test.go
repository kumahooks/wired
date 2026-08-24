package initializing

import "testing"

// makeNButtons returns a button row of length n, with placeholder labels and actions.
func makeNButtons(n int) []button {
	buttons := make([]button, n)
	for index := range buttons {
		buttons[index] = button{label: "btn", action: actionReload}
	}

	return buttons
}

func TestMoveCursor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		buttons    int
		start      int
		delta      int
		wantCursor int
	}{
		{
			name:       "delta plus one from zero advances",
			buttons:    3,
			start:      0,
			delta:      1,
			wantCursor: 1,
		},
		{
			name:       "delta minus one from zero wraps to last",
			buttons:    3,
			start:      0,
			delta:      -1,
			wantCursor: 2,
		},
		{
			name:       "delta zero stays put",
			buttons:    3,
			start:      1,
			delta:      0,
			wantCursor: 1,
		},
		{
			name:       "out-of-range positive delta normalizes via modulo",
			buttons:    3,
			start:      0,
			delta:      5,
			wantCursor: 2,
		},
		{
			name:       "cursor at last position wraps forward to zero",
			buttons:    3,
			start:      2,
			delta:      1,
			wantCursor: 0,
		},
		{
			name:       "large negative delta wraps backwards",
			buttons:    3,
			start:      1,
			delta:      -4,
			wantCursor: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := Model{
				buttons:        makeNButtons(test.buttons),
				cursorPosition: test.start,
			}

			model.moveCursor(test.delta)

			if model.cursorPosition != test.wantCursor {
				t.Fatalf(
					"cursorPosition = %d, want %d (start %d, delta %d, buttons %d)",
					model.cursorPosition, test.wantCursor, test.start, test.delta, test.buttons,
				)
			}
		})
	}
}

func TestMoveCursorEmptyButtonsIsNoOp(t *testing.T) {
	t.Parallel()

	model := Model{
		buttons:        nil,
		cursorPosition: 0,
	}

	model.moveCursor(1)

	if model.cursorPosition != 0 {
		t.Fatalf("cursorPosition = %d, want 0 on empty buttons", model.cursorPosition)
	}
}

