package whichkey

import (
	"charm.land/lipgloss/v2"
)

// Anchor wraps the rendered card in a lipgloss Layer anchored to the bottom of the window.
func Anchor(content string, windowWidth int, windowHeight int) *lipgloss.Layer {
	_, contentHeight := lipgloss.Size(content)

	return lipgloss.NewLayer(content).
		X(0).
		Y(windowHeight - contentHeight).
		Z(zIndex)
}
