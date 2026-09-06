package librarystats

import (
	"wired/internal/core/audio"
	"wired/internal/core/keymap"
)

// actionButton maps a button to the action it produces on select.
type actionButton int

const (
	scanFullLibraryAction actionButton = iota
	scanNewLibraryAction
	reloadConfigAction
)

// button is a selectable action rendered in the button row.
type button struct {
	label  string
	action actionButton
}

// Model holds the library stats view state and data.
type Model struct {
	library      *audio.Library // points at the orchestrator's library.
	libraryPaths []string       // the configured library directories read from the config.
	libraryStats audio.Stats    // the computed stats for the current library.

	isDiscovering        bool // reports whether a library discovery is currently happening.
	discoveredFilesCount int  // the running total of audio files discovered.
	isDiscoveryDone      bool // reports whether the discovery moved past file discovery into metatag parsing.
	parsedMetatagCount   int  // the running total of files whose metatags were parsed by the active discovery.

	// buttons is the full set of selectable button actions.
	buttons        []button
	cursorPosition int

	// keymap is used to properly map actions.
	keyMap keymap.KeyMap

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}

// topArtistsSection is one named sub-section of the "top artists" card.
type topArtistsSection struct {
	title       string
	counts      []audio.NamedCount
	formatValue func(int) string
}

// lengthsGroupEntry is one "label value annotation" row of a length card.
type lengthsGroupEntry struct {
	label      string
	value      string
	annotation string
}

// formatBar is one row in the "files by format" card.
type formatBar struct {
	format   string  // the file extension, or "(unknown)" for files with no extension.
	count    int     // how many files carry this format.
	bytes    int64   // the total file size for this format, formatted into the size column.
	fraction float64 // the bar fill ratio in [0, 1]: count over total files.
}

// cardLayout carries the per-frame derived dimensions a card draws within.
type cardLayout struct {
	innerWidth  int // usable content width inside the card's borders.
	innerHeight int // usable content height inside the card, excluding borders and title.
}

// card is a drawable unit of the screen: a title, a content drawer, and its fixed size.
type card struct {
	title string
	draw  func(model *Model, layout cardLayout) string

	fixedLines int // pins the card's exact inner content height.
	fixedWidth int // pins the card's exact width.
}

// renderedCards bundles every card currently plugged into the screen.
type renderedCards struct {
	librarySize    card
	libraryPaths   card
	metadataHealth card
	filesByFormat  card
	topArtists     card
	placeholder    card
	trackLengths   card
	albumLengths   card
}
