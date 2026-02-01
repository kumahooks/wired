// Package footer implements an ever-present status bar at the bottom of the screen
package footer

import (
	"fmt"
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
	LibraryLoading
	LibraryScanning
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
	state    State
	spinner  spinner.Model
	content  Content
	width    int
	style    Style
	keybinds config.KeybindMapping
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

func (footer Footer) Init() bubbletea.Cmd {
	if footer.isSpinning() {
		return footer.spinner.Tick
	}

	return nil
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

func (footer *Footer) SetScanState(current int, total int) {
	footer.content.Message = fmt.Sprintf("%d/%d", current, total)
}

func (footer *Footer) SetWidth(width int) {
	footer.width = width
}

func (footer *Footer) ApplyConfig(cfg *config.Config) {
	footer.keybinds = cfg.Keybinds
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
	padding := 1
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

	innerWidth := max(footer.width-padding*2, 0)

	var parts []string
	var contentWidth int

	// Title
	if footer.isSpinning() {
		titlePart := footer.spinner.View() + " " + footer.content.Title
		titleRendered := barStyle.Render(titlePart)

		parts = append(parts, titleRendered)

		contentWidth = lipgloss.Width(titleRendered)
	} else {
		labelText := " " + footer.content.Title + " "
		label := labelStyle.Render(labelText)

		parts = append(parts, label)

		contentWidth = lipgloss.Width(label)
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

		if contentWidth+hintTotalWidth <= innerWidth {
			parts = append(parts, sep)
			parts = append(parts, hintRendered)
			contentWidth += hintTotalWidth
		}
	}

	inner := strings.Join(parts, "")

	return lipgloss.NewStyle().
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(inner)
}

func (footer Footer) getContent() Content {
	quitHint := footer.keybindHint(footer.keybinds.Quit, "to quit", "ctrl+c to quit")
	scanHint := footer.keybindHint(footer.keybinds.ScanFiles, "to stop", "ctrl+s to stop")

	switch footer.state {
	case Starting:
		return Content{Title: "Starting...", Hint: quitHint, Separator: " · "}
	case ConfigLoading:
		return Content{Title: "Loading config...", Hint: quitHint, Separator: " · "}
	case LibraryLoading:
		return Content{Title: "Loading library...", Hint: quitHint, Separator: " · "}
	case LibraryScanning:
		return Content{
			Title:     "Scanning library...",
			Hint:      scanHint,
			Separator: " · ",
		}
	case WaitingUserInput:
		return Content{Title: "Waiting for user input...", Hint: quitHint, Separator: " · "}
	case Error:
		return Content{Title: "ERROR", Hint: quitHint, Separator: " · "}
	case Idle:
		return Content{Title: "NORMAL", Hint: quitHint, Separator: " · "}
	default:
		return Content{Title: "", Hint: quitHint, Separator: " · "}
	}
}

func (footer Footer) keybindHint(keys []string, action string, fallback string) string {
	if len(keys) == 0 {
		return fallback
	}

	return keys[0] + " " + action
}

func (footer Footer) isSpinning() bool {
	return footer.state == Starting ||
		footer.state == ConfigLoading ||
		footer.state == WaitingUserInput ||
		footer.state == LibraryLoading ||
		footer.state == LibraryScanning
}
