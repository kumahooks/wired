// Package footer implements an ever-present status bar at the bottom of the screen
package footer

import (
	"strings"

	spinner "github.com/charmbracelet/bubbles/spinner"
	bubbletea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"

	config "wired/internal/config"
)

type State int

const (
	Starting State = iota
	ConfigLoading
	WaitingUserInput
	Idle
	Error
)

type StartCompleteMsg struct{}

type Style struct {
	BarFg   lipgloss.Color
	HintFg  lipgloss.Color
	LabelBg lipgloss.Color
	LabelFg lipgloss.Color
	ErrorBg lipgloss.Color
	ErrorFg lipgloss.Color
}

func defaultStyle() Style {
	return Style{
		BarFg:   lipgloss.Color("#965363"),
		HintFg:  lipgloss.Color("#44262d"),
		LabelBg: lipgloss.Color("#6f3d49"),
		LabelFg: lipgloss.Color("#1a0f12"),
		ErrorBg: lipgloss.Color("#a52a2a"),
		ErrorFg: lipgloss.Color("#1a0f12"),
	}
}

type Content struct {
	Title     string
	Message   string
	Hint      string
	Separator string
}

type Footer struct {
	state   State
	spinner spinner.Model
	content Content
	width   int
	style   Style
}

func New() Footer {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return Footer{
		state:   Starting,
		spinner: s,
		content: Content{Title: "Starting...", Separator: " · "},
		style:   defaultStyle(),
	}
}

func (footer *Footer) SetState(state State) bubbletea.Cmd {
	wasSpinning := footer.isSpinning()

	footer.state = state
	footer.content = footer.getContent()

	if footer.isSpinning() && !wasSpinning {
		return footer.spinner.Tick
	}

	return nil
}

func (footer *Footer) SetContent(content Content) {
	footer.content = content
}

func (footer *Footer) SetWidth(width int) {
	footer.width = width
}

func (footer *Footer) ApplyConfig(cfg *config.Config) {
	footer.style = Style{
		BarFg:   lipgloss.Color(cfg.Colors.FooterBarFg),
		HintFg:  lipgloss.Color(cfg.Colors.FooterHintFg),
		LabelBg: lipgloss.Color(cfg.Colors.FooterLabelBg),
		LabelFg: lipgloss.Color(cfg.Colors.FooterLabelFg),
		ErrorBg: lipgloss.Color(cfg.Colors.FooterErrorBg),
		ErrorFg: lipgloss.Color(cfg.Colors.FooterErrorFg),
	}
}

func (footer Footer) State() State {
	return footer.state
}

func (footer *Footer) Update(msg bubbletea.Msg) bubbletea.Cmd {
	if !footer.isSpinning() {
		return nil
	}

	var cmd bubbletea.Cmd
	footer.spinner, cmd = footer.spinner.Update(msg)

	return cmd
}

func (footer Footer) View() string {
	labelBg := footer.style.LabelBg
	labelFg := footer.style.LabelFg

	if footer.state == Error {
		labelBg = footer.style.ErrorBg
		labelFg = footer.style.ErrorFg
	}

	barStyle := lipgloss.NewStyle().Foreground(footer.style.BarFg)
	hintStyle := lipgloss.NewStyle().Foreground(footer.style.HintFg)
	labelStyle := lipgloss.NewStyle().Background(labelBg).Foreground(labelFg).Bold(true)

	sep := footer.content.Separator
	if sep == "" {
		sep = " "
	}

	var parts []string
	var contentWidth int

	// Title
	if footer.isSpinning() {
		titlePart := " " + footer.spinner.View() + " " + footer.content.Title
		titleRendered := barStyle.Render(titlePart)

		parts = append(parts, titleRendered)

		contentWidth = lipgloss.Width(titleRendered)
	} else {
		labelText := " " + footer.content.Title + " "
		label := labelStyle.Render(labelText)

		parts = append(parts, " ")
		parts = append(parts, label)

		contentWidth = 1 + lipgloss.Width(label)
	}

	// Message
	if footer.content.Message != "" {
		msgRendered := barStyle.Render(footer.content.Message)

		parts = append(parts, sep)
		parts = append(parts, msgRendered)

		contentWidth += lipgloss.Width(sep) + lipgloss.Width(msgRendered)
	}

	// Hint
	if footer.content.Hint != "" {
		hintRendered := hintStyle.Render(footer.content.Hint)

		hintTotalWidth := lipgloss.Width(sep) + lipgloss.Width(hintRendered)

		if contentWidth+hintTotalWidth <= footer.width {
			parts = append(parts, sep)
			parts = append(parts, hintRendered)
			contentWidth += hintTotalWidth
		}
	}

	// Padding
	if contentWidth < footer.width {
		pad := barStyle.Render(strings.Repeat(" ", footer.width-contentWidth))
		parts = append(parts, pad)
	}

	return strings.Join(parts, "")
}

func (footer Footer) getContent() Content {
	switch footer.state {
	case Starting:
		return Content{Title: "Starting...", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	case ConfigLoading:
		return Content{Title: "Loading config...", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	case WaitingUserInput:
		return Content{Title: "Waiting for user input...", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	case Error:
		return Content{Title: "ERROR", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	case Idle:
		return Content{Title: "NORMAL", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	default:
		return Content{Title: "", Message: "", Hint: "ctrl+c to quit", Separator: " · "}
	}
}

func (footer Footer) isSpinning() bool {
	return footer.state == Starting || footer.state == ConfigLoading || footer.state == WaitingUserInput
}
