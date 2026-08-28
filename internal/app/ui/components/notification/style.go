package notification

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

type Style struct {
	card    lipgloss.Style // card is the small dialog rectangle of a notification.
	content lipgloss.Style // content represents the text within a notification.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// card draws a rounded border around the wrapped notification text.
		card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			BorderForeground(resolvedTheme.AccentInteractive),

		// content is simply the text within the card.
		content: lipgloss.NewStyle().Foreground(resolvedTheme.TextPrimary),
	}
}
