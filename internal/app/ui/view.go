package ui

import (
	"fmt"

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
	case uiBootstrapping:
		return ""
	case uiInitializing:
		return model.initializationModel.Render(model.windowWidth, model.windowHeight)
	case uiLibrary:
		return model.mockedTODOScreen("Library UI")
	case uiPlaylist:
		return model.mockedTODOScreen("Playlist UI")
	case uiLibraryStats:
		return model.libraryStatsModel.Render(model.windowWidth, model.windowHeight)
	default:
		return ""
	}
}

func (model *UIModel) mockedTODOScreen(screen string) string {
	return lipgloss.Place(
		model.windowWidth,
		model.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		fmt.Sprintf("TODO: this screen %s is mocked", screen),
	)
}
