package initializing

import (
	"wired/internal/core/keymap"
)

// actionButton maps a button to the action it produces on select.
type actionButton int

const (
	reloadConfigAction actionButton = iota
	proceedAction
)

// button is a selectable action rendered in the button row.
type button struct {
	label  string
	action actionButton
}

// initMode represents the state of the initialization screen, which decides which buttons are shown to the user.
type initMode uint8

const (
	// modeLoading is the default state while config and cache are still being checked.
	modeLoading initMode = iota

	// modeConfigError means the config failed to load or has an unparseable keymap.
	modeConfigError
)

// LogType marks a log line as normal or an error so it can be colored differently in the log area.
type LogType int

const (
	LogNormal LogType = iota
	LogWarning
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

	// mode decides which buttons are shown to the user depending on the config and cache state.
	mode initMode

	// buttons is the full set of selectable button actions, and which ones are visible is derived directly from mode.
	buttons        []button
	cursorPosition int

	// keymap is used to properly map actions, and render the hint of the supported init actions.
	keyMap keymap.KeyMap

	// style is the styles (such as lipgloss colors) used in the view rendering.
	style Style
}
