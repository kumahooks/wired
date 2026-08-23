// Package wired defines the main orchestrator of the application and its primitives.
package wired

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// WiredOrchestrator holds the whole application state. It is kinda like "game state": config, library, playlist, and
// other domain states live here.
type WiredOrchestrator struct {
	ctx        context.Context
	cancel     context.CancelFunc
	teaProgram *tea.Program
	uiModel    *ui.UIModel
	config     *config.Config
}

// New initializes the WiredOrchestrator structure.
func New(ctx context.Context) (*WiredOrchestrator, error) {
	orchestrator := &WiredOrchestrator{}

	configDefaults := config.Defaults()
	defaultKeyMap := keymap.New(configDefaults.Keybinds)
	orchestrator.config = &configDefaults

	uiModel, err := ui.New(defaultKeyMap, orchestrator.config)
	if err != nil {
		return nil, err
	}
	orchestrator.uiModel = uiModel

	orchestrator.ctx, orchestrator.cancel = context.WithCancel(ctx)

	return orchestrator, nil
}

// Run creates and runs the bubbletea terminal program based on the initialized UIModel.
func (orchestrator *WiredOrchestrator) Run() (tea.Model, error) {
	orchestrator.teaProgram = tea.NewProgram(orchestrator.uiModel, tea.WithContext(orchestrator.ctx))

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
	if orchestrator.cancel != nil {
		orchestrator.cancel()
	}
}
