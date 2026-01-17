package ui

import (
	"fmt"
	"strings"
)

func (model Model) View() string {
	// TODO: create custom component to show a message like this in the center of the screen
	if model.Error != nil {
		return fmt.Sprintf("Error: %v", model.Error)
	}

	if model.Config == nil {
		// TODO: create custom component to show a message like this in the center of the screen
		return "Loading configuration..."
	}

	// TODO: draw ui
	quitKeys := strings.Join(model.Config.Keybinds.Quit, ", ")
	return fmt.Sprintf("Welcome to %s\nPress [%s] to quit.", model.Config.Title, quitKeys)
}
