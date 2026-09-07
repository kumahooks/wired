package dialog

import "time"

// KeyGraceQuietPeriod is how long input must be quiet after the dialog opens before it starts handling keys.
const KeyGraceQuietPeriod = 150 * time.Millisecond

// Layout constants for the confirm dialog card.
const (
	// cardWidth is the content width the card's text wraps to, excluding padding and border.
	cardWidth = 48

	// cardHorizontalPadding is the blank columns the card keeps on each side of its content.
	cardHorizontalPadding = 2

	// buttonGap is the blank columns between the confirm and cancel buttons.
	buttonGap = 2

	// zIndex is the dialog's z-order in the overlay compositor.
	zIndex = 40
)
