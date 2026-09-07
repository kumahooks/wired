package dialog

import (
	"time"

	"wired/internal/app/ui/action"
	"wired/internal/core/keymap"
)

// dialogButton indexes the confirm button and the cancel button.
type dialogButton int

const (
	confirmButton dialogButton = iota
	cancelButton
)

// Model holds the confirm dialog state. a description, its two buttons, and the action the confirm button carries.
type Model struct {
	text          string        // the description rendered in the card's body.
	confirmLabel  string        // the confirm button's label.
	cancelLabel   string        // the cancel button's label.
	confirmAction action.Action // the action dispatched through the root on confirm.

	isOpen         bool      // whether the dialog is currently shown.
	cursorPosition int       // the selected button 0 is confirm, 1 is cancel.
	lastKeyAt      time.Time // the instant the most recent absorbed key arrived.

	keyMap keymap.KeyMap // the keymap used for button navigation and selection.
	style  Style         // the styles (such as lipgloss colors) used in the view rendering.
}
