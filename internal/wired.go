// Package wired defines the main orchestrator of the application and its primitives.
package wired

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// WiredOrchestrator holds the whole application state. It is kinda like "game state": config, library, playlist, UI, and
// other domain states live here.
type WiredOrchestrator struct {
	// these are used for propagating cancels to the UIModel's contexts.
	cancelContext  context.Context
	cancelFunction context.CancelFunc

	// teaProgram is a reference to the tea program, a result from `tea.NewProgram`.
	teaProgram *tea.Program

	// uiModel is the tea's UI, which has it's own lifecycle methods e.g. Init, Update, View.
	uiModel *ui.UIModel

	// config is the loaded application's config, either the defaults or the users. the initialization pipeline updates
	// this to the user's if it's a valid config schema. this is shared with uiModel so UI components can change config.
	config *config.Config

	// library holds the reference to the user's loaded audio files, and is shared with uiModel.
	library *[]audio.File
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
	orchestrator.library = &[]audio.File{}

	// UI's model is created as per bubbletea pattern. It includes a reference to every UI component in the application.
	// The orchestrator is passed down in order for background operations the UI spawns to be canceled elegantly.
	uiModel, err := ui.New(orchestrator.cancelContext, defaultKeyMap, orchestrator.config, orchestrator.library)
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
