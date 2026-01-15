// Package wired initializes the program with Run()
package wired

import (
	fmt "fmt"
	"os"

	cli "wired/src/cli"
	core "wired/src/core"
	tea "wired/src/core/tea"

	bubbletea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	cli.ClearScreen()

	coreModel := core.NewCoreModel()
	teaModel := &tea.TeaModel{CoreModel: coreModel}

	p := bubbletea.NewProgram(teaModel, bubbletea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return nil
}
