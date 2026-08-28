package notification

import "time"

// Lifetime is how long a notification stays visible after being pushed.
const Lifetime = 4 * time.Second

const maxRenderedNotifications = 4

// Layout constants for the notification dialog card.
const (
	// zIndex is the card's z-order in the overlay compositor.
	zIndex = 30
)
