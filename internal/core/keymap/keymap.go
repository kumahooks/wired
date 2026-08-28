// Package keymap loads the keymaps (keyboard commands) the application uses.
package keymap

import (
	"fmt"

	"charm.land/bubbles/v2/key"

	"wired/internal/core/config"
)

type ActionsKeyMap struct {
	Playlist     key.Binding
	LibraryStats key.Binding
}

type KeyMap struct {
	MoveLeft    key.Binding
	MoveRight   key.Binding
	Select      key.Binding
	Quit        key.Binding
	GoBack      key.Binding
	OpenActions key.Binding
	Actions     ActionsKeyMap
}

// New initializes all of the keybindings recognized by the application.
func New(bindings config.KeybindMapping) (KeyMap, error) {
	// General keybinds.
	if len(bindings.MoveLeft) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_left must have at least one binding")
	}
	if len(bindings.MoveRight) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_right must have at least one binding")
	}
	if len(bindings.Select) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] select must have at least one binding")
	}
	if len(bindings.Quit) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] quit must have at least one binding")
	}
	if len(bindings.GoBack) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] go_back must have at least one binding")
	}

	// Specific actions through the whichkey flow (lead_key + keybind).
	if len(bindings.OpenActions) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] open_actions must have at least one binding")
	}
	if len(bindings.Actions.Playlist) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] playlist must have at least one binding")
	}
	if len(bindings.Actions.LibraryStats) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] library_stats must have at least one binding")
	}

	return KeyMap{
		MoveLeft:    newBinding(bindings.MoveLeft, "move left"),
		MoveRight:   newBinding(bindings.MoveRight, "move right"),
		Select:      newBinding(bindings.Select, "select"),
		Quit:        newBinding(bindings.Quit, "quit"),
		GoBack:      newBinding(bindings.GoBack, "go back"),
		OpenActions: newBinding(bindings.OpenActions, "open actions"),

		Actions: ActionsKeyMap{
			Playlist:     newBinding(bindings.Actions.Playlist, "playlist"),
			LibraryStats: newBinding(bindings.Actions.LibraryStats, "library stats"),
		},
	}, nil
}

func newBinding(keys []string, description string) key.Binding {
	primary := keys[0]

	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(primary, description),
	)
}
