package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// handleInitializationLoadConfigResult processes a initializationLoadConfigResultMessage.
func (model *UIModel) handleInitializationLoadConfigResult(message initializationLoadConfigResultMessage) tea.Cmd {
	if message.err != nil {
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)
		model.initializationModel.SetConfigError()

		return nil
	}

	if message.isConfigDefaults {
		model.initializationModel.AppendLog("no config file found, loading one using defaults~", initializing.LogNormal)
	}

	// Publish the fresh config through the shared pointer so the orchestrator and the UI see the same values.
	*model.config = *message.config
	model.initializationModel.AppendLog("config loaded successfully~", initializing.LogNormal)

	// Parses and apply the config's loaded theme.
	model.initializationModel.AppendLog("loading themes...", initializing.LogNormal)
	model.theme = theme.New(model.config.Theme)
	model.initializationModel.ApplyTheme(model.theme)
	model.initializationModel.AppendLog("theme loaded successfully~", initializing.LogNormal)

	// Parses and apply the config's loaded keybinds.
	model.initializationModel.AppendLog("resolving keybindings...", initializing.LogNormal)
	resolvedKeyMap, err := keymap.New(model.config.Keybinds)
	if err != nil {
		model.initializationModel.AppendLog(err.Error(), initializing.LogError)
		model.initializationModel.AppendLog("falling back to default keybindings...", initializing.LogError)

		model.initializationModel.ApplyKeyMap(model.keyMap)
		model.initializationModel.AppendLog("keybindings loaded successfully~", initializing.LogNormal)

		model.initializationModel.SetConfigError()

		return nil
	}
	model.keyMap = resolvedKeyMap
	model.initializationModel.ApplyKeyMap(model.keyMap)
	model.initializationModel.AppendLog("keybindings loaded successfully~", initializing.LogNormal)

	// If the user has any invalid path in it's config, we notify it as a log warning, so they can fix it if they wish.
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
// moves straight to idle. A cache miss, when tere are valid libraries path, warns the user about there not being any,
// cache'd file, so they can scan later on. If no valid library path exists, it is treated as a config error since there
// is nothing to scan.
func (model *UIModel) handleInitializationLoadLibraryCacheResult(
	message initializationLoadLibraryCacheResultMessage,
) tea.Cmd {
	if len(*message.library.audioFiles) > 0 {
		model.setState(uiIdle)
		return nil
	}

	if len(model.config.LibrariesPaths) > 0 {
		model.initializationModel.AppendLog(
			"no scanned songs found, you might want to scan them later~",
			initializing.LogWarning,
		)

		return nil
	}

	model.initializationModel.AppendLog("no library paths found ;_;", initializing.LogError)
	model.initializationModel.SetConfigError()

	return nil
}

// handleFetchFilesStartMessage stores the scan's cancel func so reload/quit can abort the current scan, seeds the live
// counter at zero, and launches the first drainer command.
func (model *UIModel) handleFetchFilesStartMessage(message fetchFilesStartMessage) tea.Cmd {
	model.library.scanCancel = message.scanCancel
	model.library.scanGeneration = message.generation

	return fetchFilesWaitProgressCommand(
		message.progressChannel,
		message.resultChannel,
		message.generation,
	)
}

// handleFetchFilesResultMessage finalizes the fetch.
func (model *UIModel) handleFetchFilesResultMessage(message fetchFilesResultMessage) tea.Cmd {
	if message.generation == model.library.scanGeneration {
		model.library.scanCancel = nil
	}

	if message.err != nil {
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)
		return nil
	}

	// Only the current scan updates the UI.
	if message.generation != model.library.scanGeneration {
		return nil
	}

	*model.library.audioFiles = message.files

	return scanFilesMetatagStartCommand()
}
