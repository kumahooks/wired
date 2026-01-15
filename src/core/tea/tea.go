// Package tea implements the core bubbletea's methods as Init, Update and View
package tea

import (
	core "wired/src/core"

	bubbletea "github.com/charmbracelet/bubbletea"
)

type TeaModel struct {
	CoreModel *core.CoreModel
}

func (model *TeaModel) Init() bubbletea.Cmd {
	return bubbletea.SetWindowTitle("wire")
}

func (model *TeaModel) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return model, bubbletea.Quit
		}
	}

	return model, nil
}

func (model *TeaModel) View() string {
	return ""
}
