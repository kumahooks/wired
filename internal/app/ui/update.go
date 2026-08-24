package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
)

func (model *UIModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch message := message.(type) {
	case initializationLoadConfigResultMessage:
		commands = append(commands, model.handleInitializationLoadConfigResult(message))
	case initializationLoadLibraryCacheResultMessage:
		commands = append(commands, model.handleInitializationLoadLibraryCacheResult(message))
	case initializationCountFilesStartMessage:
		commands = append(commands, model.handleInitializationCountFilesStartMessage(message))
	case initializationCountFilesWaitProgressMessage:
		model.initializationModel.SetCountFilesProgress(message.filesCount)

		commands = append(
			commands,
			initializationCountFilesWaitProgressCommand(
				message.progressChannel,
				message.resultChannel,
				message.generation,
			),
		)
	case initializationCountFilesResultMessage:
		commands = append(commands, model.handleInitializationCountFilesResultMessage(message))
	case tea.WindowSizeMsg:
		if command := model.handleWindowResize(message); command != nil {
			commands = append(commands, command)
		}
	case tea.KeyPressMsg:
		if command := model.handleKeyPressMsg(message); command != nil {
			commands = append(commands, command)
		}
	}

	return model, tea.Batch(commands...)
}

// cancelInFlightInitializationCount aborts the current count, if any, so its goroutine exits and stops feeding the drainer.
func (model *UIModel) cancelInFlightInitializationCount() {
	if model.library.countingCancel != nil {
		model.library.countingCancel()
		model.library.countingCancel = nil
	}
}

func (model *UIModel) handleWindowResize(message tea.WindowSizeMsg) tea.Cmd {
	model.windowHeight = message.Height
	model.windowWidth = message.Width

	return nil
}

// handleKeyPressMsg is responsible for managing every keyboard action in the program. Quit is a global keybind.
func (model *UIModel) handleKeyPressMsg(message tea.KeyPressMsg) tea.Cmd {
	if key.Matches(message, model.keyMap.Quit) {
		model.cancelInFlightInitializationCount()
		return tea.Quit
	}

	if model.state == uiInitializing {
		return model.handleComponentAction(model.initializationModel.HandleMessage(message))
	}

	return nil
}

// handleComponentAction dispatches an action returned by a component.
func (model *UIModel) handleComponentAction(act action.Action) tea.Cmd {
	switch act := act.(type) {
	case nil, action.NoAction:
		return nil
	case action.QuitAction:
		model.cancelInFlightInitializationCount()
		return tea.Quit
	case action.ScanLibraryFullAction:
		model.initializationModel.AppendLog("scanning library...", initializing.LogNormal)

		model.cancelInFlightInitializationCount()
		model.library.countingGeneration++
		return initializationCountFilesStartCommand(
			model.orchestratorContext,
			model.library.countingGeneration,
			model.config.LibrariesPaths,
		)
	case action.ReloadConfigAction:
		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)

		model.cancelInFlightInitializationCount()
		return initializationLoadConfigCommand()
	case action.ProceedFromInitAction:
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)

		model.cancelInFlightInitializationCount()
		model.setState(uiIdle)
		return nil
	case action.ActionCommand:
		return act.Command
	default:
		return nil
	}
}
