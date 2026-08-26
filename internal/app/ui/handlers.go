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

	return initializationLoadLibraryCacheCommand()
}

// handleInitializationLoadLibraryCacheResult routes the user depending on whether a library cache exists. A cache hit
// moves straight to idle. A cache miss offers a full scan when library paths are configured, otherwise it is treated as
// a config error since there is nothing to scan.
func (model *UIModel) handleInitializationLoadLibraryCacheResult(
	message initializationLoadLibraryCacheResultMessage,
) tea.Cmd {
	if len(message.library.audioFiles) > 0 {
		model.setState(uiIdle)
		return nil
	}

	if len(model.config.LibrariesPaths) > 0 {
		model.initializationModel.AppendLog(
			"no scanned songs found, do you want to scan now?",
			initializing.LogWarning,
		)

		model.initializationModel.SetEmptyLibrary()
		return nil
	}

	model.initializationModel.AppendLog("no library paths found ;_;", initializing.LogError)
	model.initializationModel.SetConfigError()

	return nil
}

// handleInitializationFetchFilesStartMessage stores the scan's cancel func so reload/quit can abort the current scan,
// seeds the live counter at zero, and launches the first drainer command.
func (model *UIModel) handleInitializationFetchFilesStartMessage(message initializationFetchFilesStartMessage) tea.Cmd {
	model.library.scanCancel = message.scanCancel
	model.library.scanGeneration = message.generation
	model.initializationModel.SetFetchFilesProgress(0)

	return initializationFetchFilesWaitProgressCommand(
		message.progressChannel,
		message.resultChannel,
		message.generation,
	)
}

// handleInitializationFetchFilesResultMessage finalizes the fetch. It takes ownership of the file slice shipped through
// the message so no shared mutable state remains with the walk goroutine.
func (model *UIModel) handleInitializationFetchFilesResultMessage(
	message initializationFetchFilesResultMessage,
) tea.Cmd {
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

	model.library.audioFiles = message.files

	// Since the fetch has finished, we don't render the progress message anymore.
	model.initializationModel.SetFetchFilesProgress(-1)
	model.initializationModel.AppendLog(
		fmt.Sprintf("a total of %d audio files have been found~", len(message.files)),
		initializing.LogNormal,
	)

	return initializationScanFilesMetatagStartCommand()
}
