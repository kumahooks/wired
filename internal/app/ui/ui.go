// Package ui implements the main tea model of the application.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/core/keymap"
)

// The view is drawn based on these states below.
type uiState uint8

const (
	uiInitializing uiState = iota
	uiIdle
)

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	state  uiState
	keyMap keymap.KeyMap

	// TODO: this is purely for debugging purposes, remove it eventually?
	debugRender tea.Msg
}

// New Initializes the UIModel.
func New(keyMap keymap.KeyMap) (*UIModel, error) {
	model := &UIModel{
		state:  uiInitializing,
		keyMap: keyMap,
	}

	return model, nil
}

func (model *UIModel) Init() tea.Cmd {
	var command tea.Cmd
	return command
}
