package ui

import (
	"slices"

	bubbletea "github.com/charmbracelet/bubbletea"
)

func (model Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	// TODO: support for bubbletea.WindowSizeMsg
	switch msg := msg.(type) {
	case ConfigLoadedMsg:
		if msg.Err != nil {
			model.Error = msg.Err
			return model, nil
		}

		model.Config = msg.Config
		return model, nil

	case bubbletea.KeyMsg:
		messageStr := msg.String()

		if model.Config == nil {
			// TODO: if there's no config here, something went wrong. Maybe we should draw on the screen?
			if messageStr == "ctrl+c" {
				return model, bubbletea.Quit
			}

			return model, nil
		}

		keybinds := model.Config.Keybinds

		if slices.Contains(keybinds.Quit, messageStr) {
			return model, bubbletea.Quit
		}

		// TODO: further keybinds functionality
	}

	return model, nil
}
