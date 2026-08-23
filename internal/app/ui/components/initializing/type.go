package initializing

import (
	"wired/internal/core/keymap"
)

// buttonAction maps a button to the action it produces on select.
type buttonAction int

const (
	actionReload buttonAction = iota
	actionProceed
)

// button is a selectable action rendered in the button row.
type button struct {
	label  string
	action buttonAction
}

// LogType marks a log line as normal or an error so it can be colored differently in the log area.
type LogType int

const (
	LogNormal LogType = iota
	LogError
)

// logLine is a single entry in the startup log buffer.
type logLine struct {
	text    string
	logType LogType
}

// Model holds the init view state and data.
type Model struct {
	// log lines is a textarea showing a log of all the actions so far, including errors.
	// TODO: In the future I plan to have a log service in order to save the app's logs to a file. I like the idea of this
	// pulling the logs from this service then.
	logLines []logLine
	logCount int

	// countFilesProgress is the live counter shown between the log area and the hint while the files are being counted.
	// A negative value means no count is in progress and the line is hidden.
	countFilesProgress int

	// There are two types of actions in the initializitation screen, depending on the config/libraries state. Either the
	// user reloads the config while trying to fix a problem, or chooses to proceed which in this case will use the defaults.
	buttons        []button
	cursorPosition int

	// We use this keymap to properly map actions, and render the hint of the supported actions.
	keyMap keymap.KeyMap

	// This is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}
