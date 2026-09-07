package dialog

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

// Style holds the component-specific lipgloss styles for the confirm dialog card.
type Style struct {
	card lipgloss.Style // the bordered centered rectangle of the dialog.
	text lipgloss.Style // renders the question in the card's body.

	buttonFocused lipgloss.Style // renders the button the cursor is at.
	buttonBlurred lipgloss.Style // renders the button the cursor is NOT at.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// card wraps the question and the button row in a rounded border.
		card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(resolvedTheme.BorderPanel).
			Padding(0, cardHorizontalPadding),

		// text is simply the question within the card.
		text: lipgloss.NewStyle().Foreground(resolvedTheme.TextPrimary),

		// the dialog's buttons have two states: either the cursor is selecting them, or not.
		buttonFocused: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextStrong).
			Background(resolvedTheme.AccentInteractive).
			Bold(true).
			Padding(0, 2),
		buttonBlurred: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextMuted).
			Background(resolvedTheme.SurfaceAlt).
			Padding(0, 2),
	}
}
