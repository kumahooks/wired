// Package ui implements the main tea model of the application.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	// windowHeight and windowWidth are actual view space excluding borders.
	windowTitle  string
	windowHeight int
	windowWidth  int

	// state decides what state the view is, essentially separating between initialization and idle.
	state uiState

	// keymaps are either the default shortcuts, or the ones the user configured for each action there is.
	keyMap keymap.KeyMap
}

// New Initializes the UIModel.
func New(config config.Config, keyMap keymap.KeyMap) (*UIModel, error) {
	model := &UIModel{
		windowTitle: config.Title,
		state:       uiInitializing,
		keyMap:      keyMap,
	}

	return model, nil
}

func (model *UIModel) Init() tea.Cmd {
	var command tea.Cmd
	return command
}

func (model *UIModel) setState(state uiState) {
	model.state = state
}
