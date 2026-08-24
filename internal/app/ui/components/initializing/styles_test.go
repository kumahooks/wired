package initializing

import (
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"

	"wired/internal/core/theme"
)

func TestNewStyleAllFieldsNonZero(t *testing.T) {
	t.Parallel()

	style := newStyle(theme.Default())
	zeroStyle := lipgloss.NewStyle()

	styles := map[string]lipgloss.Style{
		"panel":         style.panel,
		"logArea":       style.logArea,
		"logError":      style.logError,
		"header":        style.header,
		"progress":      style.progress,
		"buttonFocused": style.buttonFocused,
		"buttonBlurred": style.buttonBlurred,
		"hint":          style.hint,
	}

	for name, gotStyle := range styles {
		if reflect.DeepEqual(gotStyle, zeroStyle) {
			t.Errorf("Style.%s equals the zero lipgloss.Style, want at least one modifier set", name)
		}
	}
}

func TestButtonFocusedForegroundMatchesTextStrong(t *testing.T) {
	t.Parallel()

	defaultTheme := theme.Default()
	style := newStyle(defaultTheme)

	gotColor := style.buttonFocused.GetForeground()
	if gotColor == nil {
		t.Fatal("buttonFocused foreground is nil")
	}

	wantColor := defaultTheme.TextStrong
	if wantColor == nil {
		t.Fatal("theme.TextStrong is nil")
	}

	gotR, gotG, gotB, gotA := gotColor.RGBA()
	wantR, wantG, wantB, wantA := wantColor.RGBA()

	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Errorf(
			"buttonFocused foreground RGBA = (%d, %d, %d, %d), want theme.TextStrong (%d, %d, %d, %d)",
			gotR, gotG, gotB, gotA, wantR, wantG, wantB, wantA,
		)
	}
}

