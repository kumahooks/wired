// Package dialog implements a centered text bubble for displaying information, config agnostic
package dialog

import (
	"strings"

	viewport "github.com/charmbracelet/bubbles/viewport"
	bubbletea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

const (
	maxContentWidth  = 60
	maxContentHeight = 20
)

var (
	borderColor = lipgloss.Color("#6f3d49")
	dimColor    = lipgloss.Color("#44262d")
	accentColor = lipgloss.Color("#965363")
)

type Options struct {
	Header string
	Body   string
	Footer string
}

type Dialog struct {
	viewport viewport.Model
	body     string
	header   string
	footer   string
	visible  bool
	width    int
	height   int
}

func New() Dialog {
	return Dialog{
		viewport: viewport.New(0, 0),
	}
}

func (dialog *Dialog) Show(opts Options) {
	dialog.header = opts.Header
	dialog.body = opts.Body
	dialog.footer = opts.Footer
	dialog.visible = true

	dialog.recalcViewport()
}

func (dialog *Dialog) Hide() {
	dialog.visible = false
}

func (dialog Dialog) Visible() bool {
	return dialog.visible
}

func (dialog *Dialog) SetSize(width int, height int) {
	dialog.width = width
	dialog.height = height

	dialog.recalcViewport()
}

func (dialog *Dialog) Update(msg bubbletea.Msg) bubbletea.Cmd {
	var cmd bubbletea.Cmd
	dialog.viewport, cmd = dialog.viewport.Update(msg)

	return cmd
}

func (dialog Dialog) View() string {
	if !dialog.visible {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	dimStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor)

	var sections []string

	if dialog.header != "" {
		sections = append(sections, headerStyle.Render(dialog.header))
		sections = append(sections, "")
	}

	sections = append(sections, dialog.viewport.View())

	if dialog.footer != "" {
		sections = append(sections, "")
		sections = append(sections, dimStyle.Render(dialog.footer))
	}

	content := strings.Join(sections, "\n")
	box := boxStyle.Render(content)

	return lipgloss.Place(dialog.width, dialog.height, lipgloss.Center, lipgloss.Center, box)
}

func (dialog *Dialog) recalcViewport() {
	contentWidth := min(maxContentWidth, max(dialog.width-30, 20))

	// Wrap body text to fit the viewport width
	wrapped := lipgloss.NewStyle().Width(contentWidth).Render(dialog.body)
	wrappedLines := strings.Count(wrapped, "\n") + 1

	// Overhead: border (2) + vertical padding (2) + header (2) + footer (2)
	overhead := 2 + 2
	if dialog.header != "" {
		overhead += 2
	}
	if dialog.footer != "" {
		overhead += 2
	}

	termMax := max(dialog.height-overhead, 3)

	viewportHeight := min(wrappedLines, maxContentHeight, termMax)
	viewportHeight = max(viewportHeight, 1)

	dialog.viewport.Width = contentWidth
	dialog.viewport.Height = viewportHeight
	dialog.viewport.SetContent(wrapped)
	dialog.viewport.GotoTop()
}
