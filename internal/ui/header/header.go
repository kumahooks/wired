// Package header renders a horizontal navigation bar for switching between views
package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"wired/internal/config"
)

type contentView int

const (
	Library contentView = iota
	Playlist
	Statistics
)

type menuEntry struct {
	label       string
	prefix      string
	contentView contentView
}

var entries = []menuEntry{
	{label: "Library", prefix: "(L) ", contentView: Library},
	{label: "Playlist", prefix: "(P) ", contentView: Playlist},
	{label: "Statistics", prefix: "(S) ", contentView: Statistics},
}

type Style struct {
	ActiveBg   lipgloss.Color
	ActiveFg   lipgloss.Color
	InactiveFg lipgloss.Color
}

func defaultStyle() Style {
	return Style{
		ActiveBg:   lipgloss.Color("#6f3d49"),
		ActiveFg:   lipgloss.Color("#1a0f12"),
		InactiveFg: lipgloss.Color("#44262d"),
	}
}

type Header struct {
	active contentView
	width  int
	style  Style
}

func New() Header {
	return Header{
		active: Library,
		style:  defaultStyle(),
	}
}

func (header *Header) SetActive(view contentView) {
	header.active = view
}

func (header Header) Active() contentView {
	return header.active
}

func (header *Header) SetWidth(width int) {
	header.width = width
}

func (header *Header) ApplyConfig(cfg *config.Config) {
	header.style = Style{
		ActiveBg:   lipgloss.Color(cfg.Colors.HeaderActiveBg),
		ActiveFg:   lipgloss.Color(cfg.Colors.HeaderActiveFg),
		InactiveFg: lipgloss.Color(cfg.Colors.HeaderInactiveFg),
	}
}

func (header Header) View() string {
	// TODO: this should be more beautiful... how?
	padding := 1

	activeStyle := lipgloss.NewStyle().
		Background(header.style.ActiveBg).
		Foreground(header.style.ActiveFg).
		Bold(true)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(header.style.InactiveFg)

	var parts []string

	for _, e := range entries {
		label := e.prefix + e.label + " "

		var rendered string
		if e.contentView == header.active {
			rendered = activeStyle.Render(label)
		} else {
			rendered = inactiveStyle.Render(label)
		}

		parts = append(parts, rendered)
	}

	inner := strings.Join(parts, "")

	return lipgloss.NewStyle().
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(inner)
}
