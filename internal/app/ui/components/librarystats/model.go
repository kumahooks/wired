// Package librarystats implements a view responsible for showing data and actions related to the user's loaded library.
package librarystats

import (
	"wired/internal/core/audio"
	"wired/internal/core/theme"
)

// Model holds the library stats view state and data.
type Model struct {
	// audioFiles points at the orchestrator's library.
	audioFiles *[]audio.File

	// libraryPaths are the configured library directories read from the config.
	libraryPaths []string

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}

// formatBar is one row in the formats card.
type formatBar struct {
	// format is the file extension, or "(unknown)" for files with no extension.
	format string

	// count is how many files carry this format.
	count int

	// fraction is count over total files (in [0, 1]).
	fraction float64
}

// New returns a Model for the given library pointer, with empty paths and default theme.
func New(audioFiles *[]audio.File) *Model {
	return &Model{
		audioFiles:   audioFiles,
		libraryPaths: []string{},
		style:        newStyle(theme.Default()),
	}
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// SetLibraryPaths sets the library paths.
func (model *Model) SetLibraryPaths(libraryPaths []string) {
	model.libraryPaths = libraryPaths
}
