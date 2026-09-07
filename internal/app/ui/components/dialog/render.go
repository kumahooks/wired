package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Render returns the dialog card, with the description wrapped on the card's width with the two buttons below it.
func (model *Model) Render() string {
	buttons := strings.Join(
		[]string{
			model.renderButton(model.confirmLabel, model.cursorPosition == int(confirmButton)),
			model.renderButton(model.cancelLabel, model.cursorPosition == int(cancelButton)),
		},
		strings.Repeat(" ", buttonGap),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		model.style.text.Width(cardWidth).Render(model.text),
		"",
		buttons,
	)

	return model.style.card.Render(content)
}

// renderButton renders a single button in either its focused or blurred state.
func (model *Model) renderButton(label string, focused bool) string {
	if focused {
		return model.style.buttonFocused.Render(label)
	}

	return model.style.buttonBlurred.Render(label)
}
