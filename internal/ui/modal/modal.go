// Package modal implements a text-input overlay for collecting user input
package modal

import (
	"slices"

	textinput "github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"

	config "wired/internal/config"
)

type Type int

const (
	MusicPath Type = iota
)

type SubmitMsg struct {
	Type  Type
	Value string
}

type CancelMsg struct {
	Type Type
}

type Style struct {
	BorderColor  lipgloss.Color
	CursorFg     lipgloss.Color
	InactiveText lipgloss.Color
}

func defaultStyle() Style {
	return Style{
		BorderColor:  lipgloss.Color("#6f3d49"),
		CursorFg:     lipgloss.Color("#965363"),
		InactiveText: lipgloss.Color("#44262d"),
	}
}

type Modal struct {
	input      textinput.Model
	title      string
	promptType Type
	visible    bool
	width      int
	height     int
	style      Style
}

func New() Modal {
	return Modal{
		input: textinput.New(),
		style: defaultStyle(),
	}
}

func (modal *Modal) Show(promptType Type, title string, placeholder string, charLimit int) bubbletea.Cmd {
	modal.promptType = promptType
	modal.title = title

	modal.input.Placeholder = placeholder
	modal.input.CharLimit = charLimit
	modal.input.SetValue("")
	modal.visible = true

	return modal.input.Focus()
}

func (modal *Modal) Hide() {
	modal.visible = false
	modal.input.Blur()
}

func (modal Modal) Visible() bool {
	return modal.visible
}

func (modal *Modal) SetSize(width int, height int) {
	modal.width = width
	modal.height = height

	modal.input.Width = min(width-10, 36)
}

func (modal *Modal) ApplyColors(cfg *config.Config) {
	modal.style = Style{
		BorderColor:  lipgloss.Color(cfg.Colors.Border),
		CursorFg:     lipgloss.Color(cfg.Colors.CursorForeground),
		InactiveText: lipgloss.Color(cfg.Colors.TextInactive),
	}

	modal.input.Cursor.Style = lipgloss.NewStyle().Foreground(modal.style.CursorFg)
	modal.input.PlaceholderStyle = lipgloss.NewStyle().Foreground(modal.style.InactiveText)
	modal.input.PromptStyle = lipgloss.NewStyle().Foreground(modal.style.InactiveText)
}

func (modal *Modal) Update(msg bubbletea.Msg, keybinds config.KeybindMapping) bubbletea.Cmd {
	if !modal.visible {
		return nil
	}

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		key := msg.String()
		promptType := modal.promptType

		// Confirm input
		if slices.Contains(keybinds.Select, key) {
			value := modal.input.Value()
			modal.Hide()

			return func() bubbletea.Msg { return SubmitMsg{Type: promptType, Value: value} }
		}

		// Leave input screen
		if slices.Contains(keybinds.Cancel, key) {
			modal.Hide()

			return func() bubbletea.Msg { return CancelMsg{Type: promptType} }
		}
	}

	var cmd bubbletea.Cmd
	modal.input, cmd = modal.input.Update(msg)

	return cmd
}

func (modal Modal) View() string {
	if !modal.visible {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(modal.style.BorderColor).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(modal.style.CursorFg)

	content := titleStyle.Render(modal.title) + "\n\n" + modal.input.View()
	box := boxStyle.Render(content)

	return lipgloss.Place(modal.width, modal.height, lipgloss.Center, lipgloss.Center, box)
}
