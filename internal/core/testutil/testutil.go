// Package testutil provides shared test helpers for the wired test suite.
package testutil

import (
	"regexp"
	"testing"

	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// ansiEscSeq matches CSI sequences (the only escape sequences lipgloss emits for styling).
var ansiEscSeq = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

// StripANSI removes CSI escape sequences so golden files and assertions compare against plain text.
func StripANSI(value string) string {
	return ansiEscSeq.ReplaceAllString(value, "")
}

// DefaultKeyMap builds the keymap from config.Defaults, fataling on a parse error.
func DefaultKeyMap(t *testing.T) keymap.KeyMap {
	t.Helper()

	keyMap, err := keymap.New(config.Defaults().Keybinds)
	if err != nil {
		t.Fatalf("keymap.New(defaults) error: %v", err)
	}

	return keyMap
}

// DefaultTheme returns the theme built from config.Defaults.
func DefaultTheme() theme.Theme {
	return theme.New(config.Defaults().Theme)
}

// CustomTheme returns a deterministic theme with colors distinct from the defaults, so tests can assert the values.
func CustomTheme() theme.Theme {
	return theme.New(config.ThemeConfig{
		Surface:           "#0a0a0a",
		SurfaceAlt:        "#0b0b0b",
		BorderPanel:       "#1a1a1a",
		BorderHairline:    "#0c0c0c",
		TextPrimary:       "#aaaaaa",
		TextStrong:        "#ffffff",
		TextMuted:         "#888888",
		TextDim:           "#999999",
		TextFaint:         "#777777",
		TextPlaceholder:   "#666666",
		AccentInteractive: "#b8748a",
		AccentDeep:        "#ad7084",
		AccentBright:      "#d94ea0",
		AccentConfirm:     "#b25c83",
		AccentLink:        "#a65e6e",
		AccentPrompt:      "#b90074",
		AccentDanger:      "#ff6b6b",
		AccentError:       "#112233",
		Track:             "#4a4a4a",
	})
}

// NewDiscoveryProgress builds a DiscoveryProgress snapshot with the given counts for SetProgress-based tests.
func NewDiscoveryProgress(discovered int, parsed int, discoveryDone bool) *audio.DiscoveryProgress {
	progress := audio.NewDiscoveryProgress()
	progress.AddDiscovered(discovered)
	progress.AddParsed(parsed)
	if discoveryDone {
		progress.SetDiscoveryDone()
	}

	return progress
}
