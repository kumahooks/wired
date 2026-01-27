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

type SpinnerTickMsg struct {
	inner spinner.TickMsg
}

type Content struct {
	Title     string
	Message   string
	Hint      string
	Separator string
}

var (
	fallbackBarFg   = lipgloss.Color("#965363")
	fallbackLabelBg = lipgloss.Color("#6f3d49")
	fallbackLabelFg = lipgloss.Color("#1a0f12")
	fallbackErrorBg = lipgloss.Color("#4b2f55")
	fallbackErrorFg = lipgloss.Color("#1a0f12")
	fallbackHintFg  = lipgloss.Color("#44262d")
)

type Footer struct {
	state   State
	spinner spinner.Model
	content Content
	width   int
}

func (footer *Footer) Init() bubbletea.Cmd {
	return wrapTickCmd(footer.spinner.Tick)
}

func (footer *Footer) SetState(state State) bubbletea.Cmd {
	wasSpinning := footer.isSpinning()

	footer.state = state
	footer.content = footer.getContent()

	if footer.isSpinning() && !wasSpinning {
		return wrapTickCmd(footer.spinner.Tick)
	}

	return nil
}

func (footer *Footer) SetContent(content Content) {
	footer.content = content
}

func (footer *Footer) SetWidth(width int) {
	footer.width = width
}

func (footer Footer) State() State {
	return footer.state
}

func (footer *Footer) Update(msg bubbletea.Msg) bubbletea.Cmd {
	if !footer.isSpinning() {
		return nil
	}

	if stm, ok := msg.(SpinnerTickMsg); ok {
		var cmd bubbletea.Cmd
		footer.spinner, cmd = footer.spinner.Update(stm.inner)

		return wrapTickCmd(cmd)
	}

	return nil
}

func (footer Footer) View(cfg *config.Config) string {
	barFg := fallbackBarFg
	hintFg := fallbackHintFg
	labelBg := fallbackLabelBg
	labelFg := fallbackLabelFg

	if footer.state == Error {
		labelBg = fallbackErrorBg
		labelFg = fallbackErrorFg
	}

	if cfg != nil {
		barFg = lipgloss.Color(cfg.Colors.FooterBarFg)
		hintFg = lipgloss.Color(cfg.Colors.FooterHintFg)

		if footer.state == Error {
			labelBg = lipgloss.Color(cfg.Colors.FooterErrorBg)
			labelFg = lipgloss.Color(cfg.Colors.FooterErrorFg)
		} else {
			labelBg = lipgloss.Color(cfg.Colors.FooterLabelBg)
			labelFg = lipgloss.Color(cfg.Colors.FooterLabelFg)
		}
	}

	barStyle := lipgloss.NewStyle().Foreground(barFg)
	hintStyle := lipgloss.NewStyle().Foreground(hintFg)
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

func New() Footer {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return Footer{
		state:   Starting,
		spinner: s,
		content: Content{Title: "Starting...", Separator: " · "},
	}
}

func wrapTickCmd(cmd bubbletea.Cmd) bubbletea.Cmd {
	if cmd == nil {
		return nil
	}

	return func() bubbletea.Msg {
		msg := cmd()

		if tm, ok := msg.(spinner.TickMsg); ok {
			return SpinnerTickMsg{inner: tm}
		}

		return msg
	}
}
