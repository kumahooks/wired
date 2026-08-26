package initializing

import (
	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

// Style holds the component-specific lipgloss styles for the initialization screen.
type Style struct {
	panel   lipgloss.Style // panel is the box that wraps the log area and buttons.
	logArea lipgloss.Style // logArea is the inner area where log lines are rendered.

	logNormal  lipgloss.Style // logNormal renders a non-error log line.
	logWarning lipgloss.Style // logWarning renders a warning log line.
	logError   lipgloss.Style // logError renders an error log line.

	header        lipgloss.Style // header renders the "wire(d) is starting..." title line.
	progress      lipgloss.Style // progress renders the live "fetching N audio files..." line.
	buttonFocused lipgloss.Style // buttonFocused is the button the cursor is at.
	buttonBlurred lipgloss.Style // buttonBlurred is the button the cursor is NOT at.
	hint          lipgloss.Style // hint renders the navigation hint line below the buttons.
}

func newStyle(resolvedTheme theme.Theme) Style {
	return Style{
		// Panel renders as a simple area with no background color and with vertical/horizontal paddings.
		panel: lipgloss.NewStyle().Foreground(resolvedTheme.TextPrimary).Padding(1, 2),

		// Log area renders the log lines within a small container with a background showing the whole area.
		logArea: lipgloss.NewStyle().Background(resolvedTheme.SurfaceAlt).Padding(0, 1),

		// Texts within the log area have different font colors so the user can tell them apart.
		logNormal:  lipgloss.NewStyle(),
		logWarning: lipgloss.NewStyle().Foreground(resolvedTheme.AccentDanger),
		logError:   lipgloss.NewStyle().Foreground(resolvedTheme.AccentError),

		// Header currently renders a short text telling the user what is this component goal.
		header: lipgloss.NewStyle().Foreground(resolvedTheme.TextStrong).Bold(true),

		// Progress renders the live counter line ("fetching N audio files...") while a fetch is happening.
		progress: lipgloss.NewStyle().Foreground(resolvedTheme.AccentInteractive),

		// Buttons in this component have two states: either the cursor is selecting them, or not.
		buttonFocused: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextStrong).
			Background(resolvedTheme.AccentInteractive).
			Bold(true).
			Padding(0, 2),
		buttonBlurred: lipgloss.NewStyle().
			Foreground(resolvedTheme.TextMuted).
			Background(resolvedTheme.SurfaceAlt).
			Padding(0, 2),

		// Hint simply renders, as the last rendered element in this component, keybindings hint in a faint color.
		hint: lipgloss.NewStyle().Foreground(resolvedTheme.TextFaint),
	}
}
