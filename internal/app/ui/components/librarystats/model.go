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
		keyMap:       defaultKeyMap,
		style:        newStyle(theme.Default()),
	}
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// ApplyKeyMap stores the keymap used for the rescan button's activation.
func (model *Model) ApplyKeyMap(resolvedKeyMap keymap.KeyMap) {
	model.keyMap = resolvedKeyMap
}

// SetLibraryPaths sets the library paths.
func (model *Model) SetLibraryPaths(libraryPaths []string) {
	model.libraryPaths = libraryPaths
}

// StartScan marks a library scan as running and resets the progress counter.
func (model *Model) StartScan() {
	model.isScanning = true
	model.scannedFilesCount = 0
}

// SetScanProgress updates the running count of found audio files for the active scan.
func (model *Model) SetScanProgress(filesCount int) {
	model.scannedFilesCount = filesCount
}

// FinishScan marks the library scan as no longer running.
func (model *Model) FinishScan() {
	model.isScanning = false
}

// HandleMessage handles keyboard navigation and actions for the button row.
func (model *Model) HandleMessage(message tea.Msg) action.Action {
	keyPress, ok := message.(tea.KeyPressMsg)
	if !ok {
		return action.NoAction{}
	}

	if key.Matches(keyPress, model.keyMap.Select) {
		return action.ScanLibraryFullAction{}
	}

	return action.NoAction{}
}
