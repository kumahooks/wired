package whichkey

import (
	"wired/internal/app/ui/action"
)

type Model struct {
	// isVisible is the flag that decides whether the user is in the middle of a whichkey combo or not. It is what makes an
	// action sink into this component, render the card with the actions mapped, etc.
	isVisible bool

	// bindings is the list of active bindings based on the current UI's state, compiled and pushed by the UIModel.
	bindings []action.Binding

	// closeKey is the key hint rendered on the card's last line for closing the card.
	closeKey string

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}
