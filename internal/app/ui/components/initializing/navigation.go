package initializing

// moveCursor shifts the selected button by delta, wrapping around the row.
func (model *Model) moveCursor(delta int) {
	if len(model.buttons) == 0 {
		return
	}

	model.cursorPosition = (model.cursorPosition + delta + len(model.buttons)) % len(model.buttons)
}
