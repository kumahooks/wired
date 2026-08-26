package initializing

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Render returns the full-screen init view with the panel (log area and buttons) centered in it.
func (model *Model) Render(width int, height int) string {
	panel := model.buildPanel()

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, panel)
}

// buildPanel assembles the initialization panel, rendering title, log area, progress line, button row, and the hint.
func (model *Model) buildPanel() string {
	sections := []string{
		model.style.header.Render("wire(d) is starting..."),
		model.renderLogArea(),
	}

	// The progress line only exists while files are being fetched.
	if model.fetchFilesProgress >= 0 {
		sections = append(sections, model.renderProgressLine())
	}

	sections = append(
		sections,
		model.renderButtonRow(),
		model.style.hint.Render(model.renderHint()),
	)

	content := strings.Join(sections, "\n\n")

	return model.style.panel.Render(content)
}

// renderLogArea builds a fixed-height log area of visibleLogLines lines.
func (model *Model) renderLogArea() string {
	rows := model.visibleLogRows()

	if maxVisibleLogLines-len(rows) > 0 {
		for range maxVisibleLogLines - len(rows) {
			rows = append(rows, model.style.logNormal.Render(""))
		}
	}

	body := strings.Join(rows, "\n")
	return model.style.logArea.Width(logAreaWidth).Height(maxVisibleLogLines).Render(body)
}

// visibleLogRows returns the rendered log lines that fit in the fixed log area (at most maxVisibleLogLines).
func (model *Model) visibleLogRows() []string {
	startIndex := 0
	if len(model.logLines) > maxVisibleLogLines {
		startIndex = len(model.logLines) - maxVisibleLogLines
	}

	visibleLines := model.logLines[startIndex:]
	logRows := make([]string, 0, len(visibleLines))

	for _, entry := range visibleLines {
		switch entry.logType {
		case LogError:
			logRows = append(logRows, model.style.logError.Render(entry.text))
		case LogWarning:
			logRows = append(logRows, model.style.logWarning.Render(entry.text))
		default:
			logRows = append(logRows, model.style.logNormal.Render(entry.text))
		}
	}

	return logRows
}

// renderProgressLine shows the live count while files are being fetched.
func (model *Model) renderProgressLine() string {
	return model.style.progress.Render(fmt.Sprintf("fetching %d audio files...", model.fetchFilesProgress))
}

// renderButtonRow renders the buttons visible in the current mode horizontally with spacing between them.
func (model *Model) renderButtonRow() string {
	visibleActions := model.visibleActions()
	buttonsRendering := make([]string, 0, len(visibleActions)*2)

	for position, action := range visibleActions {
		buttonIndex := model.canonicalIndexForVisible(action)

		// Between every button we add the spacer.
		if position > 0 {
			buttonsRendering = append(buttonsRendering, buttonSpacing)
		}

		// Buttons are styled dynamically based on whether they are currently selected by the cursor or not.
		if buttonIndex == model.cursorPosition {
			buttonsRendering = append(
				buttonsRendering,
				model.style.buttonFocused.Render(model.buttons[buttonIndex].label),
			)
		} else {
			buttonsRendering = append(
				buttonsRendering,
				model.style.buttonBlurred.Render(model.buttons[buttonIndex].label),
			)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, buttonsRendering...)
}

// renderHint builds a hint line from the keymap bindings, separated by hintSeparator. There are only four actions the
// user can do in this UI state. The move left/right actions are joined together for styling purposes.
func (model *Model) renderHint() string {
	moveLeftKey := model.keyMap.MoveLeft.Help().Key
	moveRightKey := model.keyMap.MoveRight.Help().Key
	selectKey := model.keyMap.Select.Help().Key
	quitKey := model.keyMap.Quit.Help().Key

	parts := []string{
		moveLeftKey + "/" + moveRightKey + " to move",
		selectKey + " to select",
		quitKey + " to quit",
	}

	return strings.Join(parts, hintSeparator)
}
