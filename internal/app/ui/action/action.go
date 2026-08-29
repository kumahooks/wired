// Package action defines the action contract between UI components and the UIModel. A component returns an Action from
// HandleMessage, while UIModel type-switches on it to decide state transitions and side effects.
package action

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Action is any value a component returns from HandleMessage. UIModel type-switches on it.
type Action any

// Binding associates a key binding with its action. UIModel exposes a commandBindingsFor primitive where it returns
// all of the bindings given the current UI state.
type Binding struct {
	Keys   key.Binding
	Action Action
}

// Component is a UI unit the UIModel forwards messages to and receives actions from.
type Component interface {
	HandleMessage(msg tea.Msg) Action
}

// NoAction means the component consumed the message but requests nothing.
type NoAction struct{}

// QuitAction requests application termination.
type QuitAction struct{}

// OpenWhichKeyAction opens the card with all the relevant application keybindings.
type OpenWhichKeyAction struct{}

// ReloadConfigAction requests a config reload.
type ReloadConfigAction struct{}

// OpenPlaylistAction goes to the playlist screen.
type OpenPlaylistAction struct{}

// OpenLibraryStatsAction goes to the library stats screen.
type OpenLibraryStatsAction struct{}

// ScanLibraryFullAction requests a full re-scan of the known library paths.
type ScanLibraryFullAction struct{}

// ProceedFromInitAction requests to proceed to the next state, from init state.
type ProceedFromInitAction struct{}

// ActionCommand carries a tea.Cmd for a component that needs a side effect but doesn't have a proper defined Action for it.
type ActionCommand struct {
	Command tea.Cmd
}
