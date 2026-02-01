package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"wired/internal/ui/header"
	"wired/internal/ui/notification"
)

func (model Model) View() string {
	// header (1 line) + footer (1 line) = 2 lines reserved
	contentHeight := max(model.height-2, 0)

	var base string

	if model.Dialog.Visible() {
		base = model.Dialog.View()
	} else if model.Config == nil {
		base = ""
	} else if model.Modal.Visible() {
		base = model.Modal.View()
	} else {
		base = model.viewForActivePanel()
	}

	// Pad base to exactly contentHeight lines
	baseLines := strings.Split(base, "\n")
	for len(baseLines) < contentHeight {
		baseLines = append(baseLines, "")
	}

	if len(baseLines) > contentHeight {
		baseLines = baseLines[:contentHeight]
	}

	base = strings.Join(baseLines, "\n")

	if model.Config != nil {
		visibleNotifications := model.Notifications.Visible(model.Config.Notification.NotificationShownMax)

		if len(visibleNotifications) > 0 {
			notifications := model.renderNotifications(visibleNotifications)
			base = overlayBottomRight(base, notifications, model.width, contentHeight)
		}
	}

	return model.Header.View() + "\n" + base + "\n" + model.Footer.View()
}

func (model Model) viewForActivePanel() string {
	switch model.Header.Active() {
	case header.Library:
		return "Library..."
	case header.Playlist:
		return "Playlist..."
	case header.Statistics:
		return "Statistics..."
	default:
		return "OwO Undefined"
	}
}

func (model Model) renderNotifications(notifications []notification.Notification) string {
	bubbles := make([]string, 0, len(notifications))

	for _, n := range notifications {
		bubble := model.Notifications.Render(n)
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
