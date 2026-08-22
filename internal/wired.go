// Package wired defines the main orchestrator of the application and its primitives.
package wired

import (
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui"
)

// WiredOrchestrator holds the whole application state.
type WiredOrchestrator struct {
	teaProgram             *tea.Program
	uiModel                *ui.UIModel
	initialized            bool
	initializationProgress int
}

// New initializes the WiredOrchestrator structure.
func New() (*WiredOrchestrator, error) {
	uiModel, err := ui.New()
	if err != nil {
		return nil, err
	}

	orchestrator := &WiredOrchestrator{
		uiModel:                uiModel,
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
