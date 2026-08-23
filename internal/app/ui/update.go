package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// configLoadedMsg is produced by loadConfigCmd when config.Load completes.
// TODO: find an elegant way to store these outside of update.
type configLoadedMsg struct {
	config           *config.Config
	isConfigDefaults bool
	err              error
}

// loadConfigCmd returns a tea.Cmd that loads config from disk into a fresh *Config, producing a configLoadedMsg with it.
func loadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		loaded, isConfigDefaults, err := config.Load()

		return configLoadedMsg{config: loaded, isConfigDefaults: isConfigDefaults, err: err}
	}
}

func (model *UIModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch message := message.(type) {
	case configLoadedMsg:
		commands = append(commands, model.handleConfigLoaded(message))
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

// handleConfigLoaded processes a configLoadedMsg.
// TODO: this will eventually be used elsewhere after the initialization...
func (model *UIModel) handleConfigLoaded(message configLoadedMsg) tea.Cmd {
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

	model.theme = theme.New(model.config.Theme)
	model.initializationModel.ApplyTheme(model.theme)
	model.initializationModel.AppendLog("theme loaded successfully", initializing.LogNormal)

	resolvedKeyMap, err := keymap.New(model.config.Keybinds)
	if err != nil {
		model.initializationModel.AppendLog(err.Error(), initializing.LogError)
		model.initializationModel.AppendLog("falling back to default keymaps", initializing.LogError)
	} else {
		model.keyMap = resolvedKeyMap
		model.initializationModel.AppendLog("keybindings loaded successfully", initializing.LogNormal)
	}
	model.initializationModel.ApplyKeyMap(model.keyMap)

	if len(model.config.LibrariesPaths) == 0 {
		model.initializationModel.AppendLog("no library paths found", initializing.LogError)

		return nil
	}

	model.initializationModel.AppendLog("initialization complete", initializing.LogNormal)
	model.setState(uiIdle)

	return nil
}

// handleKeyPressMsg is responsible for managing every keyboard action in the program. Quit is a global keybind.
func (model *UIModel) handleKeyPressMsg(message tea.KeyPressMsg) tea.Cmd {
	if key.Matches(message, model.keyMap.Quit) {
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
		return tea.Quit
	case action.ReloadConfigAction:
		model.initializationModel.AppendLog("reloading config...", initializing.LogNormal)
		model.setState(uiInitializing)

		return loadConfigCmd()
	case action.ProceedFromInitAction:
		model.initializationModel.AppendLog("proceeding without libraries", initializing.LogNormal)
		model.setState(uiIdle)

		return nil
	case action.ActionCommand:
		return act.Command
	default:
		return nil
	}
}

func (model *UIModel) handleWindowResize(message tea.WindowSizeMsg) tea.Cmd {
	model.windowHeight = message.Height
	model.windowWidth = message.Width

	return nil
}
