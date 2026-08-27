package whichkey

import (
	"wired/internal/core/keymap"
)

// buttonAction maps an action to the keybind the user presses.
type buttonAction int

const (
	actionScanLibraryFull buttonAction = iota
)

type Model struct {
	// isVisible is the flag that decides whether we are in the middle of a whichkey combo or not. It is what makes an
	// action sink into this component, render the card with the actions mapped, etc.
	isVisible bool

	// keymap is used to properly map actions, and render their hint on the whichkey card.
	keyMap keymap.KeyMap

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}
