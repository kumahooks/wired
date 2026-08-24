// Package theme loads the color palette shared across all UI components.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"wired/internal/core/config"
)

type Theme struct {
	Surface           color.Color
	SurfaceAlt        color.Color
	BorderPanel       color.Color
	BorderHairline    color.Color
	TextPrimary       color.Color
	TextStrong        color.Color
	TextMuted         color.Color
	TextDim           color.Color
	TextFaint         color.Color
	TextPlaceholder   color.Color
	AccentInteractive color.Color
	AccentDeep        color.Color
	AccentBright      color.Color
	AccentConfirm     color.Color
	AccentLink        color.Color
	AccentPrompt      color.Color
	AccentDanger      color.Color
	AccentError       color.Color
	Track             color.Color
}

// New initializes all of the theme's colors by mapping each hex string to a lipgloss color.
func New(configTheme config.ThemeConfig) Theme {
	return Theme{
		Surface:           lipgloss.Color(configTheme.Surface),
		SurfaceAlt:        lipgloss.Color(configTheme.SurfaceAlt),
		BorderPanel:       lipgloss.Color(configTheme.BorderPanel),
		BorderHairline:    lipgloss.Color(configTheme.BorderHairline),
		TextPrimary:       lipgloss.Color(configTheme.TextPrimary),
		TextStrong:        lipgloss.Color(configTheme.TextStrong),
		TextMuted:         lipgloss.Color(configTheme.TextMuted),
		TextDim:           lipgloss.Color(configTheme.TextDim),
		TextFaint:         lipgloss.Color(configTheme.TextFaint),
		TextPlaceholder:   lipgloss.Color(configTheme.TextPlaceholder),
		AccentInteractive: lipgloss.Color(configTheme.AccentInteractive),
		AccentDeep:        lipgloss.Color(configTheme.AccentDeep),
		AccentBright:      lipgloss.Color(configTheme.AccentBright),
		AccentConfirm:     lipgloss.Color(configTheme.AccentConfirm),
		AccentLink:        lipgloss.Color(configTheme.AccentLink),
		AccentPrompt:      lipgloss.Color(configTheme.AccentPrompt),
		AccentDanger:      lipgloss.Color(configTheme.AccentDanger),
		AccentError:       lipgloss.Color(configTheme.AccentError),
		Track:             lipgloss.Color(configTheme.Track),
	}
}

// Default returns the theme built from config.Defaults.
func Default() Theme {
	return New(config.Defaults().Theme)
}
