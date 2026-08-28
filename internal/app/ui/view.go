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
	case uiPlaylist:
		return model.playlistView()
	case uiLibraryStats:
		return model.libraryStatsView()
	default:
		return ""
	}
}

// playlistView centers the playlist message on the terminal. Temporary shit.
// TODO: this should be its own component I think
func (model *UIModel) playlistView() string {
	return lipgloss.Place(
		model.windowWidth,
		model.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		strings.Repeat(strings.Repeat("this is the playlist view ", 10)+"\n", 44),
	)
}

// libraryStatsView centers the library stats message on the terminal. Temporary shit.
// TODO: this should be its own component I think
func (model *UIModel) libraryStatsView() string {
	return lipgloss.Place(
		model.windowWidth,
		model.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		strings.Repeat(strings.Repeat("this is the library stats view ", 10)+"\n", 44),
	)
}
