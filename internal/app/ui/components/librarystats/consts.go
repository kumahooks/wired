package librarystats

// Layout constants for the librarystats view.
const (
	borderWidth    = 2  // the horizontal space a card's left+right border consumes.
	smallCardWidth = 34 // the smallest width a card can render at.
	labelWidth     = 12 // the fixed width reserved for the row labels/keys.

	// the terminal width thresholds of the grid's layout tiers.
	fullGridWidth    = librarySizeCardWidth + metadataFormatGroupWidth + topArtistsCardWidth
	compactGridWidth = librarySizeCardWidth + metadataFormatGroupWidth
	lengthsRowWidth  = 2 * lengthsGroupWidth // the lengths row renders two side-by-side length cards.

	// the terminal height thresholds of the grid's layout tiers.
	cardVerticalOverhead = 3 // the vertical space a card's borders and title line consume.
	headerLines          = 1 // the header's fixed line count.
	buttonRowLines       = 1 // the buttons row's fixed line count.

	// the columns row is a stack of two cards per column group.
	columnsRowLines  = librarySizeCardLines + pathsCardLines + 2*cardVerticalOverhead
	lengthsRowLines  = lengthsGroupLines + cardVerticalOverhead
	fullGridLines    = headerLines + columnsRowLines + lengthsRowLines + buttonRowLines + discoveryStatusLines
	compactGridLines = headerLines + columnsRowLines + buttonRowLines + discoveryStatusLines

	buttonRowLeftPadding = 1 // column of whitespaces kept at the left of the actions buttons.
	discoveryStatusLines = 2 // the fixed line count of the discovery status lines below the grid.

	formatCardVisibleRows = 8                         // how many rows "files by format" renders before it collapses into the hint.
	formatCardLines       = formatCardVisibleRows + 1 // the "files by format"'s full inner content: format lines + hint.

	librarySizeCardWidth = 27 // the pinned width of the "library size" and "library paths" cards.
	librarySizeCardLines = 4  // the inner content height of the "library size" card.

	pathsCardVisibleRows = 8                        // how many rows "library paths" render before it collapses into the hint.
	pathsCardLines       = pathsCardVisibleRows + 1 // the "library paths"'s full inner content: paths + hint.
	pathsCardHintText    = "...and %d more paths"   // the remainder line of the "library paths" card.

	metadataFormatGroupWidth   = smallCardWidth + 2 // the pinned width of the "metadata health" and "files by format" cards.
	metadataCardLines          = 4                  // the "metadata health" fixed inner content height.
	formatCardKeyWidth         = 8                  // the width reserved for the format/year name in "files by format" rows.
	formatCardCountWidth       = 5                  // the width reserved for the count column in "files by format" rows.
	formatCardBytesColumnWidth = 11                 // the width of the right-aligned bytes column in the "files by format" card.
	formatCardBarWidth         = 8                  // the glyph width of the format bars.

	lengthsGroupWidth      = smallCardWidth + 15 // the pinned width of the "track lengths" and "album lengths" cards.
	lengthsGroupLines      = 3                   // the fixed inner content height of the length cards.
	lengthsGroupLabelWidth = 9                   // the narrower label/key width the length cards use.

	topArtistsCardSections       = 3                  // how many sub-sections the top artists card hold.
	topArtistsCardSectionEntries = 3                  // how many entries each section shows at most.
	topArtistsCardWidth          = smallCardWidth + 1 // the pinned width of the top artists card.
	topArtistsCardLines          = topArtistsCardSections * (1 + topArtistsCardSectionEntries)
)

// Decorative constants for the librarystats view.
const (
	headerTitle     = "library stats"               // the screen's title line.
	headerSeparator = " · "                         // the screen's title separator, between the title and the subtitle.
	headerSubtitle  = "present day // present time" // the subtitle next to the header title.

	// emptyPlaceholder is a placeholder value for when the result is empty/zero.
	emptyPlaceholder = "--"

	sizeCardTitle           = "LIBRARY SIZE" // the title for the "library size" card.
	sizeCardFilesTotalLabel = "files total"
	sizeCardBytesTotalLabel = "bytes total"
	sizeCardAvgBytesLabel   = "avg bytes"
	sizeCardHeaviestLabel   = "heaviest"

	metadataCardTitle       = "METADATA HEALTH" // the title for the "metadata health" card.
	metadataCardTitleLabel  = "no title"
	metadataCardArtistLabel = "no artist"
	metadataCardAlbumLabel  = "no album"
	metadataCardDupesLabel  = "duplicates"

	formatCardTitle        = "FILES BY FORMAT"         // the title for the "files by format" card.
	formatCardUnknownValue = "(unknown)"               // labels files without a file extension.
	formatCardMoreText     = "...and %d more formats~" // the remainder line of the formats card.

	topArtistsCardTitle       = "TOP ARTISTS" // the title for the "top artists" card.
	topArtistsByFilesTitle    = "by files"
	topArtistsByAlbumsTitle   = "by albums"
	topArtistsByDurationTitle = "by playtime"

	// the title of the reserved card slot below the "top artists" card.
	placeholderCardTitle = "PLACEHOLDER"

	trackLengthsCardTitle    = "TRACK LENGTHS" // the title for the "track lengths" card.
	albumLengthsCardTitle    = "ALBUM LENGTHS" // the title for the "album lengths" card.
	lengthsCardLongestLabel  = "longest"
	lengthsCardShortestLabel = "shortest"
	lengthsCardAverageLabel  = "average"
	trackLengthsAverageValue = "all tracks"
	albumLengthsAverageValue = "all albums"

	annotatedValueText     = " (%s)"            // wraps the annotation appended to length card values.
	noTracksWithLengthText = "no duration data" // shown in length cards when no track carries a length tag.

	pathsCardTitle     = "LIBRARY PATHS" // the title for the "library paths" card.
	pathsCardEmptyText = "no library paths"

	buttonSpacing       = " "                       // the spacing string between every action button.
	scanFullButtonLabel = "scan library (full)"     // the "scan library (full)" button's label.
	scanNewButtonLabel  = "scan library (new only)" // the "scan library (new only)" button's label.

	// the confirm dialog opened by the "scan library (full)" and "scan library (new)" buttons.
	scanFullDialogText     = "a full scan rebuilds the library from scratch, wiping and overwriting the current cache."
	scanNewDialogText      = "a new scan searches for files not yet scanned and append to the current file cache."
	scanDialogConfirmLabel = "scan"
	scanDialogCancelLabel  = "cancel"

	scanStatusText  = "found %d audio files..."        // the discovery status line below the button, while file discovery runs.
	scanFoundText   = "%d audio files have been found" // the first status line shown while the metatag parsing phase runs.
	scanParsingText = "parsing %d/%d files"            // the second status line shown while the metatag parsing phase runs.

	// the message rendered in place of the grid when the window cannot fit any layout tier.
	smallWindowText = "the window is too small to render the library stats."
)
