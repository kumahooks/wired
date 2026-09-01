package ui

import (
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"wired/internal/app/ui/components/librarystats"
)

func (model *UIModel) View() tea.View {
	view := tea.NewView(model.viewContent())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeNone
	view.WindowTitle = model.config.Title

	return view
}

func (model *UIModel) viewContent() string {
	// TODO: eventually every component must express their minimum sizes, I think... I wonder what is the best architecture here.
	if model.state == uiLibraryStats {
		if model.windowWidth < librarystats.MinWidth || model.windowHeight < minWindowHeight {
			return "terminal size is too small"
		}
	} else if model.windowWidth < minWindowWidth || model.windowHeight < minWindowHeight {
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
		"this is the playlist view~ TODO: actually implement this lol",
	)
}

// libraryStatsView renders the library stats screen through its component.
func (model *UIModel) libraryStatsView() string {
	return model.libraryStatsModel.Render(model.windowWidth, model.windowHeight)
}
