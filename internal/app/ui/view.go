package ui

import (
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (model *UIModel) View() tea.View {
	view := tea.NewView(model.viewContent())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeNone
	view.WindowTitle = model.config.Title

	return view
}

func (model *UIModel) viewContent() string {
	if model.windowWidth < minWindowWidth || model.windowHeight < minWindowHeight {
		return "terminal size is too small"
	}

	switch model.state {
	case uiInitializing:
		return model.initializationModel.Render(model.windowWidth, model.windowHeight)
	case uiIdle:
		return model.idleView()
	default:
		return ""
	}
}

// idleView centers the idle message on the terminal. Temporary shit.
func (model *UIModel) idleView() string {
	return lipgloss.Place(
		model.windowWidth,
		model.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		"program loaded successfully. idle~",
	)
}
