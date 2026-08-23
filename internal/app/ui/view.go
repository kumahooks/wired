package ui

import (
	"charm.land/bubbletea/v2"
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
		return model.initializationModel.Render()
	case uiInitializingLibraryChoice:
		return model.awaitingLibraryChoiceView()
	case uiIdle:
		return "program loaded successfully. idle~"
	default:
		return ""
	}
}

// awaitingLibraryChoiceView renders the "no libraries" prompt.
// TODO: this will be a better dialog eventually.
func (model *UIModel) awaitingLibraryChoiceView() string {
	var lines []string

	lines = append(lines, model.initializationModel.Render())
	lines = append(lines, "")
	lines = append(lines, "no library paths found in config.")
	lines = append(lines, "")
	lines = append(lines, "  [r] reload config (edit your config file and press r)")
	lines = append(lines, "  [p] proceed without libraries")

	return joinLines(lines)
}

func joinLines(lines []string) string {
	var result string
	for _, line := range lines {
		result += line + "\n"
	}

	return result
}
