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
	model := &Model{
		mode: modeLoading,
		buttons: []button{
			{label: "scan files", action: actionScan},
			{label: "reload config", action: actionReload},
			{label: "proceed anyway", action: actionProceed},
		},
		keyMap:             defaultKeyMap,
		style:              newStyle(theme.Default()),
		fetchFilesProgress: -1,
	}

	model.cursorPosition = model.canonicalIndexForVisible(model.firstVisibleAction())

	return model
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

// HandleMessage handles keyboard navigation for the button row.
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
		case actionScan:
			return action.ScanLibraryFullAction{}
		case actionProceed:
			return action.ProceedFromInitAction{}
		}
	}

	return action.NoAction{}
}

// SetFetchFilesProgress stores the latest fetched-files total for rendering.
func (model *Model) SetFetchFilesProgress(count int) {
	model.fetchFilesProgress = count
}

// SetConfigError transitions the screen to modeConfigError. The user can either reload the config, or just proceed.
func (model *Model) SetConfigError() {
	model.setMode(modeConfigError)
}

// SetEmptyLibrary transitions the screen to modeEmptyLibrary. The user can either trigger a scan, or just proceed.
func (model *Model) SetEmptyLibrary() {
	model.setMode(modeEmptyLibrary)
}

// setMode stores the given mode and moves the cursor onto the first visible button.
func (model *Model) setMode(mode initMode) {
	model.mode = mode
	model.cursorPosition = model.canonicalIndexForVisible(model.firstVisibleAction())
}

// buttonsForMode returns the subset of buttonActions visible in the given mode.
func buttonsForMode(mode initMode) []buttonAction {
	switch mode {
	case modeConfigError:
		return []buttonAction{actionReload, actionProceed}
	case modeEmptyLibrary:
		return []buttonAction{actionScan, actionProceed}
	default:
		return []buttonAction{actionProceed}
	}
}

// firstVisibleAction returns the first visible buttonAction in the current mode.
func (model *Model) firstVisibleAction() buttonAction {
	return buttonsForMode(model.mode)[0]
}

// visibleActions returns the list of buttonActions visible in the current mode.
func (model *Model) visibleActions() []buttonAction {
	return buttonsForMode(model.mode)
}

// canonicalIndexForVisible maps a buttonAction to its index within the model's buttons slice.
func (model *Model) canonicalIndexForVisible(action buttonAction) int {
	for index, candidate := range model.buttons {
		if candidate.action == action {
			return index
		}
	}

	return 0
}

// positionOfVisible returns the position of the given action within the current mode's visible actions.
func (model *Model) positionOfVisible(action buttonAction) int {
	for index, visible := range model.visibleActions() {
		if visible == action {
			return index
		}
	}

	return 0
}

// Mode returns the current initialization mode.
func (model *Model) Mode() initMode {
	return model.mode
}

// IsConfigError reports whether the screen is in modeConfigError.
func (model *Model) IsConfigError() bool {
	return model.mode == modeConfigError
}

// LogLines returns the texts of all log lines currently stored in the model.
func (model *Model) LogLines() []string {
	texts := make([]string, len(model.logLines))
	for index, entry := range model.logLines {
		texts[index] = entry.text
	}

	return texts
}

// LastLogType returns the LogType of the most recently appended log line.
func (model *Model) LastLogType() LogType {
	if len(model.logLines) == 0 {
		return LogNormal
	}

	return model.logLines[len(model.logLines)-1].logType
}

// FetchFilesProgress returns the current live counter value. A negative value means no fetch is in progress.
func (model *Model) FetchFilesProgress() int {
	return model.fetchFilesProgress
}
