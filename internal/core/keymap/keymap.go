// Package keymap defines the keymaps (keyboard commands) the application uses.
package keymap

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit key.Binding
}

// TODO: We will want to map this to a config file and load on start.
// Furthermore, keymaps might be conditional to screen states. It's likely we will need to expand this eventually.
func New() KeyMap {
	keyMap := KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "quit"),
		),
	}

	return keyMap
}
