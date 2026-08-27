package ui

import (
	"strings"

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

	base := model.baseView()
	return model.composeOverlays(base)
}

// baseView renders the active state's view, without any overlay on top of it.
func (model *UIModel) baseView() string {
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
		strings.Repeat(strings.Repeat("program loaded successfully. idle~", 10)+"\n", 44),
	)
}
