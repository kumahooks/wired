package librarystats

// Layout constants for the librarystats view.
const (
	// borderWidth is the horizontal space a card's left+right border consumes.
	borderWidth = 2

	// smallCardWidth is the width of the smaller cards, taking half the width space.
	smallCardWidth = 34
	// smallCardHeight is the inner height of the smaller cards.
	smallCardHeight = 9

	// bigCardWidth is the width of the bigger cards, taking the whole width space.
	bigCardWidth = 2 * smallCardWidth
	// librarySizeCardHeight is the inner height of the "library size" card.
	librarySizeCardHeight = 3

	// labelWidth is the fixed width reserved for stat row labels.
	labelWidth = 11
	// barWidth is the glyph width of the share bars.
	barWidth = 10
)

// MinWidth is the smallest terminal width the card grid can render.
const MinWidth = 4 + bigCardWidth

// Decorative constants for the librarystats view.
const (
	// headerTitle is the screen's title line.
	headerTitle = "library stats"
	// headerTitle is the screen's title separator, sitting between the title and he subtitle.
	headerSeparator = " · "
	// headerSubtitle is a subtitle next to the header title.
	headerSubtitle = "present day // present time"

	// dashPlaceholder is a placeholder value for when the result is empty/zero.
	dashPlaceholder = "--"

	// librarySizeCardTitle is the title for the "library size" card.
	librarySizeCardTitle = "LIBRARY SIZE"
	// formatsCardTitle is the title for the "files by format" card.
	formatsCardTitle = "FILES BY FORMAT"
	// pathsCardTitle is the title for the "library paths" card.
	pathsCardTitle = "LIBRARY PATHS"

	// rescanButtonLabel is the "rescan library paths" button's label.
	rescanButtonLabel = "rescan library paths"

	// noPathsText is shown in the paths card when no library paths are found.
	noPathsText = "no library paths in config"

	// unknownFormatText labels files without a file extension.
	unknownFormatText = "(unknown)"

	// maxVisiblePathRows caps how many library path rows can be rendered before the remainder line.
	maxVisiblePathRows = smallCardHeight - 3
	// maxVisibleFormatRows caps how many format bars can be rendered.
	maxVisibleFormatRows = smallCardHeight - 2
)
