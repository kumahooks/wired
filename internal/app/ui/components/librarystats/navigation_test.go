package librarystats

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/core/audio"
	"wired/internal/core/testutil"
)

func TestHandleMessageMovesCursorAndWraps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		message        tea.KeyPressMsg
		setupCursor    int
		wantCursor     int
		wantCursorSet  bool
		clearButtons   bool
		wantButtonHits int
	}{
		{
			name:          "MoveLeft wraps from first button to last",
			message:       tea.KeyPressMsg{Code: 'h', Text: "h"},
			setupCursor:   0,
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight advances from first button to last",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupCursor:   0,
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight wraps from last button to first",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupCursor:   1,
			wantCursor:    0,
			wantCursorSet: true,
		},
		{
			name:          "unmapped keys leave the cursor untouched",
			message:       tea.KeyPressMsg{Code: 'j', Text: "j"},
			setupCursor:   1,
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:         "Select with empty buttons is a no-op",
			message:      tea.KeyPressMsg{Code: tea.KeyEnter},
			clearButtons: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
			if test.clearButtons {
				model.buttons = nil
			}

			model.cursorPosition = test.setupCursor

			model.HandleMessage(test.message)

			if test.wantCursorSet {
				assert.Equal(t, test.wantCursor, model.cursorPosition)
			}
		})
	}
}

func TestHandleMessageSelectReturnsDialogActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupCursor     int
		wantOpenDialog  action.OpenConfirmDialogAction
		wantDialogOpen  bool
		wantCursorAtEnd int
	}{
		{
			name:        "Select on scan full opens confirm dialog for a full discovery",
			setupCursor: 0,
			wantOpenDialog: action.OpenConfirmDialogAction{
				Text:          scanFullDialogText,
				ConfirmLabel:  scanDialogConfirmLabel,
				CancelLabel:   scanDialogCancelLabel,
				ConfirmAction: action.DiscoverLibraryFullAction{},
			},
			wantDialogOpen: true,
		},
		{
			name:        "Select on scan new opens confirm dialog for a new discovery",
			setupCursor: 1,
			wantOpenDialog: action.OpenConfirmDialogAction{
				Text:          scanNewDialogText,
				ConfirmLabel:  scanDialogConfirmLabel,
				CancelLabel:   scanDialogCancelLabel,
				ConfirmAction: action.DiscoverLibraryNewAction{},
			},
			wantDialogOpen: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(testutil.DefaultKeyMap(t), audio.NewLibrary())
			model.cursorPosition = test.setupCursor

			gotAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

			require.IsType(t, action.OpenConfirmDialogAction{}, gotAction)
			assert.Equal(t, test.wantOpenDialog, gotAction)
		})
	}
}
