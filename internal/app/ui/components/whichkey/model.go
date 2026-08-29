// Package whichkey implements the whichkey component, which is inspired directly on which-key.nvim. whichkey has a leader
// keybind, which defaults to "space", and once triggered, a card with commands are then shown with their respective keys.
package whichkey

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/core/theme"
)

// New returns a Model seeded with the default theme, while bindings are provided through the SetBindings primitive.
func New() *Model {
	model := &Model{
		isVisible: false,
		style:     newStyle(theme.Default()),
	}

	return model
}

// SetBindings replaces the active binding list.
func (model *Model) SetBindings(bindings []action.Binding) {
	model.bindings = bindings
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

func (model *Model) IsVisible() bool {
	return model.isVisible
}

// ApplyCloseKeybinding stores the close key shown in the card's hint line.
func (model *Model) ApplyCloseKeybinding(closeKeybind key.Binding) {
	model.closeKey = closeKeybind.Help().Key
}

func (model *Model) HandleMessage(message tea.Msg) action.Action {
	keyPress, ok := message.(tea.KeyPressMsg)
	if !ok {
		return action.NoAction{}
	}

	model.flipIsVisible()

	for _, binding := range model.bindings {
		if key.Matches(keyPress, binding.Keys) {
			return binding.Action
		}
	}

	return action.NoAction{}
}

func (model *Model) flipIsVisible() {
	model.isVisible = !model.isVisible
}
