package notification

import "time"

// Lifetime is how long a notification stays visible after being pushed.
const Lifetime = 4 * time.Second

// Layout constants for the notification dialog card.
const (
	// maxContentLines is the maximum number of wrapped text lines a card shows.
	maxContentLines = 3

	// cardBorderWidth is the total horizontal space the card's border occupies.
	cardBorderWidth = 2

	// maxRenderedNotifications simply represents the maximum amount of stacked vertical cards at once.
	maxRenderedNotifications = 4

	// zIndex is the card's z-order in the overlay compositor.
	zIndex = 30
)
