package librarystats

import (
	"wired/internal/core/audio"
	"wired/internal/core/keymap"
)

// Model holds the library stats view state and data.
type Model struct {
	// audioFiles points at the orchestrator's library.
	audioFiles *[]audio.File

	// libraryPaths are the configured library directories read from the config.
	libraryPaths []string

	// keymap is used to properly map actions.
	keyMap keymap.KeyMap

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
