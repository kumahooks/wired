package components

import (
	"slices"

	textinput "github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"

	config "wired/internal/config"
)

type PromptType int

const (
	PromptMusicPath PromptType = iota
)

type PromptSubmitMsg struct {
	Type  PromptType
	Value string
}

type PromptCancelMsg struct {
	Type PromptType
}

type Prompt struct {
	input      textinput.Model
	title      string
	promptType PromptType
	visible    bool
	width      int
	height     int
}

func (prompt *Prompt) Show(promptType PromptType, title string, placeholder string, charLimit int) bubbletea.Cmd {
	prompt.promptType = promptType
	prompt.title = title

	prompt.input.Placeholder = placeholder
	prompt.input.CharLimit = charLimit
	prompt.input.SetValue("")
	prompt.visible = true

	return prompt.input.Focus()
}

func (prompt *Prompt) Hide() {
	prompt.visible = false
	prompt.input.Blur()
}

func (prompt Prompt) Visible() bool {
	return prompt.visible
}

func (prompt *Prompt) SetSize(width int, height int) {
	prompt.width = width
	prompt.height = height

	prompt.input.Width = min(width-10, 36)
}

func (prompt *Prompt) Update(msg bubbletea.Msg, keybinds config.KeybindMapping) bubbletea.Cmd {
	if !prompt.visible {
		return nil
	}

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		key := msg.String()
		promptType := prompt.promptType

		// Confirm input
		if slices.Contains(keybinds.Select, key) {
			value := prompt.input.Value()
			prompt.Hide()

			return func() bubbletea.Msg { return PromptSubmitMsg{Type: promptType, Value: value} }
		}

		// Leave input screen
		if slices.Contains(keybinds.Cancel, key) {
			prompt.Hide()

			return func() bubbletea.Msg { return PromptCancelMsg{Type: promptType} }
		}
	}

	var cmd bubbletea.Cmd
	prompt.input, cmd = prompt.input.Update(msg)

	return cmd
}

func (prompt Prompt) View(cfg *config.Config) string {
	if !prompt.visible {
		return ""
	}

	borderColor := lipgloss.Color(cfg.Colors.Border)
	cursorFg := lipgloss.Color(cfg.Colors.CursorForeground)
	inactiveText := lipgloss.Color(cfg.Colors.TextInactive)

	prompt.input.Cursor.Style = lipgloss.NewStyle().
		Foreground(cursorFg)

	prompt.input.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(inactiveText)

	prompt.input.PromptStyle = lipgloss.NewStyle().
		Foreground(inactiveText)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cursorFg)

	content := titleStyle.Render(prompt.title) + "\n\n" + prompt.input.View()
	box := boxStyle.Render(content)

	return lipgloss.Place(prompt.width, prompt.height, lipgloss.Center, lipgloss.Center, box)
}

func NewPrompt() Prompt {
	return Prompt{input: textinput.New()}
}
