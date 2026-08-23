package ui

import (
	tea "charm.land/bubbletea/v2"
)

func (model *UIModel) View() tea.View {
	view := tea.NewView(model.viewContent())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeNone
	view.WindowTitle = model.windowTitle

	return view
}

func (model *UIModel) viewContent() string {
	var render string

	if model.windowWidth < minWindowWidth || model.windowHeight < minWindowHeight {
		return "terminal size is too small"
	}

	return render
}
