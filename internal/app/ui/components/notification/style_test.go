package notification

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
		"card":    style.card,
		"content": style.content,
	}

	for name, gotStyle := range styles {
		assert.NotEqual(t, zeroStyle, gotStyle, "Style.%s equals the zero lipgloss.Style", name)
	}
}

func TestStyleColorsMatchTheme(t *testing.T) {
	t.Parallel()

	resolvedTheme := testutil.DefaultTheme()
	style := newStyle(resolvedTheme)

	cardColor := style.card.GetBorderTopForeground()
	require.NotNil(t, cardColor, "card border foreground is nil")

	wantColor := resolvedTheme.AccentInteractive
	require.NotNil(t, wantColor, "theme.AccentInteractive is nil")

	gotR, gotG, gotB, _ := cardColor.RGBA()
	wantR, wantG, wantB, _ := wantColor.RGBA()

	assert.Equal(t, wantR, gotR, "card border red mismatch")
	assert.Equal(t, wantG, gotG, "card border green mismatch")
	assert.Equal(t, wantB, gotB, "card border blue mismatch")

	contentColor := style.content.GetForeground()
	require.NotNil(t, contentColor, "content foreground is nil")

	wantContentColor := resolvedTheme.TextPrimary
	require.NotNil(t, wantContentColor, "theme.TextPrimary is nil")

	gotR, gotG, gotB, _ = contentColor.RGBA()
	wantR, wantG, wantB, _ = wantContentColor.RGBA()

	assert.Equal(t, wantR, gotR, "content red mismatch")
	assert.Equal(t, wantG, gotG, "content green mismatch")
	assert.Equal(t, wantB, gotB, "content blue mismatch")
}
