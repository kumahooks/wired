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
