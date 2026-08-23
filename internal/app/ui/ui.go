// Package ui implements the main tea model of the application.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	// windowHeight and windowWidth are actual view space excluding borders.
	windowHeight int
	windowWidth  int

	// state decides what state the view is, essentially separating between initialization and idle.
	state uiState

	// keymaps are either the default shortcuts, or the ones the user configured for each action there is.
	keyMap keymap.KeyMap

	// config is the shared config pointer.
	config *config.Config

	// Components models
	initializationModel *initializing.Model
}

// New initializes the UIModel with the default keymap and a config pointer.
func New(defaultKeyMap keymap.KeyMap, config *config.Config) (*UIModel, error) {
	model := &UIModel{
		state:               uiInitializing,
		keyMap:              defaultKeyMap,
		config:              config,
		initializationModel: initializing.New(),
	}

	return model, nil
}

// Init sends a tea.Cmd message to load the user's config.
func (model *UIModel) Init() tea.Cmd {
	return loadConfigCmd()
}

func (model *UIModel) setState(state uiState) {
	model.state = state
}
