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

// handleInitializationLoadConfigResult processes a initializationLoadConfigResultMessage.
func (model *UIModel) handleInitializationLoadConfigResult(message initializationLoadConfigResultMessage) tea.Cmd {
	if message.err != nil {
		model.showInitializationScreen()
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)
		model.initializationModel.SetConfigError()

		return nil
	}

	if message.isConfigDefaults {
		model.initializationModel.AppendLog("no config file found, loading one using defaults~", initializing.LogNormal)
	}

	// Publish the loaded config through the shared pointer so the orchestrator and the UI see the same values.
	*model.config = *message.config
	model.libraryStatsModel.SetLibraryPaths(model.config.LibrariesPaths)
	model.initializationModel.AppendLog("config loaded successfully~", initializing.LogNormal)

	// Parses and apply the config's loaded theme.
	model.initializationModel.AppendLog("loading themes...", initializing.LogNormal)

	model.theme = theme.New(model.config.Theme)
	model.initializationModel.ApplyTheme(model.theme)
	model.libraryStatsModel.ApplyTheme(model.theme)
	model.whichkeyModel.ApplyTheme(model.theme)
	model.notificationModel.ApplyTheme(model.theme)

	model.initializationModel.AppendLog("theme loaded successfully~", initializing.LogNormal)

	// Parses and apply the config's loaded keybinds.
	model.initializationModel.AppendLog("resolving keybindings...", initializing.LogNormal)
	resolvedKeyMap, err := keymap.New(model.config.Keybinds)
	if err != nil {
		model.showInitializationScreen()

		model.initializationModel.AppendLog(err.Error(), initializing.LogError)
		model.initializationModel.AppendLog("falling back to default keybindings...", initializing.LogError)

		model.initializationModel.ApplyKeyMap(model.keyMap)
		model.whichkeyModel.SetBindings(model.commandBindingsFor(model.state))
		model.whichkeyModel.ApplyCloseKeybinding(model.keyMap.GoBack)

		model.initializationModel.AppendLog(
			"keybindings failed to load, fallbacking to previous bindings",
			initializing.LogError,
		)

		model.initializationModel.SetConfigError()

		return nil
	}
	model.keyMap = resolvedKeyMap
	model.initializationModel.ApplyKeyMap(model.keyMap)
	model.libraryStatsModel.ApplyKeyMap(model.keyMap)
	model.whichkeyModel.SetBindings(model.commandBindingsFor(model.state))
	model.whichkeyModel.ApplyCloseKeybinding(model.keyMap.GoBack)
	model.initializationModel.AppendLog("keybindings loaded successfully~", initializing.LogNormal)

	// If the user has any invalid path in it's config, we notify it as a log warning in the initialization component.
	if len(message.invalidLibraryPaths) > 0 {
		var pluralSuffix string = ""
		if len(message.invalidLibraryPaths) > 1 {
			pluralSuffix = "s"
		}

		model.initializationModel.AppendLog(
			fmt.Sprintf(
				"invalid path%s found (╥﹏╥): %s",
				pluralSuffix,
				strings.Join(message.invalidLibraryPaths, ", "),
			),
			initializing.LogWarning,
		)
	}

	// Once we load configs, we attempt to load any library cache.
	return initializationLoadLibraryCacheCommand()
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
	}

	model.showInitializationScreen()
	model.initializationModel.SetConfigError()

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

	parseContext, parseCancel := context.WithCancel(model.orchestratorContext)
	model.libraryDiscoveryCancel = parseCancel

	// The parse goroutine works on a snapshot of the discovered snapshotFiles and never touches the library's maps.
	snapshotFiles := make([]*audio.AudioFile, 0, len(message.library.File))
	for _, audioFile := range message.library.File {
		snapshotFiles = append(snapshotFiles, audioFile)
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
	err := audio.WriteCache(model.library.File)
	if err != nil {
		model.PushNotification("there was an error when trying to save the library to cache x_x")
	}

	model.libraryDiscoveryCancel = nil
	model.libraryStatsModel.FinishDiscovery()

	model.PushNotification(
		fmt.Sprintf("%d files have been discovered and parsed successfully", model.library.FilesCount()),
	)

	return nil
}
