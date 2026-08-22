// Package main is the entry point of our application.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"wired/internal/ui"
	"wired/internal/wired"
)

func main() {
	wired, err := wired.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}

	uiModel, err := ui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}

	wired.Foo()
	wiredProgram := tea.NewProgram(uiModel)
	if _, err := wiredProgram.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
