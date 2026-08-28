package notification

import (
	"charm.land/lipgloss/v2"
)

// Anchor wraps the rendered notifications in a lipgloss Layer anchored to the bottom-right of the window.
func Anchor(content string, windowWidth int, windowHeight int) *lipgloss.Layer {
	contentWidth, contentHeight := lipgloss.Size(content)

	return lipgloss.NewLayer(content).
		X(windowWidth - contentWidth).
		Y(windowHeight - contentHeight).
		Z(zIndex)
}
