package initializing

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
		"panel":         style.panel,
		"logArea":       style.logArea,
		"logWarning":    style.logWarning,
		"logError":      style.logError,
		"header":        style.header,
		"buttonFocused": style.buttonFocused,
		"buttonBlurred": style.buttonBlurred,
		"hint":          style.hint,
	}

	for name, gotStyle := range styles {
		assert.NotEqual(t, zeroStyle, gotStyle, "Style.%s equals the zero lipgloss.Style", name)
	}
}

func TestButtonFocusedForegroundMatchesTextStrong(t *testing.T) {
	t.Parallel()

	defaultTheme := testutil.DefaultTheme()
	style := newStyle(defaultTheme)

	gotColor := style.buttonFocused.GetForeground()
	require.NotNil(t, gotColor, "buttonFocused foreground is nil")

	wantColor := defaultTheme.TextStrong
	require.NotNil(t, wantColor, "theme.TextStrong is nil")

	gotR, gotG, gotB, _ := gotColor.RGBA()
	wantR, wantG, wantB, _ := wantColor.RGBA()

	assert.Equal(t, wantR, gotR, "buttonFocused red mismatch")
	assert.Equal(t, wantG, gotG, "buttonFocused green mismatch")
	assert.Equal(t, wantB, gotB, "buttonFocused blue mismatch")
}
