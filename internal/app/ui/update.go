package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/core/config"
	"wired/internal/core/keymap"
)

// configLoadedMsg is produced by loadConfigCmd when config.Load completes.
type configLoadedMsg struct {
	config *config.Config
	err    error
}

// loadConfigCmd returns a tea.Cmd that loads config from disk into a fresh *Config, and produces a configLoadedMsg carrying it.
func loadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		loaded, err := config.Load()
		return configLoadedMsg{config: loaded, err: err}
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

// handleConfigLoaded processes a configLoadedMsg. On success it swaps the fresh config pointer in, rebuilds the keymap
// from it, and transitions to uiIdle when libraries are configured or uiInitializingLibraryChoice when none are.
func (model *UIModel) handleConfigLoaded(msg configLoadedMsg) tea.Cmd {
	if msg.err != nil {
		model.initializationModel.AppendLog(msg.err.Error())
		return nil
	}

	// Publish the fresh config through the shared pointer so the orchestrator and the UI see the same values.
	*model.config = *msg.config

	model.initializationModel.AppendLog("config loaded successfully")
	model.keyMap = keymap.New(model.config.Keybinds)

	if len(model.config.LibrariesPaths) == 0 {
		model.setState(uiInitializingLibraryChoice)
		return nil
	}

	model.initializationModel.AppendLog("initialization complete")
	model.setState(uiIdle)

	return nil
}

// handleKeyPressMsg is responsible for managing every keyboard action in the program.
func (model *UIModel) handleKeyPressMsg(message tea.KeyPressMsg) tea.Cmd {
	var commands []tea.Cmd

	switch model.state {
	case uiInitializingLibraryChoice:
		return model.handleInitializingLibraryChoice(message)
	default:
		if key.Matches(message, model.keyMap.Quit) {
			commands = append(commands, tea.Quit)
		}
	}

	return tea.Batch(commands...)
}

// handleInitializingLibraryChoice handles keypresses on the "no libraries" prompt.
// TODO: this will be a better dialog eventually.
func (model *UIModel) handleInitializingLibraryChoice(message tea.KeyPressMsg) tea.Cmd {
	switch message.String() {
	case "r":
		model.setState(uiInitializing)
		model.initializationModel.AppendLog("reloading config...")

		return loadConfigCmd()
	case "p":
		model.initializationModel.AppendLog("proceeding without libraries")
		model.setState(uiIdle)

		return nil
	}

	return nil
}

func (model *UIModel) handleWindowResize(message tea.WindowSizeMsg) tea.Cmd {
	model.windowHeight = message.Height
	model.windowWidth = message.Width

	return nil
}
