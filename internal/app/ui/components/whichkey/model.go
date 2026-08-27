// Package whichkey implements the whichkey component, which is inspired directly on which-key.nvim. whichkey has a leader
// keybind, which defaults to "space", and once triggered, a card with commands are then shown with their respective keys.
package whichkey

import (
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// New returns a Model seeded with the buttons actions, the default keymap, and the default theme.
func New(defaultKeyMap keymap.KeyMap) *Model {
	model := &Model{
		isVisible: false,
		keyMap:    defaultKeyMap,
		style:     newStyle(theme.Default()),
	}

	return model
}

// ApplyKeyMap stores the keymap used for mapping the keys list with their respective hints.
func (model *Model) ApplyKeyMap(resolvedKeyMap keymap.KeyMap) {
	model.keyMap = resolvedKeyMap
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

func (model *Model) IsVisible() bool {
	return model.isVisible
}

func (model *Model) FlipIsVisible() {
	model.isVisible = !model.isVisible
}
