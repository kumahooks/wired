// Package keymap loads the keymaps (keyboard commands) the application uses.
package keymap

import (
	"strings"

	"charm.land/bubbles/v2/key"

	"wired/internal/core/config"
)

type KeyMap struct {
	Quit key.Binding
}

// New initializes all of the keybindings recognized by the application.
func New(bindings config.KeybindMapping) KeyMap {
	keyMap := KeyMap{
		Quit: newBinding(bindings.Quit, "quit the application"),
	}

	return keyMap
}

func newBinding(keys []string, description string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(strings.Join(keys, "/"), description),
	)
}
