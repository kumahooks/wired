// Package ui implements the main tea model of the application.
package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/app/ui/components/notification"
	"wired/internal/app/ui/components/whichkey"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

type Library struct {
	// scanGeneration tags the in-flight scan, incremented each time a new scan starts so stale results from a previous
	// scan or a cancelled scan are ignored.
	scanGeneration uint64
	// scanCancel aborts the current scan, if any. It is nil when no scan is running.
	scanCancel context.CancelFunc

	// audioFiles is a pointer to the orchestrator's owned library data structure. It holds every information of every
	// audio file the application has both loaded and saved in cache. This is seeded at load time, and later on command
	// through the ScanLibraryFullAction.
	audioFiles *[]audio.File
}

// UIModel holds the tea model state, primitives, and components.
type UIModel struct {
	// windowHeight and windowWidth are actual view space excluding borders.
	windowHeight int
	windowWidth  int

	// state decides what state the view is, essentially separating between initialization and idle.
	state uiState

	// keymaps are the shortcuts (default or user's) for the application's actions.
	keyMap keymap.KeyMap

	// theme is the resolved (default or user's) color palette shared with every component.
	theme theme.Theme

	// config is the shared config pointer.
	config *config.Config

	// orchestratorContext is the application's context, used to propagate cancel everywhere else.
	orchestratorContext context.Context

	// library is the user's loaded library data.
	library Library

	// expireNotificationCmds holds expire commands queued by PushNotification. Update drains it into the batch before returning.
	expireNotificationCmds []tea.Cmd

	// Components models.
	initializationModel *initializing.Model
	whichkeyModel       *whichkey.Model
	notificationModel   *notification.Model
}

// New initializes the UIModel, which is basically the UI orchestrator of the application. It starts with the state `uiInitializing`.
// To avoid locking the user out of actions, default configs, keymaps, and styles are loaded at first.
func New(
	orchestratorCtx context.Context,
	defaultKeyMap keymap.KeyMap,
	config *config.Config,
	audioFiles *[]audio.File,
) (*UIModel, error) {
	model := &UIModel{
		state:               uiInitializing,
		keyMap:              defaultKeyMap,
		theme:               theme.Default(),
		config:              config,
		orchestratorContext: orchestratorCtx,
		initializationModel: initializing.New(defaultKeyMap),
		whichkeyModel:       whichkey.New(),
		notificationModel:   notification.New(),
		library: Library{
			audioFiles: audioFiles,
		},
	}

	model.whichkeyModel.SetBindings(model.commandBindingsFor(model.state))

	return model, nil
}

// Init is the first function that will be called. It returns an optional initial command. (bubbletea's own words)
func (model *UIModel) Init() tea.Cmd {
	// Init sends a tea.Cmd message to load the user's config. It's the very first thing we run, getting the app ready to use.
	return initializationLoadConfigCommand()
}

// PushNotification queues a notification for display and schedules its expiry command.
func (model *UIModel) PushNotification(message string) {
	model.notificationModel.PushNotification(message)
	model.expireNotificationCmds = append(model.expireNotificationCmds, notificationExpireCommand())
}

// drainNotificationCmds hands over the expire commands queued and resets the queue.
func (model *UIModel) drainNotificationCmds() []tea.Cmd {
	commands := model.expireNotificationCmds
	model.expireNotificationCmds = nil

	return commands
}

func (model *UIModel) setState(state uiState) {
	model.state = state
	model.whichkeyModel.SetBindings(model.commandBindingsFor(state))
}

// commandBindingsFor compiles the bindings active for the given UI state.
func (model *UIModel) commandBindingsFor(state uiState) []action.Binding {
	// TODO: there will obviously be more actions and states... is this way really good enough?
	var bindings []action.Binding

	switch state {
	case uiPlaylist:
		bindings = append(
			bindings,
			action.Binding{Keys: model.keyMap.Actions.LibraryStats, Action: action.OpenLibraryStatsAction{}},
		)
	case uiLibraryStats:
		bindings = append(
			bindings,
			action.Binding{Keys: model.keyMap.Actions.Playlist, Action: action.OpenPlaylistAction{}},
		)
	}

	return bindings
}
