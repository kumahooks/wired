// Package ui implements the main tea model of the application.
package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/core/keymap"
)

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	keyMap keymap.KeyMap

	// TODO: this is purely for debugging purposes, remove it eventually?
	debugRender tea.Msg
}

// New Initializes the UIModel.
func New(keyMap keymap.KeyMap) (*UIModel, error) {
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
	var render string

	if model.debugRender != "" {
		render += fmt.Sprint(model.debugRender)
	}

	return tea.NewView(render)
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
