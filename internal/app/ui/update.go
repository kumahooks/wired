package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (model *UIModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch message := message.(type) {
	case tea.KeyPressMsg:
		if cmd := model.handleKeyPressMsg(message); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	return model, tea.Batch(commands...)
}

// handleKeyPressMsg is responsible for managing every keyboard action in the program.
func (model *UIModel) handleKeyPressMsg(message tea.KeyPressMsg) tea.Cmd {
	var commands []tea.Cmd

	if key.Matches(message, model.keyMap.Quit) {
		commands = append(commands, tea.Quit)
	} else {
		model.debugRender = fmt.Sprintf("Pressed: %v", message)
	}

	return tea.Batch(commands...)
}
