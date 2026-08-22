// Package ui implements the main tea model of the application.
package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/core/keymap"
)

type UIModel struct {
	keyMap keymap.KeyMap

	// TODO: this is purely for debugging purposes, remove it eventually?
	debugRender tea.Msg
}

func New() (*UIModel, error) {
	keyMap := keymap.New()

	model := &UIModel{
		keyMap: keyMap,
	}

	return model, nil
}

func (model *UIModel) Init() tea.Cmd {
	var command tea.Cmd
	return command
}

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

func (model *UIModel) View() tea.View {
	return tea.NewView(fmt.Sprintf("%s", model.debugRender))
}

func (model *UIModel) handleKeyPressMsg(message tea.KeyPressMsg) tea.Cmd {
	var commands []tea.Cmd

	if key.Matches(message, model.keyMap.Quit) {
		commands = append(commands, tea.Quit)
	}

	return tea.Batch(commands...)
}
