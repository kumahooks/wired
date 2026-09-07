package initializing

// moveCursor shifts the selected button by delta, wrapping around the buttons visible in the current mode.
func (model *Model) moveCursor(delta int) {
	if len(model.buttons) == 0 {
		return
	}

	visible := model.visibleActions()
	if len(visible) == 0 {
		return
	}

	currentPosition := model.positionOfVisible(model.buttons[model.cursorPosition].action)
	nextPosition := (currentPosition + delta) % len(visible)
	if nextPosition < 0 {
		nextPosition += len(visible)
	}

	model.cursorPosition = model.canonicalIndexForVisible(visible[nextPosition])
}

// moveLogScroll shifts the log viewport by delta lines, clamping between the oldest line and the tail.
func (model *Model) moveLogScroll(delta int) {
	maxOffset := model.maxLogScrollOffset()

	nextOffset := model.scrollOffset + delta
	if nextOffset < 0 {
		nextOffset = 0
	}
	if nextOffset > maxOffset {
		nextOffset = maxOffset
	}

	model.scrollOffset = nextOffset
}

// moveLogScrollToHead jumps the log viewport to the oldest log lines.
func (model *Model) moveLogScrollToHead() {
	model.scrollOffset = model.maxLogScrollOffset()
}

// moveLogScrollToTail jumps the log viewport back to the newest log lines.
func (model *Model) moveLogScrollToTail() {
	model.scrollOffset = 0
}

// maxLogScrollOffset returns the largest valid scroll offset, which is zero when the buffer fits the viewport.
func (model *Model) maxLogScrollOffset() int {
	maxOffset := len(model.logLines) - maxVisibleLogLines
	if maxOffset < 0 {
		maxOffset = 0
	}

	return maxOffset
}
