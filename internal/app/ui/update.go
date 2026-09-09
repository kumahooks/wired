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
	case configLoadedMessage:
		model.isConfigLoading = false
		commands = append(commands, model.handleConfigLoadedMessage(message))
	// Initialization Step 2: Load library cache (if any).
	case initializationLoadLibraryCacheResultMessage:
		commands = append(commands, model.handleInitializationLoadLibraryCacheResult(message))

	// Library discovery:
	// Step 1: discovering audio files at the library paths.
	case discoverFilesStartMessage:
		commands = append(commands, model.handleDiscoverFilesStartMessage(message))
	// Step 2.1: start parsing the discovered files' metatags.
	case discoverFilesResultMessage:
		commands = append(commands, model.handleDiscoverFilesResultMessage(message))
	// Step 2.2: metatag parsing running.
	case metatagParseStartMessage:
		commands = append(commands, model.handleMetatagParseStartMessage(message))
	// Step 2.3: file discovery, parsing and indexing is finished.
	case metatagParseResultMessage:
		commands = append(commands, model.handleMetatagParseResultMessage(message))
	// Step (generic): ticking discovery progress.
	case discoveryProgressTickMessage:
		commands = append(commands, model.handleDiscoveryProgressTickMessage(message))

	// Notification expiration routine: each push schedules its own expiry.
	case notificationExpireMessage:
		model.notificationModel.PruneExpired()

	// Tea commands are below.
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

// cancelCurrentLibraryDiscovery aborts the current library discovery, if any, so its goroutine exits and stops feeding the waiter.
func (model *UIModel) cancelCurrentLibraryDiscovery() {
	if model.libraryDiscoveryCancel != nil {
		model.libraryDiscoveryCancel()
		model.libraryDiscoveryCancel = nil
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
		model.cancelCurrentLibraryDiscovery()
		return tea.Quit
	}

	if model.state == uiBootstrapping {
		return nil
	}

	// The confirm dialog, while open, consumes every key so the underlying screen stays frozen.
	if model.dialogModel.IsOpen() {
		return model.handleComponentAction(model.dialogModel.HandleMessage(message))
	}

	// The init screen is "unskippable": user must click the "proceed" button to close it.
	if model.state == uiInitializing {
		return model.handleComponentAction(model.initializationModel.HandleMessage(message))
	}

	// The action (whichkey) catalogue takes priority over every "normal" action.
	if key.Matches(message, model.keyMap.OpenActions) || model.whichkeyModel.IsVisible() {
		return model.handleComponentAction(model.whichkeyModel.HandleMessage(message))
	}

	// Screen quick movements if configured.
	if key.Matches(message, model.keyMap.ViewLibrary) {
		return model.handleComponentAction(action.OpenLibraryAction{})
	}
	if key.Matches(message, model.keyMap.ViewPlaylist) {
		return model.handleComponentAction(action.OpenPlaylistAction{})
	}
	if key.Matches(message, model.keyMap.ViewLibraryStats) {
		return model.handleComponentAction(action.OpenLibraryStatsAction{})
	}

	if model.state == uiLibraryStats {
		return model.handleComponentAction(model.libraryStatsModel.HandleMessage(message))
	}

	return nil
}

// handleComponentAction dispatches an action returned by a component.
func (model *UIModel) handleComponentAction(act action.Action) tea.Cmd {
	switch act := act.(type) {
	// Librarystats screen actions
	case action.DiscoverLibraryFullAction:
		model.cancelCurrentLibraryDiscovery()

		// A full rediscovery starts from scratch, so the in-memory library is wiped before the walk.
		model.library.Reset()
		model.libraryDiscoveryGeneration++
		return discoverFilesStartCommand(
			model.orchestratorContext,
			model.libraryDiscoveryGeneration,
			model.config.LibrariesPaths,
			model.library,
			false,
		)
	case action.DiscoverLibraryNewAction:
		model.cancelCurrentLibraryDiscovery()

		model.libraryDiscoveryGeneration++
		return discoverFilesStartCommand(
			model.orchestratorContext,
			model.libraryDiscoveryGeneration,
			model.config.LibrariesPaths,
			model.library,
			true,
		)

	// Initialization screen actions
	case action.ReloadConfigFromInitAction:
		model.cancelCurrentLibraryDiscovery()

		if model.isConfigLoading {
			return nil
		}
		model.isConfigLoading = true

		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)

		return configLoadCommand(configLoadOriginInit)
	case action.ProceedFromInitAction:
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)
		model.cancelCurrentLibraryDiscovery()
		model.setState(uiPlaylist)
		return nil

	// UI movement actions
	case action.OpenLibraryAction:
		model.setState(uiLibrary)
		return nil
	case action.OpenPlaylistAction:
		model.setState(uiPlaylist)
		return nil
	case action.OpenLibraryStatsAction:
		model.setState(uiLibraryStats)
		return nil

	// Whichkey specific actions
	case action.ReloadConfigAction:
		model.cancelCurrentLibraryDiscovery()

		if model.isConfigLoading {
			return nil
		}
		model.isConfigLoading = true

		return configLoadCommand(configLoadOriginUser)

	// Miscellaneous
	case nil, action.NoAction:
		return nil
	case action.QuitAction:
		model.cancelCurrentLibraryDiscovery()
		return tea.Quit
	case action.OpenConfirmDialogAction:
		model.dialogModel.Open(act.Text, act.ConfirmLabel, act.CancelLabel, act.ConfirmAction)
		return nil
	case action.ActionCommand:
		return act.Command
	default:
		return nil
	}
}
