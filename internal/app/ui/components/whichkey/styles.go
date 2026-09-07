package whichkey

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

// Style holds the component-specific lipgloss styles for the whichkey card.
type Style struct {
	card lipgloss.Style // card is the full-width bottom rectangle.

	key         lipgloss.Style // key renders the command's shortcut key.
	separator   lipgloss.Style // separator renders the "->" between the key and the description.
	description lipgloss.Style // description renders the command description.

	hint lipgloss.Style // hint is a full-width row that centers the go back action.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// card renders as a simple area with no background color and with no bottom padding.
		card: lipgloss.NewStyle().Padding(1, 2, 0, 2),

		// for each action key mapped in this component styles the key, separator (->), and description, separately.
		key:         lipgloss.NewStyle().Foreground(resolvedTheme.AccentPrompt),
		separator:   lipgloss.NewStyle().Foreground(resolvedTheme.Track),
		description: lipgloss.NewStyle().Foreground(resolvedTheme.AccentDeep),

		// hint simply renders, as the last rendered element in this component. Currentlyt it's just the keybind "go_back".
		hint: lipgloss.NewStyle(),
	}
}
