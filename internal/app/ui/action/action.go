// Package action defines the action contract between UI components and the UIModel. A component returns an Action from
// HandleMessage, while UIModel type-switches on it to decide state transitions and side effects.
package action

import (
	tea "charm.land/bubbletea/v2"
)

// Action is any value a component returns from HandleMessage. UIModel type-switches on it.
type Action any

// Component is a UI unit the UIModel forwards messages to and receives actions from.
type Component interface {
	HandleMessage(msg tea.Msg) Action
}

// NoAction means the component consumed the message but requests nothing.
type NoAction struct{}

// QuitAction requests application termination.
type QuitAction struct{}

// ReloadConfigAction requests a config reload.
type ReloadConfigAction struct{}

// ProceedFromInitAction requests to proceed to the next state, from init state.
type ProceedFromInitAction struct{}

// ActionCommand carries a tea.Cmd for a component that needs a side effect but doesn't have a proper defined Action for it.
type ActionCommand struct {
	Command tea.Cmd
}
