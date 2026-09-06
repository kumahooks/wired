package librarystats

// moveCursor shifts the selected button by delta.
func (model *Model) moveCursor(delta int) {
	if len(model.buttons) == 0 {
		return
	}

	nextPosition := (model.cursorPosition + delta) % len(model.buttons)
	if nextPosition < 0 {
		nextPosition += len(model.buttons)
	}

	model.cursorPosition = nextPosition
}
