package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"wired/internal/config"
	"wired/internal/ui/notification"
)

func (model Model) View() string {
	var base string

	if model.Config == nil {
		// TODO: maybe msg bubble with the msg?
		base = "Loading configuration..."
	} else if model.Modal.Visible() {
		base = model.Modal.View(model.Config)
	} else if model.Error != nil {
		// TODO: maybe msg bubble with the error?
		base = fmt.Sprintf("Error: %v", model.Error)
	} else {
		// TODO: player view
		quitKeys := strings.Join(model.Config.Keybinds.Quit, ", ")
		base = fmt.Sprintf("Welcome to %s\nPress [%s] to quit.", model.Config.Title, quitKeys)
	}

	if model.Config != nil {
		visibleNotifications := model.Notifications.Visible(model.Config.Notification.NotificationShownMax)

		if len(visibleNotifications) > 0 {
			notifications := renderNotifications(visibleNotifications, model.Config)
			return overlayBottomRight(base, notifications, model.width, model.height)
		}
	}

	return base
}

func renderNotifications(notifications []notification.Notification, cfg *config.Config) string {
	bubbles := make([]string, 0, len(notifications))

	for _, n := range notifications {
		bubble := notification.Render(n, cfg)
		bubbles = append(bubbles, bubble)
	}

	return lipgloss.JoinVertical(lipgloss.Right, bubbles...)
}

func overlayBottomRight(base string, overlay string, width int, height int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	if len(baseLines) > height {
		baseLines = baseLines[:height]
	}

	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := len(overlayLines)
	if overlayHeight > height {
		overlayLines = overlayLines[overlayHeight-height:]
		overlayHeight = height
	}

	startRow := height - overlayHeight
	startCol := max(width-overlayWidth, 0)

	// Merge overlay onto base, preserving base content to the left
	for i, overlayLine := range overlayLines {
		rowIdx := startRow + i
		if rowIdx < 0 || rowIdx >= height {
			continue
		}

		baseLine := baseLines[rowIdx]
		baseVisualWidth := lipgloss.Width(baseLine)

		if baseVisualWidth <= startCol {
			// Base line is shorter than where overlay starts, just pad and append
			padding := strings.Repeat(" ", startCol-baseVisualWidth)
			baseLines[rowIdx] = baseLine + padding + overlayLine
		} else {
			truncated := ansi.Truncate(baseLine, startCol, "")

			truncatedWidth := lipgloss.Width(truncated)
			paddingLength := max(startCol-truncatedWidth, 0)
			padding := strings.Repeat(" ", paddingLength)

			baseLines[rowIdx] = truncated + padding + overlayLine
		}
	}

	return strings.Join(baseLines, "\n")
}
