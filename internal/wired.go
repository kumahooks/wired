// Package wired defines the main orchestrator of the application and its primitives.
package wired

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// WiredOrchestrator holds the whole application state. It is kinda like "game state": config, library, playlist, UI, and
// other domain states live here.
type WiredOrchestrator struct {
	cancelContext  context.Context
	cancelFunction context.CancelFunc
	teaProgram     *tea.Program
	uiModel        *ui.UIModel
	config         *config.Config
}

// New initializes the WiredOrchestrator structure.
func New(ctx context.Context) (*WiredOrchestrator, error) {
	orchestrator := &WiredOrchestrator{}

	// We store a cancel context mainly just to shutdown cleanly.
	orchestrator.cancelContext, orchestrator.cancelFunction = context.WithCancel(ctx)

	// Loads the default configs and keymaps, so the user is not left stuck in case their custom config is broken.
	configDefaults := config.Defaults()
	defaultKeyMap, err := keymap.New(configDefaults.Keybinds)
	if err != nil {
		return nil, fmt.Errorf("[wired:New] build default keymap: %w", err)
	}
	orchestrator.config = &configDefaults

	// UI's model is created as per bubbletea pattern. It includes a reference to every UI component in the application.
	// The orchestrator is passed down in order for background operations the UI spawns to be canceled elegantly.
	uiModel, err := ui.New(orchestrator.cancelContext, defaultKeyMap, orchestrator.config)
	if err != nil {
		return nil, err
	}
	orchestrator.uiModel = uiModel

	return orchestrator, nil
}

// Run creates and runs the bubbletea terminal program based on the initialized UIModel.
func (orchestrator *WiredOrchestrator) Run() (tea.Model, error) {
	orchestrator.teaProgram = tea.NewProgram(orchestrator.uiModel, tea.WithContext(orchestrator.cancelContext))

	model, err := orchestrator.teaProgram.Run()
	return model, err
}

// Config returns the currently loaded config.
func (orchestrator *WiredOrchestrator) Config() *config.Config {
	return orchestrator.config
}

// NotifyTea sends to the already initialized bubbletea program a tea.Msg.
func (orchestrator *WiredOrchestrator) NotifyTea(message tea.Msg) {
	if orchestrator.teaProgram == nil {
		return
	}

	orchestrator.teaProgram.Send(message)
}

// Shutdown cancels the orchestrator context and stops the tea program.
func (orchestrator *WiredOrchestrator) Shutdown() {
	if orchestrator.cancelFunction != nil {
		orchestrator.cancelFunction()
	}
}
