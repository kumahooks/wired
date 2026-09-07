package dialog

import (
	"charm.land/lipgloss/v2"
)

// Anchor wraps the rendered dialog in a lipgloss Layer centered on the window.
func Anchor(content string, windowWidth int, windowHeight int) *lipgloss.Layer {
	contentWidth, contentHeight := lipgloss.Size(content)

	x := max((windowWidth-contentWidth)/2, 0)
	y := max((windowHeight-contentHeight)/2, 0)

	return lipgloss.NewLayer(content).X(x).Y(y).Z(zIndex)
}
