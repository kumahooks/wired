package ui

import (
	bubbletea "github.com/charmbracelet/bubbletea"
	"wired/internal/config"
)

type Model struct {
	Config *config.Config
	Error  error
	// TODO: parse hex colors from config into lipgloss.Color types
}

func New() Model {
	return Model{}
}

func (model Model) Init() bubbletea.Cmd {
	return bubbletea.Batch(
		bubbletea.SetWindowTitle("wire(d)"),
		func() bubbletea.Msg {
			cfg, err := config.Load()
			return ConfigLoadedMsg{Config: cfg, Err: err}
		},
	)
}
