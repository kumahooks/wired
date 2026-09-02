// Package librarystats implements a view responsible for showing data and actions related to the user's loaded library.
package librarystats

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/core/audio"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// New returns a Model for the given library pointer, with empty paths and default theme.
func New(defaultKeyMap keymap.KeyMap, library *audio.Library) *Model {
	return &Model{
		library:      library,
		libraryPaths: []string{},
		libraryStats: audio.Stats{},
		keyMap:       defaultKeyMap,
		style:        newStyle(theme.Default()),
	}
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// ApplyKeyMap stores the keymap used for the buttons actions on the view.
func (model *Model) ApplyKeyMap(resolvedKeyMap keymap.KeyMap) {
	model.keyMap = resolvedKeyMap
}

// SetLibraryPaths sets the library paths.
func (model *Model) SetLibraryPaths(libraryPaths []string) {
	model.libraryPaths = libraryPaths
}

// ComputeStats recomputes the current library stats from the library pointer.
func (model *Model) ComputeStats() {
	model.libraryStats = audio.ComputeStats(model.library)
}

// StartDiscovery marks a library discovery as running and resets all discovery progress state.
func (model *Model) StartDiscovery() {
	model.isDiscovering = true
	model.discoveredFilesCount = 0
	model.isDiscoveryDone = false
	model.parsedMetatagCount = 0
}

// SetProgress reads the discovery progress reporter into the component's render state. The phase shown is decided by
// the reporter's discoveryDone flag.
func (model *Model) SetProgress(progress *audio.DiscoveryProgress) {
	if progress == nil {
		return
	}

	model.discoveredFilesCount = progress.DiscoveredCount()

	if !progress.DiscoveryDone() {
		return
	}

	model.isDiscoveryDone = true
	model.parsedMetatagCount = progress.ParsedCount()
}

// DiscoveredFilesCount returns the last ticked count of discovered audio files.
func (model *Model) DiscoveredFilesCount() int {
	return model.discoveredFilesCount
}

// FinishDiscovery marks the library discovery as no longer running and clears all discovery progress state.
func (model *Model) FinishDiscovery() {
	model.isDiscovering = false
	model.isDiscoveryDone = false
	model.parsedMetatagCount = 0
}

// HandleMessage handles keyboard navigation and actions for the button row.
func (model *Model) HandleMessage(message tea.Msg) action.Action {
	keyPress, ok := message.(tea.KeyPressMsg)
	if !ok {
		return action.NoAction{}
	}

	if key.Matches(keyPress, model.keyMap.Select) {
		return action.DiscoverLibraryFullAction{}
	}

	return action.NoAction{}
}
