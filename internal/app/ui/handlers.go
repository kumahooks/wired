package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/audio"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// showInitializationScreen flips the UI from the silent bootstrapping state to the initialization screen.
func (model *UIModel) showInitializationScreen() {
	if model.state == uiBootstrapping {
		model.setState(uiInitializing)
	}
}

// handleConfigLoadedMessage processes a configLoadedMessage. Feedback and failure branch on the origin: initialization
// logs into its screen buffer and chains the library cache load, while user actions push notifications only.
func (model *UIModel) handleConfigLoadedMessage(message configLoadedMessage) tea.Cmd {
	// feedback branches a notification log on the type of `message.origin`.
	var feedback func(string, initializing.LogType)

	if message.origin == configLoadOriginInit {
		feedback = func(text string, level initializing.LogType) {
			model.initializationModel.AppendLog(text, level)
		}
	} else {
		feedback = func(text string, level initializing.LogType) {
			switch level {
			case initializing.LogError, initializing.LogWarning:
				model.PushNotification(text)
			}
		}
	}

	if message.err != nil {
		feedback(message.err.Error(), initializing.LogError)

		if message.origin == configLoadOriginInit {
			model.showInitializationScreen()
			model.initializationModel.SetConfigError()
		}

		return nil
	}

	if message.isConfigDefaults {
		feedback("no config file found, loading one using defaults~", initializing.LogNormal)
	}

	// Publish the loaded config through the shared pointer so the orchestrator and the UI see the same values.
	*model.config = *message.config
	model.libraryStatsModel.SetLibraryPaths(model.config.LibrariesPaths)
	feedback("config loaded successfully~", initializing.LogNormal)

	// Load and apply the config's loaded theme.
	feedback("loading themes...", initializing.LogNormal)
	model.theme = theme.New(model.config.Theme)
	model.applyThemeToComponents()
	feedback("theme loaded successfully~", initializing.LogNormal)

	// Parse and apply the config's loaded keybinds.
	feedback("resolving keybindings...", initializing.LogNormal)
	resolvedKeyMap, keymapErr := keymap.New(model.config.Keybinds)
	if keymapErr != nil {
		return model.handleConfigLoadedKeymapError(message.origin, feedback, keymapErr)
	}
	model.keyMap = resolvedKeyMap
	model.applyKeyMapToComponents()
	feedback("keybindings loaded successfully~", initializing.LogNormal)

	// If the user has any invalid path in its config, report it.
	if len(message.invalidLibraryPaths) > 0 {
		pluralSuffix := ""
		if len(message.invalidLibraryPaths) > 1 {
			pluralSuffix = "s"
		}

		feedback(
			fmt.Sprintf(
				"invalid path%s found (╥﹏╥): %s",
				pluralSuffix,
				strings.Join(message.invalidLibraryPaths, ", "),
			),
			initializing.LogWarning,
		)
	}

	// Only the initialization pipeline continues into the library cache load.
	if message.origin == configLoadOriginInit {
		return initializationLoadLibraryCacheCommand()
	}

	// For a user config reload push notification for the success.
	model.PushNotification("config loaded successfully~")

	return nil
}

// applyThemeToComponents rebuilds every component's style from the UIModel's current theme.
func (model *UIModel) applyThemeToComponents() {
	model.initializationModel.ApplyTheme(model.theme)
	model.libraryStatsModel.ApplyTheme(model.theme)
	model.whichkeyModel.ApplyTheme(model.theme)
	model.notificationModel.ApplyTheme(model.theme)
	model.dialogModel.ApplyTheme(model.theme)
}

// handleConfigLoadedKeymapError handles keymap.New failing during a config load. Initialization logs the failure and
// marks a config error for the user to fix. In an error scenarioo, the previously resolved keymap is kept in both cases.
func (model *UIModel) handleConfigLoadedKeymapError(
	origin configLoadOrigin,
	feedback func(string, initializing.LogType),
	keymapErr error,
) tea.Cmd {
	if origin == configLoadOriginInit {
		model.showInitializationScreen()

		feedback(keymapErr.Error(), initializing.LogError)
		feedback("keybindings failed to load, fallbacking to previous bindings x_x", initializing.LogError)

		model.initializationModel.SetConfigError()
		return nil
	}

	feedback(fmt.Sprintf("keybindings failed to load: %s", keymapErr.Error()), initializing.LogError)
	feedback("keeping previous keybindings x_x", initializing.LogError)

	return nil
}

// applyKeyMapToComponents pushes the UIModel's current keymap to every component that renders or matches keys.
func (model *UIModel) applyKeyMapToComponents() {
	model.initializationModel.ApplyKeyMap(model.keyMap)
	model.libraryStatsModel.ApplyKeyMap(model.keyMap)
	model.dialogModel.ApplyKeyMap(model.keyMap)
	model.whichkeyModel.SetBindings(model.commandBindingsFor(model.state))
	model.whichkeyModel.ApplyCloseKeybinding(model.keyMap.GoBack)
}

// handleInitializationLoadLibraryCacheResult routes the user depending on whether a library cache exists. A cache hit
// moves straight to idle. A cache miss, when there are valid library paths, resolves the startup silently so the user
// is taken to the idle UI (they can discover files later on). If no valid library path exists, it is treated as a
// config error since there is nothing to discover, so the initialization screen is shown.
func (model *UIModel) handleInitializationLoadLibraryCacheResult(
	message initializationLoadLibraryCacheResultMessage,
) tea.Cmd {
	if message.err != nil {
		model.PushNotification("there was an error when trying to load the library cache... q_q")
	}

	if message.library.FilesCount() > 0 {
		model.library.File = message.library.File
		audio.BuildLibraryIndexes(model.library)
		model.libraryStatsModel.ComputeStats()

		model.setState(uiPlaylist)
		return nil
	}

	if len(model.config.LibrariesPaths) > 0 {
		model.initializationModel.AppendLog(
			"no songs found, you might want to discover them later~",
			initializing.LogWarning,
		)
	} else {
		model.initializationModel.AppendLog("no library paths found ;_;", initializing.LogError)
		model.initializationModel.SetConfigError()
	}

	model.showInitializationScreen()

	return nil
}

// handleDiscoverFilesStartMessage stores the discovery's cancel func so reload/quit can abort the current discovery,
// seeds the live counter at zero, and starts both the progress tick chain and the result waiter.
func (model *UIModel) handleDiscoverFilesStartMessage(message discoverFilesStartMessage) tea.Cmd {
	model.libraryDiscoveryCancel = message.discoveryCancel
	model.libraryDiscoveryGeneration = message.generation
	model.libraryStatsModel.StartDiscovery()

	return tea.Batch(
		discoveryProgressTickCommand(message.progress, message.generation),
		waitForDiscoverResultCommand(message.result),
	)
}

// handleDiscoveryProgressTickMessage pulls the discovery progress reporter into the library stats view and schedules
// the next tick. Ticks from old tick chains or a finished discovery are dropped, ending that chain.
func (model *UIModel) handleDiscoveryProgressTickMessage(message discoveryProgressTickMessage) tea.Cmd {
	if message.generation != model.libraryDiscoveryGeneration || model.libraryDiscoveryCancel == nil {
		return nil
	}

	model.libraryStatsModel.SetProgress(message.progress)

	return discoveryProgressTickCommand(message.progress, message.generation)
}

// handleDiscoverFilesResultMessage finalizes the file discovery routine and hands the discovered library over to the
// metatag parsing.
func (model *UIModel) handleDiscoverFilesResultMessage(message discoverFilesResultMessage) tea.Cmd {
	if message.generation != model.libraryDiscoveryGeneration {
		return nil
	}

	if message.err != nil {
		model.libraryDiscoveryCancel = nil
		model.libraryStatsModel.FinishDiscovery()

		return nil
	}

	if message.progress.DiscoveredCount() == 0 {
		if message.onlyNew {
			model.PushNotification("no new files have been found~")
		}

		model.libraryDiscoveryCancel = nil
		model.libraryStatsModel.FinishDiscovery()

		return nil
	}

	parseContext, parseCancel := context.WithCancel(model.orchestratorContext)
	model.libraryDiscoveryCancel = parseCancel

	// The parse goroutine works on a snapshot of the library's files and never touches the library's maps.
	var snapshotFiles []*audio.AudioFile
	if message.onlyNew {
		snapshotFiles = message.library.UntaggedFiles()
	} else {
		snapshotFiles = make([]*audio.AudioFile, 0, len(message.library.File))
		for _, audioFile := range message.library.File {
			snapshotFiles = append(snapshotFiles, audioFile)
		}
	}

	message.progress.SetDiscoveryDone()

	return parseFilesMetatagStartCommand(
		parseContext,
		model.libraryDiscoveryGeneration,
		snapshotFiles,
		message.progress,
	)
}

// handleMetatagParseStartMessage arms the progress tick chain and the result waiter under the parse's generation,
// which the discovery result handler already registered alongside the parse's cancel func.
func (model *UIModel) handleMetatagParseStartMessage(message metatagParseStartMessage) tea.Cmd {
	if message.generation != model.libraryDiscoveryGeneration {
		return nil
	}

	model.libraryStatsModel.SetProgress(message.progress)

	return tea.Batch(
		discoveryProgressTickCommand(message.progress, message.generation),
		waitForMetatagResultCommand(message.result),
	)
}

// handleMetatagParseResultMessage finalizes the metatag parse and hands the library over to the index rebuild.
func (model *UIModel) handleMetatagParseResultMessage(message metatagParseResultMessage) tea.Cmd {
	if message.generation != model.libraryDiscoveryGeneration {
		return nil
	}

	if message.err != nil {
		model.libraryDiscoveryCancel = nil
		model.libraryStatsModel.FinishDiscovery()

		return nil
	}

	audio.BuildLibraryIndexes(model.library)
	model.libraryStatsModel.ComputeStats()

	err := audio.WriteCache(model.library.File)
	if err != nil {
		model.PushNotification("there was an error when trying to save the library to cache x_x")
	}

	model.libraryDiscoveryCancel = nil
	model.libraryStatsModel.FinishDiscovery()

	model.PushNotification(
		fmt.Sprintf("%d files have been discovered and parsed successfully", message.parsedCount),
	)

	return nil
}
