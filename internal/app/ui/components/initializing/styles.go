package initializing

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

// Style holds the component-specific lipgloss styles for the initialization screen.
type Style struct {
	card    lipgloss.Style // card is the box that wraps the log area and buttons.
	logArea lipgloss.Style // logArea is the inner area where log lines are rendered.

	logNormal  lipgloss.Style // logNormal renders a non-error log line.
	logWarning lipgloss.Style // logWarning renders a warning log line.
	logError   lipgloss.Style // logError renders an error log line.

	header        lipgloss.Style // header renders the "wire(d) is starting..." title line.
	buttonFocused lipgloss.Style // buttonFocused is the button the cursor is at.
	buttonBlurred lipgloss.Style // buttonBlurred is the button the cursor is NOT at.
	hint          lipgloss.Style // hint renders the navigation hint line below the buttons.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// card renders as a simple area with no background color and with vertical/horizontal paddings.
		card: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextPrimary).
			Width(logAreaWidth+4).
			AlignHorizontal(lipgloss.Center).
			Padding(1, 2),

		// logArea renders the log lines within a small container with a background showing the whole area.
		logArea: lipgloss.NewStyle().Background(resolvedTheme.SurfaceAlt).Padding(0, 1),

		// the texts within the log area have different font colors so the user can tell them apart.
		logNormal:  lipgloss.NewStyle(),
		logWarning: lipgloss.NewStyle().Foreground(resolvedTheme.AccentDanger),
		logError:   lipgloss.NewStyle().Foreground(resolvedTheme.AccentError),

		// header currently renders a short text telling the user what is this component goal.
		header: lipgloss.NewStyle().Foreground(resolvedTheme.TextStrong).Bold(true),

		// the buttons in this component have two states: either the cursor is selecting them, or not.
		buttonFocused: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextStrong).
			Background(resolvedTheme.AccentInteractive).
			Bold(true).
			Padding(0, 2),
		buttonBlurred: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextMuted).
			Background(resolvedTheme.SurfaceAlt).
			Padding(0, 2),

		// hint simply renders, as the last rendered element in this component, keybindings hint in a faint color.
		hint: lipgloss.NewStyle().Foreground(resolvedTheme.TextFaint),
	}
}
