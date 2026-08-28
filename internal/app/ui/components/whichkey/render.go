package whichkey

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// Render returns the whichkey card.
//
// the mapped actions are sorted by rendered width in ascending order, packe into rows, and the card grows with the number
// of rows up to maxHeight (half the window).
func (model *Model) Render(windowWidth int, windowHeight int) string {
	actions := model.mappedActions()
	slices.SortFunc(actions, func(first actionEntry, second actionEntry) int {
		return first.width - second.width
	})

	rows := model.layoutRows(actions, windowWidth, windowHeight)

	lines := slices.Clone(rows)
	for gap := 0; gap < hintGapLines; gap++ {
		lines = append(lines, "")
	}
	lines = append(
		lines,
		model.style.hint.Width(windowWidth).AlignHorizontal(lipgloss.Center).Render(model.renderHint()),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return model.style.card.
		Width(windowWidth).
		MaxWidth(windowWidth).
		Height(model.cardHeight(len(rows))).
		Render(content)
}

// cardHeight is the card's rendered height for a given number of action rows.
func (model *Model) cardHeight(rows int) int {
	return cardTopPadding + rows + hintGapLines + 1
}

// actionEntry is a single action mapped from the keymap and its rendered text width.
type actionEntry struct {
	text  string
	width int
}

// mappedActions compiles every ActionsKeyMap key into an actionEntry.
func (model *Model) mappedActions() []actionEntry {
	playlist := model.keyMap.Actions.Playlist.Help()
	libraryStats := model.keyMap.Actions.LibraryStats.Help()

	return []actionEntry{
		model.actionEntry(playlist.Key, playlist.Desc),
		model.actionEntry(libraryStats.Key, libraryStats.Desc),
	}
}

// actionEntry renders one "{key} -> {description}" line together with its width.
func (model *Model) actionEntry(key string, description string) actionEntry {
	text := model.renderEntry(key, description)
	return actionEntry{text: text, width: lipgloss.Width(text)}
}

// layoutRows renders the mapped action entries as rows.
func (model *Model) layoutRows(actions []actionEntry, windowWidth int, windowHeight int) []string {
	maxHeight := windowHeight / 2
	usableRows := maxHeight - cardTopPadding - hintGapLines - 1 // this 1 is the hint line

	var rows []string
	var currentRow []string
	currentWidth := 0

	flushRow := func() {
		rows = append(rows, strings.Join(currentRow, strings.Repeat(" ", columnGap)))
		currentRow = nil
		currentWidth = 0
	}

	for _, act := range actions {
		rowWidth := act.width
		if currentWidth > 0 {
			rowWidth += columnGap
		}

		if currentWidth+rowWidth > windowWidth && len(currentRow) > 0 {
			flushRow()
			rowWidth = act.width
		}

		currentRow = append(currentRow, act.text)
		currentWidth += rowWidth
	}

	if len(currentRow) > 0 {
		flushRow()
	}

	if len(rows) > usableRows {
		rows = rows[:max(usableRows, 0)]
		panic("TODO: whichkey action rendering pagination Q_Q")
	}

	return rows
}

// renderEntry formats a single command line in the "{key} -> {description}" form.
func (model *Model) renderEntry(key string, description string) string {
	parts := []string{
		model.style.key.Render(key),
		model.style.separator.Render("->"),
		model.style.description.Render(description),
	}

	return strings.Join(parts, " ")
}

// renderHint builds the close hint shown on the card's last line.
func (model *Model) renderHint() string {
	goBackKey := model.keyMap.GoBack.Help().Key

	parts := []string{
		model.style.key.Render(goBackKey),
		model.style.separator.Render("·"),
		model.style.description.Render("close"),
	}

	return strings.Join(parts, " ")
}
