// Package wired defines the main orchestrator of the application and its primitives.
package wired

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// WiredOrchestrator holds the whole application state.
type WiredOrchestrator struct {
	teaProgram             *tea.Program
	uiModel                *ui.UIModel
	config                 *config.Config
	initialized            bool
	initializationProgress int
}

// New initializes the WiredOrchestrator structure.
func New() (*WiredOrchestrator, error) {
	// TODO: we need to catch these errors. They are not a problem.
	// Two types of errors:
	// 1. File is unreadable, we must then show the user the errors and offer the option to reload or rewrite to defaults;
	// 2. Library Path is empty/invalid, which in this case we will allow the user to input their libraries through the UI.
	configData, _ := config.Load()

	keyMaps := keymap.New(configData.Keybinds)
	uiModel, err := ui.New(keyMaps)
	if err != nil {
		return nil, err
	}

	orchestrator := &WiredOrchestrator{
		uiModel:                uiModel,
		config:                 configData,
		initialized:            true,
		initializationProgress: 100,
	}

	return orchestrator, nil
}

// Run creates and run the bubbletea's terminal program based on the initialized UIModel.
func (orchestrator *WiredOrchestrator) Run() (tea.Model, error) {
	orchestrator.teaProgram = tea.NewProgram(orchestrator.uiModel)

	model, err := orchestrator.teaProgram.Run()
	return model, err
}

// NotifyTea sends to the already initialized bubbletea program a `tea.Msg`.
func (orchestrator *WiredOrchestrator) NotifyTea(message tea.Msg) {
	orchestrator.teaProgram.Send(message)
}
