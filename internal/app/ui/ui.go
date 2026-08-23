// Package ui implements the main tea model of the application.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	// windowHeight and windowWidth are actual view space excluding borders.
	windowHeight int
	windowWidth  int

	// state decides what state the view is, essentially separating between initialization and idle.
	state uiState

	// keymaps are the shortcuts (default or user's) for the application's actions.
	keyMap keymap.KeyMap

	// theme is the resolved (default or user's) color palette shared with every component.
	theme theme.Theme

	// config is the shared config pointer.
	config *config.Config

	// Components models.
	initializationModel *initializing.Model
}

// New initializes the UIModel, which is basically the UI orchestrator of the application.
// It initializes the state to `uiInitializing`. To avoid locking the user out of actions, default configs, keymaps, and
// styles are loaded at first.
func New(defaultKeyMap keymap.KeyMap, config *config.Config) (*UIModel, error) {
	model := &UIModel{
		state:               uiInitializing,
		keyMap:              defaultKeyMap,
		theme:               theme.Default(),
		config:              config,
		initializationModel: initializing.New(defaultKeyMap),
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
