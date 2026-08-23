// Package initializing implements the initialization view, responsible for sending feedbacks to the user regarding the
// initialization pipeline.
package initializing

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// New returns a Model seeded with the buttons actions, the default keymap, and the default theme.
func New(defaultKeyMap keymap.KeyMap) *Model {
	return &Model{
		buttons: []button{
			{label: "reload config", action: actionReload},
			{label: "proceed anyway", action: actionProceed},
		},
		keyMap: defaultKeyMap,
		style:  newStyle(theme.Default()),
	}
}

// ApplyTheme rebuilds the component style from a resolved `theme.Theme`.
func (model *Model) ApplyTheme(resolvedTheme theme.Theme) {
	model.style = newStyle(resolvedTheme)
}

// ApplyKeyMap stores the keymap used for button navigation and the hint line.
func (model *Model) ApplyKeyMap(resolvedKeyMap keymap.KeyMap) {
	model.keyMap = resolvedKeyMap
}

// AppendLog adds a line to the log buffer shown in the view.
func (model *Model) AppendLog(line string, logType LogType) {
	model.logCount++
	model.logLines = append(model.logLines, logLine{
		text:    fmt.Sprintf("[%d] %s", model.logCount, line),
		logType: logType,
	})
}

// HandleMsg handles keyboard navigation for the button row.
func (model *Model) HandleMessage(message tea.Msg) action.Action {
	keyPress, ok := message.(tea.KeyPressMsg)
	if !ok {
		return action.NoAction{}
	}

	switch {
	case key.Matches(keyPress, model.keyMap.MoveLeft):
		model.moveCursor(-1)
	case key.Matches(keyPress, model.keyMap.MoveRight):
		model.moveCursor(1)
	case key.Matches(keyPress, model.keyMap.Select):
		if len(model.buttons) == 0 {
			return action.NoAction{}
		}

		switch model.buttons[model.cursorPosition].action {
		case actionReload:
			return action.ReloadConfigAction{}
		case actionProceed:
			return action.ProceedFromInitAction{}
		}
	}

	return action.NoAction{}
}
