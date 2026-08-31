package librarystats

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

// Style holds the component-specific lipgloss styles for the library stats screen.
type Style struct {
	header          lipgloss.Style // header renders the screen title line.
	headerSeparator lipgloss.Style // headerSeparator renders the dot separator between the header and its subtitle.
	headerSubtitle  lipgloss.Style // headerSubtitle renders the subtitle next to the screen's title.

	card      lipgloss.Style // card is the base bordered box every stat card wraps its content with.
	cardTitle lipgloss.Style // cardTitle renders the title written into the card's top border.

	label lipgloss.Style // label renders stat row labels.
	value lipgloss.Style // value renders their values.

	formatShareBar      lipgloss.Style // formatShareBar renders the bar characters in the "files by format" card.
	formatShareEmptyBar lipgloss.Style // formatShareEmptyBar renders the unfilled remainder in the "files by format" card.

	libraryPathIndex lipgloss.Style // libraryPathIndex renders the number prefix of a row in "library paths" card.
	libraryPath      lipgloss.Style // libraryPath renders the path of a row in "library paths" card.

	buttonFocused lipgloss.Style // buttonFocused renders the button the cursor is at.
	scanStatus    lipgloss.Style // scanStatus renders the live scan progress line below the button.
	dash          lipgloss.Style // dash renders the placeholder dashes for empty values.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// header is the very first text in the screen, with its subtitle right next to it.
		header:          lipgloss.NewStyle().Foreground(resolvedTheme.TextStrong).Bold(true),
		headerSeparator: lipgloss.NewStyle().Foreground(resolvedTheme.TextFaint),
		headerSubtitle:  lipgloss.NewStyle().Foreground(resolvedTheme.AccentDeep),

		// card wraps every stat panel in a rounded border.
		card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(resolvedTheme.BorderPanel),
		// cardTitle renders card names.
		cardTitle: lipgloss.NewStyle().Foreground(resolvedTheme.AccentInteractive).Bold(true),

		// used in "library size" card, we show this as "{label}		{value}".
		label: lipgloss.NewStyle().Foreground(resolvedTheme.TextMuted),
		value: lipgloss.NewStyle().Foreground(resolvedTheme.AccentBright).Bold(true),

		// "files by format" card renders the formats of each track.
		formatShareBar:      lipgloss.NewStyle().Foreground(resolvedTheme.AccentPrompt),
		formatShareEmptyBar: lipgloss.NewStyle().Foreground(resolvedTheme.Track),

		// "library paths" numbers each library path in [0, n).
		libraryPathIndex: lipgloss.NewStyle().Foreground(resolvedTheme.AccentDeep),
		libraryPath:      lipgloss.NewStyle().Foreground(resolvedTheme.TextDim),

		// TODO: create more actions (reload config maybe?). The button has two states: either the cursor is selecting it, or not.
		buttonFocused: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextStrong).
			Background(resolvedTheme.AccentInteractive).
			Bold(true).
			Padding(0, 2),

		// dash renders as dashes for values that are either empty or nil.
		dash: lipgloss.NewStyle().Foreground(resolvedTheme.TextFaint),

		// scanStatus shows scan progress below the button while a scan is running.
		scanStatus: lipgloss.NewStyle().Foreground(resolvedTheme.TextMuted).Italic(true),
	}
}
