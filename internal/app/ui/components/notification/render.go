package notification

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Render draws up to maxRenderedNotifications cards, stacked vertically and aligned right. Each card wraps its text to
// half the window width and clips its content to maxContentLines lines.
func (model *Model) Render(windowWidth int, windowHeight int) string {
	maxCardWidth := max(windowWidth/2, 1) - cardBorderWidth
	notifications := make([]string, 0, maxRenderedNotifications)

	for _, notification := range model.notifications {
		if len(notifications) >= maxRenderedNotifications {
			break
		}

		notifications = append(notifications, model.renderCard(notification.message, maxCardWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Right, notifications...)
}

// renderCard wraps a message to maxWidth, clips it to maxContentLines, and draws it with the card's rounded border.
func (model *Model) renderCard(message string, maxCardWidth int) string {
	wrappedMessage := lipgloss.Wrap(message, maxCardWidth, "")

	lines := strings.Split(wrappedMessage, "\n")
	if len(lines) > maxContentLines {
		lines = lines[:maxContentLines]
	}

	body := model.style.content.Render(strings.Join(lines, "\n"))
	return model.style.card.MaxWidth(maxCardWidth + cardBorderWidth).Render(body)
}
