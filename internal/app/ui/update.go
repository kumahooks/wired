package ui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// initializationLoadConfigResultMessage is produced by initializationLoadConfigCommand when config.Load completes.
// TODO: find an elegant way to store these outside of update.
type initializationLoadConfigResultMessage struct {
	config           *config.Config
	isConfigDefaults bool
	err              error
}

// initializationLoadConfigCommand returns a tea.Cmd that loads config from disk into a fresh *Config.
func initializationLoadConfigCommand() tea.Cmd {
	return func() tea.Msg {
		loadedConfig, isConfigDefaults, err := config.Load()
		return initializationLoadConfigResultMessage{config: loadedConfig, isConfigDefaults: isConfigDefaults, err: err}
	}
}

// initializationCountFilesStartMessage is produced by initializationCountFilesStartCommand right after the config is
// loaded and libraries exist. It carries the channels and cancel func that the drainer and the cancel path share.
type initializationCountFilesStartMessage struct {
	progressChannel <-chan int
	resultChannel   <-chan initializationCountFilesResultMessage
	countCancel     context.CancelFunc
	generation      uint64
}

type initializationCountFilesResultMessage struct {
	filesCount int
	err        error
	generation uint64
}

// initializationCountFilesStartCommand launches the count goroutine and returns a StartMessage with the channels. The
// count context is derived from the orchestrator context so an orchestrator shutdown cancels the count. The channel
// is buffered so the walk does not stall on the tea message round-trip.
func initializationCountFilesStartCommand(
	orchestratorContext context.Context,
	generation uint64,
	filePaths []string,
) tea.Cmd {
	return func() tea.Msg {
		progressChannel := make(chan int, countFilesProgressChannelBuffer)
		resultChannel := make(chan initializationCountFilesResultMessage, 1)

		countContext, countCancel := context.WithCancel(orchestratorContext)

		go func() {
			filesCount, err := audio.CountFiles(countContext, filePaths, nil, progressChannel)
			close(progressChannel)

			resultChannel <- initializationCountFilesResultMessage{
				filesCount: filesCount,
				err:        err,
				generation: generation,
			}
		}()

		return initializationCountFilesStartMessage{
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			countCancel:     countCancel,
			generation:      generation,
		}
	}
}

type initializationCountFilesWaitProgressMessage struct {
	filesCount      int
	progressChannel <-chan int
	resultChannel   <-chan initializationCountFilesResultMessage
	generation      uint64
}

func initializationCountFilesWaitProgressCommand(
	progressChannel <-chan int,
	resultChannel <-chan initializationCountFilesResultMessage,
	generation uint64,
) tea.Cmd {
	return func() tea.Msg {
		filesCount, ok := <-progressChannel
		if !ok {
			result := <-resultChannel

			return initializationCountFilesResultMessage{
				filesCount: result.filesCount,
				err:        result.err,
				generation: result.generation,
			}
		}

		return initializationCountFilesWaitProgressMessage{
			filesCount:      filesCount,
			progressChannel: progressChannel,
			resultChannel:   resultChannel,
			generation:      generation,
		}
	}
}

func (model *UIModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch message := message.(type) {
	case initializationLoadConfigResultMessage:
		commands = append(commands, model.handleInitializationLoadConfigResult(message))
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

// handleInitializationLoadConfigResult processes a initializationLoadConfigResultMessage.
func (model *UIModel) handleInitializationLoadConfigResult(message initializationLoadConfigResultMessage) tea.Cmd {
	if message.err != nil {
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)

		return nil
	}

	if message.isConfigDefaults {
		model.initializationModel.AppendLog("no config file found, loading one using defaults", initializing.LogNormal)
	}

	// Publish the fresh config through the shared pointer so the orchestrator and the UI see the same values.
	*model.config = *message.config
	model.initializationModel.AppendLog("config loaded successfully", initializing.LogNormal)

	// Parses and apply the config's loaded theme.
	model.initializationModel.AppendLog("loading themes...", initializing.LogNormal)
	model.theme = theme.New(model.config.Theme)
	model.initializationModel.ApplyTheme(model.theme)
	model.initializationModel.AppendLog("theme loaded successfully", initializing.LogNormal)

	// Parses and apply the config's loaded keybinds.
	model.initializationModel.AppendLog("resolving keybindings...", initializing.LogNormal)
	resolvedKeyMap, err := keymap.New(model.config.Keybinds)
	if err != nil {
		model.initializationModel.AppendLog(err.Error(), initializing.LogError)
		model.initializationModel.AppendLog("falling back to default keybindings", initializing.LogError)
	} else {
		model.keyMap = resolvedKeyMap
		model.initializationModel.AppendLog("keybindings loaded successfully", initializing.LogNormal)
	}
	model.initializationModel.ApplyKeyMap(model.keyMap)

	// If there is no library path to scan, we tell the user and let him decide what to do.
	if len(model.config.LibrariesPaths) == 0 {
		model.initializationModel.AppendLog("no library paths found", initializing.LogError)
		return nil
	}

	// Immediately starts scanning files in the library paths.
	// TODO: this will not be the case in the future, we will skip loading if there's cache, and also ask before any scan.
	model.countGeneration++
	return initializationCountFilesStartCommand(
		model.orchestratorContext,
		model.countGeneration,
		model.config.LibrariesPaths,
	)
}

// handleInitializationCountFilesStartMessage stores the count's cancel func so reload/quit can abort the current count,
// seeds the live counter at zero, and launches the first drainer command.
func (model *UIModel) handleInitializationCountFilesStartMessage(message initializationCountFilesStartMessage) tea.Cmd {
	model.initializationModel.AppendLog("counting total library files", initializing.LogNormal)

	model.cancelInitializationCount = message.countCancel
	model.countGeneration = message.generation
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
	if message.generation == model.countGeneration {
		model.cancelInitializationCount = nil
	}

	if message.err != nil {
		model.initializationModel.AppendLog(message.err.Error(), initializing.LogError)
		return nil
	}

	// Only the current count updates the UI.
	if message.generation != model.countGeneration {
		return nil
	}

	// Since the counter has finished, we don't render the progress message anymore.
	model.initializationModel.SetCountFilesProgress(-1)
	model.initializationModel.AppendLog(
		fmt.Sprintf("counted a total of %d audio files successfully", message.filesCount),
		initializing.LogNormal,
	)

	// TODO: next step after this would be metatag scanning.

	return nil
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

// cancelInFlightInitializationCount aborts the current count, if any, so its goroutine exits and stops feeding the drainer.
func (model *UIModel) cancelInFlightInitializationCount() {
	if model.cancelInitializationCount != nil {
		model.cancelInitializationCount()
		model.cancelInitializationCount = nil
	}
}

// handleComponentAction dispatches an action returned by a component.
func (model *UIModel) handleComponentAction(act action.Action) tea.Cmd {
	switch act := act.(type) {
	case nil, action.NoAction:
		return nil
	case action.QuitAction:
		model.cancelInFlightInitializationCount()
		return tea.Quit
	case action.ReloadConfigAction:
		model.cancelInFlightInitializationCount()
		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)
		model.setState(uiInitializing)

		return initializationLoadConfigCommand()
	case action.ProceedFromInitAction:
		model.cancelInFlightInitializationCount()
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)
		model.setState(uiIdle)

		return nil
	case action.ActionCommand:
		return act.Command
	default:
		return nil
	}
}
