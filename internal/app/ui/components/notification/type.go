package notification

import "time"

type Notification struct {
	message   string
	expiresAt time.Time
}

type Model struct {
	// notifications hold a FIFO queue of every pushed notification.
	notifications []Notification

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}
