package whichkey

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/testutil"
)

func TestNewStyleAllFieldsNonZero(t *testing.T) {
	t.Parallel()

	style := newStyle(testutil.DefaultTheme())
	zeroStyle := lipgloss.NewStyle()

	styles := map[string]lipgloss.Style{
		"card":        style.card,
		"key":         style.key,
		"separator":   style.separator,
		"description": style.description,
	}

	for name, gotStyle := range styles {
		assert.NotEqual(t, zeroStyle, gotStyle, "Style.%s equals the zero lipgloss.Style", name)
	}
}

func TestStyleColorsMatchTheme(t *testing.T) {
	t.Parallel()

	resolvedTheme := testutil.CustomTheme()
	style := newStyle(resolvedTheme)

	tests := []struct {
		name      string
		style     lipgloss.Style
		wantColor interface{ RGBA() (r, g, b, a uint32) }
	}{
		{"key", style.key, resolvedTheme.AccentPrompt},
		{"separator", style.separator, resolvedTheme.Track},
		{"description", style.description, resolvedTheme.AccentDeep},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gotColor := test.style.GetForeground()
			require.NotNil(t, gotColor, "style %q foreground is nil", test.name)
			require.NotNil(t, test.wantColor, "theme color for %q is nil", test.name)

			gotR, gotG, gotB, _ := gotColor.RGBA()
			wantR, wantG, wantB, _ := test.wantColor.RGBA()

			assert.Equal(t, wantR, gotR, "style %q red mismatch", test.name)
			assert.Equal(t, wantG, gotG, "style %q green mismatch", test.name)
			assert.Equal(t, wantB, gotB, "style %q blue mismatch", test.name)
		})
	}
}
