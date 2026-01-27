// Package ui handles the bubbletea lifecycle and components
package ui

import (
	bubbletea "github.com/charmbracelet/bubbletea"

	cli "wired/internal/cli"
)

func Start() error {
	cli.ClearScreen()

	p := bubbletea.NewProgram(NewModel(), bubbletea.WithAltScreen())
	_, err := p.Run()

	return err
}
