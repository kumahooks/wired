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
	if len(message.library.filePaths) > 0 {
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

// handleInitializationCountFilesStartMessage stores the count's cancel func so reload/quit can abort the current count,
// seeds the live counter at zero, and launches the first drainer command.
func (model *UIModel) handleInitializationCountFilesStartMessage(message initializationCountFilesStartMessage) tea.Cmd {
	model.library.countingCancel = message.countCancel
	model.library.countingGeneration = message.generation
	model.initializationModel.SetCountFilesProgress(0)

	return initializationCountFilesWaitProgressCommand(
		message.progressChannel,
		message.resultChannel,
		message.generation,
	)
}

// handleInitializationCountFilesResultMessage finalizes the count.
func (model *UIModel) handleInitializationCountFilesResultMessage(
	message initializationCountFilesResultMessage,
) tea.Cmd {
	if message.generation == model.library.countingGeneration {
		model.library.countingCancel = nil
	}

	if message.err != nil {
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)
		return nil
	}

	// Only the current count updates the UI.
	if message.generation != model.library.countingGeneration {
		return nil
	}

	// Since the counter has finished, we don't render the progress message anymore.
	model.initializationModel.SetCountFilesProgress(-1)
	model.initializationModel.AppendLog(
		fmt.Sprintf("a total of %d audio files have been found~", message.filesCount),
		initializing.LogNormal,
	)

	// TODO: next step after this would be metatag scanning.

	return nil
}
