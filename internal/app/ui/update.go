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
	// Initialization Step 1: Config loaded.
	case initializationLoadConfigResultMessage:
		commands = append(commands, model.handleInitializationLoadConfigResult(message))
	// Initialization Step 2: Load library cache (if any).
	case initializationLoadLibraryCacheResultMessage:
		commands = append(commands, model.handleInitializationLoadLibraryCacheResult(message))
	// User Action - Scan files Step 1: starting with fetching them.
	case fetchFilesStartMessage:
		model.PushNotification("starting file scan...")
		commands = append(commands, model.handleFetchFilesStartMessage(message))
	// User Action - Scan files Step 2: waiting fetching to finish.
	case fetchFilesWaitProgressMessage:
		commands = append(
			commands,
			fetchFilesWaitProgressCommand(
				message.progressChannel,
				message.resultChannel,
				message.generation,
			),
		)
	// User Action - Scan files Step 3: start scanning the fetched files metatag.
	case fetchFilesResultMessage:
		commands = append(commands, model.handleFetchFilesResultMessage(message))
	// Notification expiration routine: each push schedules its own expiry.
	case notificationExpireMessage:
		model.notificationModel.PruneExpired()
	case tea.WindowSizeMsg:
		if command := model.handleWindowResize(message); command != nil {
			commands = append(commands, command)
		}
	case tea.KeyPressMsg:
		if command := model.handleKeyPressMsg(message); command != nil {
			commands = append(commands, command)
		}
	}

	// Notification expiration: drain expire commands queued, if any.
	commands = append(commands, model.drainNotificationCmds()...)
	return model, tea.Batch(commands...)
}

// cancelCurrentFileScan aborts the current scan, if any, so its goroutine exits and stops feeding the drainer.
func (model *UIModel) cancelCurrentFileScan() {
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
		model.cancelCurrentFileScan()
		return tea.Quit
	}

	if model.state == uiInitializing {
		return model.handleComponentAction(model.initializationModel.HandleMessage(message))
	}

	if key.Matches(message, model.keyMap.OpenActions) || model.whichkeyModel.IsVisible() {
		return model.handleComponentAction(model.whichkeyModel.HandleMessage(message))
	}

	if model.state == uiLibraryStats {
		return model.handleComponentAction(model.libraryStatsModel.HandleMessage(message))
	}

	return nil
}

// handleComponentAction dispatches an action returned by a component.
func (model *UIModel) handleComponentAction(act action.Action) tea.Cmd {
	switch act := act.(type) {
	case nil, action.NoAction:
		return nil
	case action.QuitAction:
		model.cancelCurrentFileScan()
		return tea.Quit
	case action.ScanLibraryFullAction:
		model.cancelCurrentFileScan()

		model.library.scanGeneration++
		return fetchFilesStartCommand(
			model.orchestratorContext,
			model.library.scanGeneration,
			model.config.LibrariesPaths,
		)
	case action.ReloadConfigAction:
		model.cancelCurrentFileScan()
		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)

		return initializationLoadConfigCommand()
	case action.ProceedFromInitAction:
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)
		model.cancelCurrentFileScan()
		model.setState(uiPlaylist)

		return nil
	case action.ActionCommand:
		return act.Command
	case action.OpenPlaylistAction:
		model.setState(uiPlaylist)
		return nil
	case action.OpenLibraryStatsAction:
		model.setState(uiLibraryStats)
		return nil
	default:
		return nil
	}
}
