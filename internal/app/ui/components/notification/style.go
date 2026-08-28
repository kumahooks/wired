package notification

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

type Style struct {
	card    lipgloss.Style // card is the small dialog rectangle of a notification.
	content lipgloss.Style // text represents the content within a notification.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		card:    lipgloss.NewStyle(),
		content: lipgloss.NewStyle().Foreground(resolvedTheme.AccentDeep),
	}
}
