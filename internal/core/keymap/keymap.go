// Package keymap loads the keymaps (keyboard commands) the application uses.
package keymap

import (
	"fmt"

	"charm.land/bubbles/v2/key"

	"wired/internal/core/config"
)

type ActionsKeyMap struct {
	Library      key.Binding
	Playlist     key.Binding
	LibraryStats key.Binding
	ReloadConfig key.Binding
}

type KeyMap struct {
	MoveLeft   key.Binding
	MoveRight  key.Binding
	MoveUp     key.Binding
	MoveDown   key.Binding
	MoveTop    key.Binding
	MoveBottom key.Binding

	Select key.Binding
	Quit   key.Binding
	GoBack key.Binding

	ViewLibrary      key.Binding
	ViewPlaylist     key.Binding
	ViewLibraryStats key.Binding

	OpenActions key.Binding
	Actions     ActionsKeyMap
}

// New initializes all of the keybindings recognized by the application.
func New(bindings config.KeybindMapping) (KeyMap, error) {
	// General keybinds.
	moveLeftKeys := canonicalTeaKeyNames(bindings.MoveLeft)
	moveRightKeys := canonicalTeaKeyNames(bindings.MoveRight)
	moveUpKeys := canonicalTeaKeyNames(bindings.MoveUp)
	moveDownKeys := canonicalTeaKeyNames(bindings.MoveDown)
	moveTopKeys := canonicalTeaKeyNames(bindings.MoveTop)
	moveBottomKeys := canonicalTeaKeyNames(bindings.MoveBottom)

	selectKeys := canonicalTeaKeyNames(bindings.Select)
	quitKeys := canonicalTeaKeyNames(bindings.Quit)
	goBackKeys := canonicalTeaKeyNames(bindings.GoBack)

	viewLibraryKeys := canonicalTeaKeyNames(bindings.ViewLibrary)
	viewPlaylistKeys := canonicalTeaKeyNames(bindings.ViewPlaylist)
	viewLibraryStatsKeys := canonicalTeaKeyNames(bindings.ViewLibraryStats)

	// Action keybinds.
	openActionsKeys := canonicalTeaKeyNames(bindings.OpenActions)
	libraryKeys := canonicalTeaKeyNames(bindings.Actions.Library)
	playlistKeys := canonicalTeaKeyNames(bindings.Actions.Playlist)
	libraryStatsKeys := canonicalTeaKeyNames(bindings.Actions.LibraryStats)
	reloadConfigKeys := canonicalTeaKeyNames(bindings.Actions.ReloadConfig)

	if len(bindings.MoveLeft) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_left must have at least one binding")
	}
	if len(bindings.MoveRight) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_right must have at least one binding")
	}
	if len(bindings.MoveUp) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_up must have at least one binding")
	}
	if len(bindings.MoveDown) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_down must have at least one binding")
	}
	if len(bindings.MoveTop) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_top must have at least one binding")
	}
	if len(bindings.MoveBottom) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] move_bottom must have at least one binding")
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
	if len(bindings.Actions.Library) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] library must have at least one binding")
	}
	if len(bindings.Actions.LibraryStats) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] library_stats must have at least one binding")
	}
	if len(bindings.Actions.ReloadConfig) == 0 {
		return KeyMap{}, fmt.Errorf("[keymap:New] reload_config must have at least one binding")
	}

	generalKeys := [][]string{
		moveLeftKeys,
		moveRightKeys,
		moveUpKeys,
		moveDownKeys,
		moveTopKeys,
		moveBottomKeys,

		selectKeys,
		quitKeys,
		goBackKeys,

		viewLibraryKeys,
		viewPlaylistKeys,
		viewLibraryStatsKeys,

		openActionsKeys,
	}
	actionKeys := [][]string{
		libraryKeys,
		playlistKeys,
		libraryStatsKeys,
		reloadConfigKeys,
	}

	if err := rejectDuplicateKeys(generalKeys); err != nil {
		return KeyMap{}, err
	}
	if err := rejectDuplicateKeys(actionKeys); err != nil {
		return KeyMap{}, err
	}

	return KeyMap{
		MoveLeft:   newBinding(moveLeftKeys, "move left"),
		MoveRight:  newBinding(moveRightKeys, "move right"),
		MoveUp:     newBinding(moveUpKeys, "scroll up"),
		MoveDown:   newBinding(moveDownKeys, "scroll down"),
		MoveTop:    newBinding(moveTopKeys, "scroll to top"),
		MoveBottom: newBinding(moveBottomKeys, "scroll to bottom"),
		Select:     newBinding(selectKeys, "select"),

		Quit:        newBinding(quitKeys, "quit"),
		GoBack:      newBinding(goBackKeys, "go back"),
		OpenActions: newBinding(openActionsKeys, "open actions"),

		ViewLibrary:      newBinding(viewLibraryKeys, "view library ui"),
		ViewPlaylist:     newBinding(viewPlaylistKeys, "view playlist ui"),
		ViewLibraryStats: newBinding(viewLibraryStatsKeys, "view library stats ui"),

		Actions: ActionsKeyMap{
			Library:      newBinding(libraryKeys, "library screen"),
			Playlist:     newBinding(playlistKeys, "playlist screen"),
			LibraryStats: newBinding(libraryStatsKeys, "library stats screen"),
			ReloadConfig: newBinding(reloadConfigKeys, "reload config"),
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

// keyAliases maps config-spelled key names to bubbletea's rendered keystroke strings.
var keyAliases = map[string]string{
	"escape": "esc",
}

// canonicalTeaKeyNames rewrites config key names to the runtime keystroke spelling the key.Matches comparison uses.
func canonicalTeaKeyNames(keys []string) []string {
	canonicalKeys := make([]string, 0, len(keys))
	for _, keyName := range keys {
		if alias, isAliased := keyAliases[keyName]; isAliased {
			canonicalKeys = append(canonicalKeys, alias)
			continue
		}

		canonicalKeys = append(canonicalKeys, keyName)
	}

	return canonicalKeys
}

// rejectDuplicateKeys rejects any key string appearing in more than one of the given binding lists, or twice in one list.
func rejectDuplicateKeys(keyLists [][]string) error {
	seenKeys := make(map[string]bool)
	for _, keys := range keyLists {
		for _, binding := range keys {
			if seenKeys[binding] {
				return fmt.Errorf("[keymap:New] duplicate keybinding %q", binding)
			}

			seenKeys[binding] = true
		}
	}

	return nil
}
