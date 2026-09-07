// Package dialog implements the confirm dialog overlay.
package dialog

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// New returns a Model seeded with the default theme, hidden until an Open call shows it.
func New() *Model {
	return &Model{
		isOpen: false,
		style:  newStyle(theme.Default()),
	}
}

// Open shows the dialog with the given question and buttons.
func (model *Model) Open(text string, confirmLabel string, cancelLabel string, confirmAction action.Action) {
	model.text = text
	model.confirmLabel = confirmLabel
	model.cancelLabel = cancelLabel
	model.confirmAction = confirmAction
	model.cursorPosition = int(cancelButton)
	model.lastKeyAt = time.Now()
	model.isOpen = true
}

// Close hides the dialog, clearing the stored question and action.
func (model *Model) Close() {
	model.isOpen = false
	model.text = ""
	model.confirmLabel = ""
	model.cancelLabel = ""
	model.confirmAction = nil
}

// IsOpen reports whether the dialog is currently shown.
func (model *Model) IsOpen() bool {
	return model.isOpen
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// ApplyKeyMap stores the keymap used for the dialog's navigation and selection.
func (model *Model) ApplyKeyMap(resolvedKeyMap keymap.KeyMap) {
	model.keyMap = resolvedKeyMap
}

// HandleMessage handles keyboard navigation between the two buttons and their selection. We drop keys within the key grace
// period to avoid the user confirming an action by mistake.
func (model *Model) HandleMessage(message tea.Msg) action.Action {
	keyPress, ok := message.(tea.KeyPressMsg)
	if !ok || !model.isOpen {
		return action.NoAction{}
	}

	// The dialog handles keys only once input has been quiet for the grace period.
	if time.Since(model.lastKeyAt) < KeyGraceQuietPeriod {
		model.lastKeyAt = time.Now()
		return action.NoAction{}
	}

	switch {
	case key.Matches(keyPress, model.keyMap.MoveLeft), key.Matches(keyPress, model.keyMap.MoveRight):
		model.cursorPosition = int(confirmButton) + int(cancelButton) - model.cursorPosition
	case key.Matches(keyPress, model.keyMap.Select):
		confirmAction := model.confirmAction
		isConfirm := model.cursorPosition == int(confirmButton)

		model.Close()

		if isConfirm && confirmAction != nil {
			return confirmAction
		}

		return action.NoAction{}
	case key.Matches(keyPress, model.keyMap.GoBack):
		model.Close()
	}

	return action.NoAction{}
}
