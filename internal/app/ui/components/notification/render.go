package notification

import (
	"charm.land/lipgloss/v2"
)

// Render draws up to maxRenderedNotifications cards.
func (model *Model) Render(windowWidth int, windowHeight int) string {
	notifications := make([]string, 0, maxRenderedNotifications)

	for _, notification := range model.notifications {
		if len(notifications) >= maxRenderedNotifications {
			break
		}

		// TODO: we need to border each notification, and only take as many height as needed for its content
		card := model.style.card.Width(10).MaxWidth(16).Height(6).MaxHeight(8).Render(notification.message)
		notifications = append(notifications, card)
	}

	return lipgloss.JoinVertical(lipgloss.Right, notifications...)
}
