// Package notification implements a FIFO queue of rendered notification dialogs.
package notification

import (
	"time"

	"wired/internal/core/theme"
)

func New() *Model {
	return &Model{
		notifications: []Notification{},
		style:         newStyle(theme.Default()),
	}
}

func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// PruneExpired drops the expired prefix of the FIFO queue.
func (model *Model) PruneExpired() {
	now := time.Now().UTC()
	expiredCount := 0

	for _, notification := range model.notifications {
		if now.Before(notification.expiresAt) {
			break
		}

		expiredCount++
	}

	model.notifications = model.notifications[expiredCount:]
}

func (model *Model) PushNotification(message string) {
	model.notifications = append(model.notifications, Notification{
		message:   message,
		expiresAt: time.Now().UTC().Add(Lifetime),
	})
}

func (model *Model) HasActiveNotifications() bool {
	return len(model.notifications) > 0
}
