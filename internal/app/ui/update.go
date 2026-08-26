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
	// Initialization Step 1: Config loaded
	case initializationLoadConfigResultMessage:
		commands = append(commands, model.handleInitializationLoadConfigResult(message))
	// Initialization Step 2: Load library cache (if any)
	case initializationLoadLibraryCacheResultMessage:
		commands = append(commands, model.handleInitializationLoadLibraryCacheResult(message))
	// Initialization (optional) Step 3.1: Scan files, starting with fetching them
	case initializationFetchFilesStartMessage:
		commands = append(commands, model.handleInitializationFetchFilesStartMessage(message))
	// Initialization (optional) Step 3.2: Fetching files
	case initializationFetchFilesWaitProgressMessage:
		model.initializationModel.SetFetchFilesProgress(message.filesCount)

		commands = append(
			commands,
			initializationFetchFilesWaitProgressCommand(
				message.progressChannel,
				message.resultChannel,
				message.generation,
			),
		)
	// Initialization (optional) Step 4: After fetching files, we parse their metatag
	case initializationFetchFilesResultMessage:
		commands = append(commands, model.handleInitializationFetchFilesResultMessage(message))
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

// cancelInFlightInitializationScan aborts the current scan, if any, so its goroutine exits and stops feeding the drainer.
func (model *UIModel) cancelInFlightInitializationScan() {
	if model.library.scanCancel != nil {
		model.library.scanCancel()
		model.library.scanCancel = nil
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
		model.cancelInFlightInitializationScan()
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
		model.cancelInFlightInitializationScan()
		return tea.Quit
	case action.ScanLibraryFullAction:
		model.initializationModel.AppendLog("scanning library...", initializing.LogNormal)

		model.cancelInFlightInitializationScan()
		model.library.scanGeneration++
		return initializationFetchFilesStartCommand(
			model.orchestratorContext,
			model.library.scanGeneration,
			model.config.LibrariesPaths,
		)
	case action.ReloadConfigAction:
		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)

		model.cancelInFlightInitializationScan()
		return initializationLoadConfigCommand()
	case action.ProceedFromInitAction:
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)

		model.cancelInFlightInitializationScan()
		model.setState(uiIdle)
		return nil
	case action.ActionCommand:
		return act.Command
	default:
		return nil
	}
}
