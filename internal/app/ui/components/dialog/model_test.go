package dialog

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/core/testutil"
)

// newTestModel builds a Model with the default keymap and style applied.
func newTestModel(t *testing.T) *Model {
	t.Helper()

	model := New()
	model.ApplyKeyMap(testutil.DefaultKeyMap(t))

	return model
}

// openedTestModel builds a Model already showing a question whose confirm action is OpenPlaylistAction.
func openedTestModel(t *testing.T) *Model {
	t.Helper()

	model := newTestModel(t)
	model.Open("proceed?", "yes", "no", action.OpenPlaylistAction{})

	return model
}

// afterGrace moves the model's grace anchor back past the key grace window.
func afterGrace(model *Model) {
	model.lastKeyAt = time.Now().Add(-(KeyGraceQuietPeriod + time.Millisecond))
}

// withinGrace moves the model's grace anchor forward.
func withinGrace(model *Model) {
	model.lastKeyAt = time.Now().Add(KeyGraceQuietPeriod)
}

func TestNew(t *testing.T) {
	t.Parallel()

	model := New()

	assert.False(t, model.IsOpen(), "a new model must start closed")
	assert.Empty(t, model.text, "a new model must start with no text")
	assert.Nil(t, model.confirmAction, "a new model must start with no confirm action")
	assert.Equal(t, newStyle(testutil.DefaultTheme()), model.style)
}

func TestOpenStoresQuestionAndDefaultsToCancel(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)
	storedAction := action.OpenPlaylistAction{}

	model.Open("proceed?", "yes", "no", storedAction)

	assert.True(t, model.IsOpen(), "Open must show the dialog")
	assert.Equal(t, "proceed?", model.text)
	assert.Equal(t, "yes", model.confirmLabel)
	assert.Equal(t, "no", model.cancelLabel)
	assert.Equal(t, storedAction, model.confirmAction)
	assert.Equal(t, int(cancelButton), model.cursorPosition, "cursor must start on the cancel button")
}

func TestCloseDropsQuestionAndAction(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)

	model.Close()

	assert.False(t, model.IsOpen(), "Close must hide the dialog")
	assert.Empty(t, model.text, "Close must drop the stored text")
	assert.Nil(t, model.confirmAction, "Close must drop the stored action")
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New()

	model.ApplyTheme(testutil.CustomTheme())

	assert.Equal(t, newStyle(testutil.CustomTheme()), model.style)
}

func TestApplyKeyMapStoresKeyMap(t *testing.T) {
	t.Parallel()

	model := New()
	keyMap := testutil.DefaultKeyMap(t)

	model.ApplyKeyMap(keyMap)

	assert.Equal(t, keyMap, model.keyMap)
}

func TestHandleMessageConfirmReturnsStoredActionAndCloses(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	afterGrace(model)

	// The cursor starts on cancel: move left once to reach confirm, then select.
	model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})
	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, action.OpenPlaylistAction{}, returnedAction, "confirm must dispatch the stored action")
	assert.False(t, model.IsOpen(), "confirm must close the dialog")
	assert.Nil(t, model.confirmAction, "confirm must drop the stored action")
}

func TestHandleMessageCancelReturnsNoAction(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	afterGrace(model)

	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, action.NoAction{}, returnedAction, "cancel must not dispatch the stored action")
	assert.False(t, model.IsOpen(), "cancel must close the dialog")
}

func TestHandleMessageGoBackClosesWithoutDispatching(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	afterGrace(model)

	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Equal(t, action.NoAction{}, returnedAction)
	assert.False(t, model.IsOpen(), "go back must close the dialog")
}

func TestHandleMessageMovesCursorBetweenButtons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		keyPress     tea.KeyPressMsg
		wantPosition int
	}{
		{
			name:         "left moves from cancel to confirm",
			keyPress:     tea.KeyPressMsg{Code: 'h', Text: "h"},
			wantPosition: int(confirmButton),
		},
		{
			name:         "right moves from cancel to confirm",
			keyPress:     tea.KeyPressMsg{Code: 'l', Text: "l"},
			wantPosition: int(confirmButton),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := openedTestModel(t)
			afterGrace(model)

			model.HandleMessage(test.keyPress)

			assert.Equal(t, test.wantPosition, model.cursorPosition)
		})
	}
}

func TestHandleMessageCursorWrapsAround(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	afterGrace(model)

	// Two moves in the same direction must wrap back to the starting button.
	model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})
	model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})

	assert.Equal(t, int(cancelButton), model.cursorPosition, "cursor must wrap back to the cancel button")
}

func TestHandleMessageInGraceWindowDropsKeys(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	withinGrace(model)

	// Still inside the grace window: nothing moves, nothing dispatches, nothing closes.
	model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})
	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, action.NoAction{}, returnedAction, "keys inside the grace window must be dropped")
	assert.True(t, model.IsOpen(), "keys inside the grace window must not close the dialog")
	assert.Equal(t, int(cancelButton), model.cursorPosition, "keys inside the grace window must not move the cursor")
}

func TestHandleMessageAbsorbedKeyExtendsGraceWindow(t *testing.T) {
	t.Parallel()

	// Each absorbed keystroke pushes the quiet-period window back.
	model := openedTestModel(t)
	afterGrace(model)
	model.lastKeyAt = time.Now().Add(-(KeyGraceQuietPeriod - time.Millisecond))

	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})

	assert.Equal(t, action.NoAction{}, returnedAction, "a key inside the quiet period must be dropped")
	assert.True(t, model.IsOpen(), "a key inside the quiet period must not close the dialog")
}

func TestHandleMessageNonKeyPressIsIgnored(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	afterGrace(model)

	returnedAction := model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 24})

	assert.Equal(t, action.NoAction{}, returnedAction)
	assert.True(t, model.IsOpen(), "a non key message must not close the dialog")
}

func TestHandleMessageClosedDialogIsIgnored(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)

	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, action.NoAction{}, returnedAction)
}

func TestHandleMessageConfirmWithoutActionReturnsNoAction(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)
	model.Open("proceed?", "yes", "no", nil)
	afterGrace(model)

	model.HandleMessage(tea.KeyPressMsg{Code: 'h', Text: "h"})
	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, action.NoAction{}, returnedAction, "confirming a nil action must not dispatch")
	assert.False(t, model.IsOpen(), "confirming a nil action must still close the dialog")
}

func TestOpenReplacesPreviousQuestion(t *testing.T) {
	t.Parallel()

	model := openedTestModel(t)
	secondAction := action.QuitAction{}

	model.Open("second question?", "ok", "nah", secondAction)

	require.True(t, model.IsOpen())
	assert.Equal(t, "second question?", model.text)
	assert.Equal(t, secondAction, model.confirmAction)
	assert.Equal(t, int(cancelButton), model.cursorPosition, "cursor must reset to the cancel button on reopen")
}
