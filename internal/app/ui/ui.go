// Package ui implements the main tea model of the application.
package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/dialog"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/app/ui/components/librarystats"
	"wired/internal/app/ui/components/notification"
	"wired/internal/app/ui/components/whichkey"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

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

	// library is a pointer to the orchestrator's owned library data structure.
	library *audio.Library
	// libraryDiscoveryGeneration tags the in-flight library discovery phase, incremented each time a phase starts.
	libraryDiscoveryGeneration uint64
	// libraryDiscoveryCancel aborts the current discovery, if any.
	libraryDiscoveryCancel context.CancelFunc

	// isConfigLoading guards against double-triggering a config load while one is already in flight.
	isConfigLoading bool

	// expireNotificationCmds holds expire commands queued by PushNotification. Update drains it into the batch before returning.
	expireNotificationCmds []tea.Cmd

	// Components models.
	initializationModel *initializing.Model
	libraryStatsModel   *librarystats.Model
	whichkeyModel       *whichkey.Model
	notificationModel   *notification.Model
	dialogModel         *dialog.Model
}

// New initializes the UIModel, which is basically the UI orchestrator of the application. It starts with the state
// `uiBootstrapping`, rendering nothing while the pipeline decides between the initialization screen (on error) and
// the idle UI. To avoid locking the user out of actions, default configs, keymaps, and styles are loaded at first.
func New(
	orchestratorCtx context.Context,
	defaultKeyMap keymap.KeyMap,
	config *config.Config,
	audioLibrary *audio.Library,
) (*UIModel, error) {
	model := &UIModel{
		state:               uiBootstrapping,
		keyMap:              defaultKeyMap,
		theme:               theme.Default(),
		config:              config,
		orchestratorContext: orchestratorCtx,
		initializationModel: initializing.New(defaultKeyMap),
		libraryStatsModel:   librarystats.New(defaultKeyMap, audioLibrary),
		whichkeyModel:       whichkey.New(),
		notificationModel:   notification.New(),
		dialogModel:         dialog.New(),
		library:             audioLibrary,
	}

	model.whichkeyModel.SetBindings(model.commandBindingsFor(model.state))
	model.dialogModel.ApplyKeyMap(defaultKeyMap)

	return model, nil
}

// Init is the first function that will be called. It returns an optional initial command. (bubbletea's own words)
func (model *UIModel) Init() tea.Cmd {
	// Init sends a tea.Cmd message to load the user's config. It's the very first thing ran, getting the app ready to use.
	model.isConfigLoading = true
	return configLoadCommand(configLoadOriginInit)
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
	case uiLibrary:
		bindings = append(
			bindings,
			action.Binding{Keys: model.keyMap.Actions.Playlist, Action: action.OpenPlaylistAction{}},
			action.Binding{Keys: model.keyMap.Actions.LibraryStats, Action: action.OpenLibraryStatsAction{}},
			action.Binding{Keys: model.keyMap.Actions.ReloadConfig, Action: action.ReloadConfigAction{}},
		)
	case uiPlaylist:
		bindings = append(
			bindings,
			action.Binding{Keys: model.keyMap.Actions.Library, Action: action.OpenLibraryAction{}},
			action.Binding{Keys: model.keyMap.Actions.LibraryStats, Action: action.OpenLibraryStatsAction{}},
			action.Binding{Keys: model.keyMap.Actions.ReloadConfig, Action: action.ReloadConfigAction{}},
		)
	case uiLibraryStats:
		bindings = append(
			bindings,
			action.Binding{Keys: model.keyMap.Actions.Library, Action: action.OpenLibraryAction{}},
			action.Binding{Keys: model.keyMap.Actions.Playlist, Action: action.OpenPlaylistAction{}},
			action.Binding{Keys: model.keyMap.Actions.ReloadConfig, Action: action.ReloadConfigAction{}},
		)
	}

	return bindings
}
