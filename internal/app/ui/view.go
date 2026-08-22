package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (model *UIModel) View() tea.View {
	var render string

	if model.debugRender != "" {
		render += fmt.Sprint(model.debugRender)
	}

	return tea.NewView(render)
}
