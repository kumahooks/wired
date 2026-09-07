package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/dialog"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/testutil"
	"wired/internal/core/theme"
)

func isTeaQuit(command tea.Cmd) bool {
	if command == nil {
		return false
	}

	_, ok := command().(tea.QuitMsg)

	return ok
}

// dialogKeyGraceTestMargin is the extra delay tests add past the dialog's key grace quiet period so keys are accepted.
const dialogKeyGraceTestMargin = 50 * time.Millisecond

// executeCmd runs a command with a short timeout and returns its message. It fails the test on timeout.
func executeCmd(t *testing.T, command tea.Cmd) tea.Msg {
	t.Helper()

	if command == nil {
		return nil
	}

	result := make(chan tea.Msg, 1)
	go func() { result <- command() }()

	select {
	case message := <-result:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for command to produce a message")
		return nil
	}
}

func plantAudioFiles(t *testing.T, dir string, count int) {
	t.Helper()

	for index := range count {
		path := filepath.Join(dir, "track"+strconv.Itoa(index)+".mp3")
		require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	}
}

// initLogContains reports whether any log line in the initialization model contains substring.
func initLogContains(model *UIModel, substring string) bool {
	for _, text := range model.initializationModel.LogLines() {
		if strings.Contains(text, substring) {
			return true
		}
	}

	return false
}

func initLastLog(model *UIModel) (string, initializing.LogType) {
	texts := model.initializationModel.LogLines()
	if len(texts) == 0 {
		return "", initializing.LogNormal
	}

	return texts[len(texts)-1], model.initializationModel.LastLogType()
}

func assertCmdDoesNotChainLibraryCache(t *testing.T, command tea.Cmd) {
	t.Helper()

	if command == nil {
		return
	}

	result := make(chan tea.Msg, 1)
	go func() { result <- command() }()

	select {
	case message := <-result:
		_, isCacheLoad := message.(initializationLoadLibraryCacheResultMessage)
		assert.False(t, isCacheLoad, "user origin chained the init library cache load")
	case <-time.After(100 * time.Millisecond):
		// Blocked on the expire tick: the expected behavior.
	}
}

func TestConfigLoadedInitOriginError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadError := errors.New("[config:Load] parse config file: boom")
	_, command := model.Update(configLoadedMessage{
		config:           &config.Config{},
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              loadError,
	})

	assert.Nil(t, command, "returned cmd should be nil on error")
	assert.True(t, initLogContains(model, loadError.Error()), "log missing error %q", loadError.Error())

	text, logType := initLastLog(model)
	assert.Equal(t, initializing.LogError, logType, "last log type = %v, want LogError (text: %q)", logType, text)
}

func TestConfigLoadedInitOriginDefaults(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	defaultsConfig := config.Defaults()
	defaultsConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &defaultsConfig,
		isConfigDefaults: true,
		origin:           configLoadOriginInit,
		err:              nil,
	})

	assert.True(t, initLogContains(model, "no config file found, loading one using defaults"))

	for _, text := range model.initializationModel.LogLines() {
		if strings.Contains(text, "[keymap:New]") || strings.Contains(text, "falling back") {
			t.Errorf("unexpected keymap error log on defaults: %q", text)
		}
	}

	assert.NotNil(t, command, "returned cmd = nil, want non-nil (libraries present)")
}

func TestConfigLoadedInitOriginInvalidLibraryPaths(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	customConfig := config.Defaults()
	customConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:              &customConfig,
		isConfigDefaults:    false,
		invalidLibraryPaths: []string{"/this/path/does/not/exist"},
		err:                 nil,
	})

	assert.True(t, initLogContains(model, "invalid path found (╥﹏╥): /this/path/does/not/exist"))

	assert.NotNil(t, command, "returned cmd = nil, want non-nil (libraries present)")
}

func TestConfigLoadedInitOriginHappyPath(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	customConfig := config.Defaults()
	customConfig.Theme.Surface = "#ff0000"
	customConfig.Keybinds.MoveLeft = []string{"j"}
	customConfig.Keybinds.MoveRight = []string{"k"}
	customConfig.Keybinds.Select = []string{"enter"}
	customConfig.Keybinds.Quit = []string{"q"}
	customConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &customConfig,
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              nil,
	})

	assert.Equal(t, customConfig, *model.config)

	wantTheme := theme.New(customConfig.Theme)
	assert.Equal(t, wantTheme, model.theme)

	wantKeyMap, err := keymap.New(customConfig.Keybinds)
	require.NoError(t, err)

	assert.Equal(t, wantKeyMap, model.keyMap)
	assert.NotEqual(t, initialKeyMap, model.keyMap, "model.keyMap unchanged after loading custom keybinds")

	assert.True(t, initLogContains(model, "config loaded successfully"))
	assert.True(t, initLogContains(model, "theme loaded successfully"))
	assert.True(t, initLogContains(model, "keybindings loaded successfully"))
	assert.False(t, initLogContains(model, "no config file found"))

	assert.NotNil(t, command, "returned cmd = nil, want non-nil (libraries present)")
}

func TestConfigLoadedInitOriginKeymapParseFailure(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	badConfig := config.Defaults()
	badConfig.Keybinds.Select = []string{}
	badConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &badConfig,
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              nil,
	})

	assert.True(t, initLogContains(model, "[keymap:New]"))
	assert.True(t, initLogContains(model, "keybindings failed to load, fallbacking to previous bindings"))

	assert.Equal(t, initialKeyMap, model.keyMap, "model.keyMap changed on parse failure")
	assert.True(t, model.initializationModel.IsConfigError(), "expected modeConfigError on keymap parse failure")
	assert.Nil(t, command, "returned cmd = nil, want nil on keymap parse failure")
}

func TestConfigLoadedInitOriginNoLibrariesErrorsOut(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	emptyConfig := config.Defaults()
	emptyConfig.LibrariesPaths = []string{}

	_, command := model.Update(configLoadedMessage{
		config:           &emptyConfig,
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              nil,
	})

	require.NotNil(t, command, "returned cmd = nil, want a cache load command after config")

	_, command = model.Update(initializationLoadLibraryCacheResultMessage{
		library: audio.NewLibrary(),
		err:     nil,
	})

	assert.Nil(t, command, "returned cmd should be nil on empty cache with no library paths")
	assert.True(t, initLogContains(model, "no library paths found"))
	assert.True(t, model.initializationModel.IsConfigError(), "expected modeConfigError when no library paths")

	text, logType := initLastLog(model)
	assert.Equal(t, initializing.LogError, logType, "last log type = %v, want LogError (text: %q)", logType, text)
}

func TestConfigLoadedInitOriginLibrariesEmitsCacheLoad(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              nil,
	})

	require.NotNil(t, command, "returned cmd = nil, want a cache load command")

	message := executeCmd(t, command)
	_, ok := message.(initializationLoadLibraryCacheResultMessage)
	require.True(t, ok, "cmd produced %T, want initializationLoadLibraryCacheResultMessage", message)
}

func TestConfigLoadedInitOriginEmptyLibraryCacheWarnsUser(t *testing.T) {
	// The cache load command reads the real user config dir, so we set it like this.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)

	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		origin:           configLoadOriginInit,
		err:              nil,
	})
	require.NotNil(t, command)

	message := executeCmd(t, command)
	_, command = model.Update(message)

	assert.Nil(t, command, "returned cmd should be nil on empty cache")
	assert.True(t, initLogContains(model, "no songs found, you might want to discover them later~"))
	assert.False(
		t,
		model.initializationModel.IsConfigError(),
		"empty cache with library paths is not a config error",
	)
	assert.Equal(t, uiInitializing, model.state, "empty cache with library paths should land on initialization")
}

func TestHandleUserConfigLoadedHappyPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	customConfig := config.Defaults()
	customConfig.Theme.Surface = "#ff0000"
	customConfig.Keybinds.MoveLeft = []string{"j"}
	customConfig.Keybinds.MoveRight = []string{"k"}
	customConfig.Keybinds.Select = []string{"enter"}
	customConfig.Keybinds.Quit = []string{"q"}
	customConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &customConfig,
		isConfigDefaults: false,
		err:              nil,
		origin:           configLoadOriginUser,
	})

	// The only cmd Update may return here is the notification expiry tick, not the init cache-load chaining.
	assertCmdDoesNotChainLibraryCache(t, command)

	assert.Equal(t, customConfig, *model.config)

	wantTheme := theme.New(customConfig.Theme)
	assert.Equal(t, wantTheme, model.theme)

	wantKeyMap, err := keymap.New(customConfig.Keybinds)
	require.NoError(t, err)
	assert.Equal(t, wantKeyMap, model.keyMap)
	assert.NotEqual(t, initialKeyMap, model.keyMap, "model.keyMap unchanged after loading custom keybinds")

	// Feedback is push notifications only: no init log, no config error state.
	assert.True(t, model.notificationModel.HasActiveNotifications(), "expected success notification for user reload")
	assert.False(t, initLogContains(model, "config loaded successfully"), "user reload must not log to init screen")
	assert.False(t, model.initializationModel.IsConfigError(), "user reload must not set config error")
	assert.Equal(t, uiBootstrapping, model.state, "user reload must not change state")
}

func TestHandleUserConfigLoadedErrorKeepsPreviousState(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	initialKeyMap := model.keyMap
	previousConfig := *model.config

	loadError := errors.New("[config:Load] parse config file: boom")
	_, command := model.Update(configLoadedMessage{
		config:           &config.Config{},
		isConfigDefaults: false,
		err:              loadError,
		origin:           configLoadOriginUser,
	})

	// The error notification queues an expiry cmd, so assert it is not a cache load and check user-visible feedback.
	assertCmdDoesNotChainLibraryCache(t, command)

	assert.True(t, model.notificationModel.HasActiveNotifications(), "expected error notification for user reload")
	assert.False(t, model.initializationModel.IsConfigError(), "user reload error must not set config error")
	assert.Equal(t, previousConfig, *model.config, "config must stay unchanged on load error")
	assert.Equal(t, initialKeyMap, model.keyMap, "keymap must stay unchanged on load error")
}

func TestHandleUserConfigLoadedKeymapFailureKeepsPreviousKeymap(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	badConfig := config.Defaults()
	badConfig.Keybinds.Select = []string{}
	badConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(configLoadedMessage{
		config:           &badConfig,
		isConfigDefaults: false,
		err:              nil,
		origin:           configLoadOriginUser,
	})

	// The failure notification queues an expiry cmd, so assert it is not a cache load and check the kept keymap.
	assertCmdDoesNotChainLibraryCache(t, command)

	// Previous keymap is kept so the session stays usable; error is reported via notification only.
	assert.Equal(t, initialKeyMap, model.keyMap, "keymap must stay unchanged on keymap failure")
	assert.True(t, model.notificationModel.HasActiveNotifications(), "expected keymap failure notification")
	assert.False(t, model.initializationModel.IsConfigError(), "user reload keymap failure must not set config error")

	// The theme from the new config was already applied before the keymap failure.
	wantTheme := theme.New(badConfig.Theme)
	assert.Equal(t, wantTheme, model.theme)
}

func TestHandleUserConfigLoadedInvalidPathsNotifies(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)

	customConfig := config.Defaults()
	customConfig.LibrariesPaths = []string{t.TempDir()}

	model.Update(configLoadedMessage{
		config:              &customConfig,
		isConfigDefaults:    false,
		invalidLibraryPaths: []string{"/this/path/does/not/exist"},
		err:                 nil,
		origin:              configLoadOriginUser,
	})

	// The warning notification queues an expiry cmd, so only assert the user-visible feedback.
	assert.True(
		t,
		model.notificationModel.HasActiveNotifications(),
		"invalid paths must notify the user (config.Load prunes them silently)",
	)
}

func TestConfigLoadDoubleTriggerGuard(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	model.state = uiPlaylist

	// The first reload starts a load and arms the guard, while the second must be rejected without touching anything.
	firstCommand := model.handleComponentAction(action.ReloadConfigAction{})
	assert.NotNil(t, firstCommand, "first reload should return a load cmd")
	assert.True(t, model.isConfigLoading, "first reload must arm the in-flight guard")

	command := model.handleComponentAction(action.ReloadConfigAction{})

	assert.Nil(t, command, "reload while a config load is in flight must be ignored")
	assert.True(t, model.isConfigLoading, "in-flight flag must stay set")
}

func TestConfigLoadGuardLifecycle(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	model.state = uiPlaylist

	// A successful user reload arms the flag.
	reloadCommand := model.handleComponentAction(action.ReloadConfigAction{})
	assert.NotNil(t, reloadCommand, "first reload should return a load cmd")
	assert.True(t, model.isConfigLoading, "reload must arm the in-flight guard")

	// The load cmd itself is not executed (it touches the real config dir).
	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, _ = model.Update(configLoadedMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		err:              nil,
		origin:           configLoadOriginUser,
	})
	assert.False(t, model.isConfigLoading, "arriving configLoadedMessage must reset the guard")
}

func TestHandleLoadedLibraryCachePopulatesLibrary(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	cache := map[string]*audio.AudioFile{
		"/music/boa/fool.flac": {Path: "/music/boa/fool.flac", Title: "Fool", Artist: "bôa", Album: "Twilight"},
		"/music/other.mp3":     {Path: "/music/other.mp3", Title: "Other"},
	}
	library := audio.NewLibrary()
	library.File = cache

	_, command := model.Update(initializationLoadLibraryCacheResultMessage{
		library: library,
		err:     nil,
	})

	assert.Nil(t, command, "no follow-up cmd is expected after loading the cache")
	assert.Equal(t, uiPlaylist, model.state, "a non-empty cache should skip discovery and land on the playlist")
	assert.False(
		t,
		model.notificationModel.HasActiveNotifications(),
		"a successful cache load should not notify the user",
	)
	assert.Equal(t, cache, model.library.File, "library files should come straight from the cache")
	assert.NotEmpty(t, model.library.ByArtist, "artist index should be built after a cache load")
	assert.NotEmpty(t, model.library.ByAlbum, "album index should be built after a cache load")
}

func TestHandleErroredLibraryCacheWarnsUserAndFallsThrough(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.config.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadLibraryCacheResultMessage{
		library: audio.NewLibrary(),
		err:     fmt.Errorf("[audio:LoadCache] read cache file: boom"),
	})

	// PushNotification queues a notification expiry cmd, which is batched into the returned command.
	assert.NotNil(t, command)
	assert.True(t, model.notificationModel.HasActiveNotifications(), "a failed cache read should notify the user")
	assert.Equal(
		t,
		uiInitializing,
		model.state,
		"a failed cache read should land on the initialization screen",
	)
	assert.False(
		t,
		model.initializationModel.IsConfigError(),
		"a failed cache read with library paths is not a config error",
	)
}

func TestHandleDiscoverFilesStartMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 5

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	message := discoverFilesStartMessage{
		progress:        audio.NewDiscoveryProgress(),
		discoveryCancel: cancel,
		generation:      7,
	}

	_, command := model.Update(message)

	require.NotNil(t, model.libraryDiscoveryCancel, "discoveryCancel = nil, want the cancel func")
	assert.Equal(t, uint64(7), model.libraryDiscoveryGeneration)
	require.NotNil(t, command, "returned cmd = nil, want the progress tick cmd")
}

func TestHandleDiscoveryProgressTickMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 3
	model.libraryStatsModel.StartDiscovery()
	model.libraryDiscoveryCancel = func() {}

	progress := audio.NewDiscoveryProgress()
	progress.AddDiscovered(42)

	_, command := model.Update(discoveryProgressTickMessage{
		progress:   progress,
		generation: 3,
	})

	assert.NotNil(t, command, "returned cmd = nil, want the next tick cmd")
	assert.Equal(
		t,
		42,
		model.libraryStatsModel.DiscoveredFilesCount(),
		"ticked discovered count should reach the component",
	)

	// A stale generation must not update the display nor re-arm the tick chain.
	progressStale := audio.NewDiscoveryProgress()
	progressStale.AddDiscovered(99)
	_, command = model.Update(discoveryProgressTickMessage{
		progress:   progressStale,
		generation: 2,
	})
	assert.Nil(t, command, "stale tick should not re-arm the tick chain")
}

func TestHandleDiscoverFilesResultMessageStaleGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	initialLogCount := len(model.initializationModel.LogLines())

	_, command := model.Update(discoverFilesResultMessage{
		library:    nil,
		err:        nil,
		generation: 5,
	})

	assert.Nil(t, command, "returned cmd should be nil on stale generation")
	assert.False(t, sentinelCanceled, "stale result cleared discoveryCancel")

	assert.NotNil(
		t,
		model.libraryDiscoveryCancel,
		"discoveryCancel = nil, want unchanged sentinel on stale generation",
	)

	assert.Len(t, model.initializationModel.LogLines(), initialLogCount, "logCount changed on stale")
}

func TestHandleDiscoverFilesResultMessageCurrentGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	_, command := model.Update(discoverFilesResultMessage{
		library:    audio.NewLibrary(),
		progress:   audio.NewDiscoveryProgress(),
		err:        nil,
		generation: 10,
	})

	assert.NotNil(t, command, "returned cmd should be the next-step parse cmd on current generation result")
	assert.False(t, sentinelCanceled, "current result should replace discoveryCancel, not call it")
	require.NotNil(
		t,
		model.libraryDiscoveryCancel,
		"discoveryCancel = nil, want the parse cancel func registered before dispatch",
	)

	// The parse phase continues under the pipeline's generation.
	assert.Equal(t, uint64(10), model.libraryDiscoveryGeneration)
}

func TestHandleDiscoverFilesResultMessageError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10

	discoveryError := errors.New("[audio:DiscoverFiles] walk failed")
	_, command := model.Update(discoverFilesResultMessage{
		library:    nil,
		err:        discoveryError,
		generation: 10,
	})

	assert.Nil(t, command, "returned cmd should be nil on error result")
	assert.Empty(t, model.initializationModel.LogLines(), "discovery errors should not log")
}

func TestHandleWindowResize(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	assert.Equal(t, 100, model.windowWidth)
	assert.Equal(t, 40, model.windowHeight)
	assert.Nil(t, command, "returned cmd should be nil for window resize")
}

func TestHandleKeyPressMsgQuit(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'd'})

	require.True(t, isTeaQuit(command), "returned cmd is not tea.Quit")
	assert.True(t, sentinelCanceled, "cancelCurrentLibraryDiscovery was not called on quit")
	assert.Nil(t, model.libraryDiscoveryCancel, "discoveryCancel = non-nil, want nil after quit")
}

func TestHandleKeyPressMsgQuitDoesNotMatchNonQuitKey(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	assert.Nil(t, command, "returned cmd should be nil for unmatched key")
}

func TestHandleKeyPressMsgForwardsToComponentReload(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.initializationModel.SetConfigError()
	model.setState(uiInitializing)

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, command, "returned cmd = nil, want non-nil for reload action")

	assert.True(t, sentinelCanceled, "cancelCurrentLibraryDiscovery was not called on reload")
	assert.Equal(t, uiInitializing, model.state, "reload should stay on the initialization screen")
	assert.True(t, initLogContains(model, "reloading config..."))
}

func TestHandleKeyPressMsgForwardsToComponentProceed(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.setState(uiInitializing)

	model.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, command, "returned cmd should be nil for proceed action")

	assert.True(t, sentinelCanceled, "cancelCurrentLibraryDiscovery was not called on proceed")
	assert.Equal(t, uiPlaylist, model.state)
	assert.True(t, initLogContains(model, "proceeding without libraries"))
}

func TestHandleKeyPressMsgMoveLeftIsNoOp(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	assert.Nil(t, command, "returned cmd should be nil for move left (NoAction)")
}

func TestHandleKeyPressMsgWhichKeyRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sequence     []tea.KeyPressMsg
		wantHidden   bool
		wantState    uiState
		wantStateSet bool
	}{
		{
			name:         "open actions opens the card and maps playlist key to playlist state",
			sequence:     []tea.KeyPressMsg{{Code: ' '}, {Code: 'p'}},
			wantHidden:   true,
			wantState:    uiPlaylist,
			wantStateSet: true,
		},
		{
			name:         "open actions opens the card and maps library stats key to library stats state",
			sequence:     []tea.KeyPressMsg{{Code: ' '}, {Code: 'l'}},
			wantHidden:   true,
			wantState:    uiLibraryStats,
			wantStateSet: true,
		},
		{
			name:       "open actions opens the card and go back closes it keeping the state",
			sequence:   []tea.KeyPressMsg{{Code: ' '}, {Code: tea.KeyEscape}},
			wantHidden: true,
			wantState:  uiPlaylist,
		},
		{
			name:       "unmapped key while card is open closes it and does nothing",
			sequence:   []tea.KeyPressMsg{{Code: ' '}, {Code: 'x', Text: "x"}},
			wantHidden: true,
			wantState:  uiPlaylist,
		},
		{
			name:       "open actions while card is open closes it keeping the state",
			sequence:   []tea.KeyPressMsg{{Code: ' '}, {Code: ' '}},
			wantHidden: true,
			wantState:  uiPlaylist,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := newTestUI(t)
			model.setState(uiPlaylist)

			for _, keyPress := range test.sequence {
				model.Update(keyPress)
			}

			assert.Equal(t, test.wantHidden, !model.whichkeyModel.IsVisible(), "whichkey card visibility mismatch")

			if test.wantStateSet {
				assert.Equal(t, test.wantState, model.state)
			}
		})
	}
}

func TestHandleKeyPressMsgWhichKeyDoesNotSinkIntoInitializing(t *testing.T) {
	t.Parallel()

	// During bootstrapping (pre-initialization) keys are swallowed entirely.
	model := newTestUI(t)
	require.Equal(t, uiBootstrapping, model.state)

	_, command := model.Update(tea.KeyPressMsg{Code: ' ', Text: " "})

	assert.Nil(t, command, "returned cmd should be nil while bootstrapping")
	assert.False(t, model.whichkeyModel.IsVisible(), "whichkey must not capture keys while bootstrapping")

	// Once the initialization screen is confirmed, whichkey stays out of it too.
	model.setState(uiInitializing)

	_, command = model.Update(tea.KeyPressMsg{Code: ' ', Text: " "})

	assert.Nil(t, command, "returned cmd should be nil while initializing")
	assert.False(t, model.whichkeyModel.IsVisible(), "whichkey must not capture keys during initialization")
}

func TestHandleKeyPressMsgDialogGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		prepare        func(model *UIModel)
		keyPress       tea.KeyPressMsg
		wantDialogOpen bool
		wantDiscovery  bool
	}{
		{
			name: "key inside the grace quiet period is dropped",
			prepare: func(model *UIModel) {
				model.setState(uiLibraryStats)
				model.handleComponentAction(action.OpenConfirmDialogAction{
					Text:          "question?",
					ConfirmAction: action.DiscoverLibraryFullAction{},
				})
			},
			keyPress:       tea.KeyPressMsg{Code: 'h', Text: "h"},
			wantDialogOpen: true,
			wantDiscovery:  false,
		},
		{
			name: "key after the grace quiet period reaches the dialog",
			prepare: func(model *UIModel) {
				model.setState(uiLibraryStats)
				model.handleComponentAction(action.OpenConfirmDialogAction{
					Text:          "question?",
					ConfirmAction: action.DiscoverLibraryFullAction{},
				})
				time.Sleep(dialog.KeyGraceQuietPeriod + dialogKeyGraceTestMargin)
			},
			keyPress:       tea.KeyPressMsg{Code: 'h', Text: "h"},
			wantDialogOpen: true,
			wantDiscovery:  false,
		},
		{
			name: "go back closes the dialog without dispatching",
			prepare: func(model *UIModel) {
				model.setState(uiLibraryStats)
				model.handleComponentAction(action.OpenConfirmDialogAction{
					Text:          "question?",
					ConfirmAction: action.DiscoverLibraryFullAction{},
				})
				time.Sleep(dialog.KeyGraceQuietPeriod + dialogKeyGraceTestMargin)
			},
			keyPress:       tea.KeyPressMsg{Code: tea.KeyEscape},
			wantDialogOpen: false,
			wantDiscovery:  false,
		},
		{
			name: "keys while the dialog is open do not reach the whichkey overlay",
			prepare: func(model *UIModel) {
				model.setState(uiLibraryStats)
				model.handleComponentAction(action.OpenConfirmDialogAction{
					Text:          "question?",
					ConfirmAction: action.DiscoverLibraryFullAction{},
				})
				time.Sleep(dialog.KeyGraceQuietPeriod + dialogKeyGraceTestMargin)
				model.whichkeyModel.SetBindings(model.commandBindingsFor(uiLibraryStats))
				model.whichkeyModel.HandleMessage(tea.KeyPressMsg{Code: ' '})
			},
			keyPress:       tea.KeyPressMsg{Code: 'x', Text: "x"},
			wantDialogOpen: true,
			wantDiscovery:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := newTestUI(t)
			test.prepare(model)
			whichKeyVisibleBefore := model.whichkeyModel.IsVisible()

			_, command := model.Update(test.keyPress)

			assert.Equal(t, test.wantDialogOpen, model.dialogModel.IsOpen(), "dialog visibility mismatch")
			assert.Equal(
				t,
				whichKeyVisibleBefore,
				model.whichkeyModel.IsVisible(),
				"whichkey visibility must not change while the dialog gate holds the key",
			)

			if test.wantDiscovery {
				require.NotNil(t, command, "confirm must return the discovery start cmd")

				startMessage := executeCmd(t, command)
				_, isDiscoveryStart := startMessage.(discoverFilesStartMessage)
				assert.True(t, isDiscoveryStart, "confirm must dispatch the discovery pipeline start")
			} else {
				assert.Nil(t, command, "returned cmd = non-nil, want nil")
			}
		})
	}
}

func TestHandleKeyPressMsgDialogConfirmDispatchesDiscovery(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.setState(uiLibraryStats)
	model.handleComponentAction(action.OpenConfirmDialogAction{
		Text:          "question?",
		ConfirmAction: action.DiscoverLibraryFullAction{},
	})

	// Cross the key grace quiet period (which absorbs the open keypress's auto-repeat tail), then move the cursor to
	// the confirm button before confirming.
	time.Sleep(dialog.KeyGraceQuietPeriod + dialogKeyGraceTestMargin)
	model.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, command, "confirm must return the discovery start cmd")
	assert.False(t, model.dialogModel.IsOpen(), "confirm must close the dialog")

	startMessage := executeCmd(t, command)
	start, isDiscoveryStart := startMessage.(discoverFilesStartMessage)
	require.True(t, isDiscoveryStart, "confirm must dispatch the discovery pipeline start")

	// The dispatched action runs the same pipeline as a direct DiscoverLibraryFullAction: it starts from scratch under
	// the first generation, carrying the cancel func the root registers when the start message is handled.
	require.NotNil(t, start.discoveryCancel, "discovery start must carry the cancel func")
	assert.Equal(t, uint64(1), start.generation, "confirm must start the first discovery generation")
}

func TestHandleKeyPressMsgQuitWhileDialogOpen(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.setState(uiLibraryStats)
	model.handleComponentAction(action.OpenConfirmDialogAction{
		Text:          "question?",
		ConfirmAction: action.DiscoverLibraryFullAction{},
	})

	_, command := model.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'd'})

	require.True(t, isTeaQuit(command), "quit must stay reachable while the dialog is open")
}

func TestHandleComponentAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		action        action.Action
		wantQuit      bool
		wantCmdNonNil bool
		wantState     uiState
		wantStateSet  bool
		wantLogSubstr string
		wantCanceled  bool
		wantCleared   bool
		skipLogCheck  bool
	}{
		{
			name:         "nil action returns nil",
			action:       nil,
			skipLogCheck: true,
		},
		{
			name:         "NoAction returns nil",
			action:       action.NoAction{},
			skipLogCheck: true,
		},
		{
			name:         "QuitAction returns tea.Quit and clears cancel",
			action:       action.QuitAction{},
			wantQuit:     true,
			wantCanceled: true,
			wantCleared:  true,
			skipLogCheck: true,
		},
		{
			name:          "ReloadConfigAction returns load cmd and keeps the current state",
			action:        action.ReloadConfigAction{},
			wantCmdNonNil: true,
			wantCanceled:  true,
			wantCleared:   true,
			skipLogCheck:  true,
		},
		{
			name:          "DiscoverLibraryFullAction returns discovery start cmd",
			action:        action.DiscoverLibraryFullAction{},
			wantCmdNonNil: true,
			wantCanceled:  true,
			wantCleared:   true,
			skipLogCheck:  true,
		},
		{
			name:          "ProceedFromInitAction returns nil and goes to playlist",
			action:        action.ProceedFromInitAction{},
			wantState:     uiPlaylist,
			wantStateSet:  true,
			wantLogSubstr: "proceeding without libraries",
			wantCanceled:  true,
			wantCleared:   true,
		},
		{
			name:         "OpenPlaylistAction returns nil and goes to playlist",
			action:       action.OpenPlaylistAction{},
			wantState:    uiPlaylist,
			wantStateSet: true,
			skipLogCheck: true,
		},
		{
			name:         "OpenLibraryStatsAction returns nil and goes to library stats",
			action:       action.OpenLibraryStatsAction{},
			wantState:    uiLibraryStats,
			wantStateSet: true,
			skipLogCheck: true,
		},
		{
			name:         "ActionCommand returns the carried cmd",
			action:       action.ActionCommand{Command: tea.Quit},
			wantQuit:     true,
			skipLogCheck: true,
		},
		{
			name: "OpenConfirmDialogAction opens the dialog and returns nil",
			action: action.OpenConfirmDialogAction{
				Text:          "question?",
				ConfirmAction: action.DiscoverLibraryFullAction{},
			},
			skipLogCheck: true,
		},
		{
			name:         "unknown action returns nil",
			action:       struct{}{},
			skipLogCheck: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := newTestUI(t)

			sentinelCanceled := false
			model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

			command := model.handleComponentAction(test.action)

			if test.wantQuit {
				require.True(t, isTeaQuit(command), "returned cmd is not tea.Quit")
			} else if test.wantCmdNonNil {
				require.NotNil(t, command, "returned cmd = nil, want non-nil")
			} else {
				assert.Nil(t, command, "returned cmd = non-nil, want nil")
			}

			if _, isOpenDialog := test.action.(action.OpenConfirmDialogAction); isOpenDialog {
				assert.True(t, model.dialogModel.IsOpen(), "OpenConfirmDialogAction must open the dialog")
			}

			if test.wantCanceled {
				assert.True(t, sentinelCanceled, "cancelCurrentLibraryDiscovery was not called")
			}

			if test.wantCleared {
				assert.Nil(
					t,
					model.libraryDiscoveryCancel,
					"discoveryCancel = non-nil, want nil after action",
				)
			}

			if test.wantStateSet {
				assert.Equal(t, test.wantState, model.state)
			}

			if !test.skipLogCheck {
				assert.True(
					t,
					initLogContains(model, test.wantLogSubstr),
					"missing log line containing %q",
					test.wantLogSubstr,
				)
			}
		})
	}
}

func TestDiscoverFilesStartCommandAsync(t *testing.T) {
	t.Parallel()

	libraryDir := t.TempDir()
	plantAudioFiles(t, libraryDir, 5)

	contextForDiscovery, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startCommand := discoverFilesStartCommand(contextForDiscovery, 1, []string{libraryDir}, audio.NewLibrary())
	startMessage := executeCmd(t, startCommand).(discoverFilesStartMessage)

	// The start message carries its own progress reporter, and the discovery goroutine reports into it.
	progress := startMessage.progress
	deadline := time.After(3 * time.Second)
	for {
		if progress.DiscoveryDone() && progress.DiscoveredCount() == 5 {
			break
		}

		select {
		case <-deadline:
			t.Fatalf(
				"timed out waiting for discovery: discovered=%d done=%v",
				progress.DiscoveredCount(),
				progress.DiscoveryDone(),
			)
		default:
			runtime.Gosched()
		}
	}

	tickedModel := newTestUI(t)
	tickedModel.libraryDiscoveryGeneration = 1
	tickedModel.libraryStatsModel.StartDiscovery()
	tickedModel.libraryDiscoveryCancel = func() {}

	_, command := tickedModel.Update(discoveryProgressTickMessage{progress: progress, generation: 1})

	require.NotNil(t, command, "tick should re-arm while the discovery is current")
	assert.Equal(t, 5, tickedModel.libraryStatsModel.DiscoveredFilesCount(), "tick should surface the discovered count")
}

func TestParseFilesMetatagStartCommandAsync(t *testing.T) {
	t.Parallel()

	files := []*audio.AudioFile{{Path: "/this/path/does/not/exist.flac"}}

	parseContext, parseCancel := context.WithCancel(context.Background())
	t.Cleanup(parseCancel)

	startCommand := parseFilesMetatagStartCommand(parseContext, 1, files, audio.NewDiscoveryProgress())

	message := executeCmd(t, startCommand)
	startMessage, ok := message.(metatagParseStartMessage)
	require.True(t, ok, "cmd produced %T, want metatagParseStartMessage", message)

	require.NotNil(t, startMessage.progress)
	assert.Equal(t, uint64(1), startMessage.generation)

	// The parse goroutine reports into the start message's progress reporter.
	deadline := time.After(3 * time.Second)
	for {
		if startMessage.progress.ParsedCount() == 1 {
			break
		}

		select {
		case <-deadline:
			t.Fatalf("timed out waiting for parse: parsed=%d", startMessage.progress.ParsedCount())
		default:
			runtime.Gosched()
		}
	}

	assert.Equal(t, 1, startMessage.progress.ParsedCount(), "failed parses should still count as attempted")
}

func TestHandleMetatagParseStartMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 4
	model.libraryStatsModel.StartDiscovery()

	_, command := model.Update(metatagParseStartMessage{
		progress:   audio.NewDiscoveryProgress(),
		generation: 4,
	})

	require.NotNil(t, command, "returned cmd = nil, want the progress tick cmd")
}

func TestHandleMetatagParseStartMessageStaleGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10
	model.libraryStatsModel.StartDiscovery()

	_, command := model.Update(metatagParseStartMessage{
		progress:   audio.NewDiscoveryProgress(),
		generation: 5,
	})

	assert.Nil(t, command, "stale parse start should be rejected")
	assert.Equal(t, uint64(10), model.libraryDiscoveryGeneration, "stale parse start moved the generation")
}

func TestHandleMetatagParseResultMessageSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 4
	model.libraryStatsModel.StartDiscovery()
	model.libraryStatsModel.SetProgress(testutil.NewDiscoveryProgress(7, 7, true))
	model.libraryDiscoveryCancel = func() {}

	command := model.handleMetatagParseResultMessage(metatagParseResultMessage{
		generation: 4,
	})

	assert.Nil(t, command)
	assert.Nil(t, model.libraryDiscoveryCancel, "discoveryCancel = non-nil, want nil after result")

	// Success notification pushed.
	assert.True(
		t,
		model.notificationModel.HasActiveNotifications(),
		"expected an active notification after discovery success",
	)

	// The library indexes are built synchronously in the handler.
	assert.Empty(t, model.library.ByArtist, "no files were parsed, indexes should stay empty")

	populatedLibrary := audio.NewLibrary()
	populatedLibrary.Add("/some/path.flac", 1)
	populatedLibrary.File["/some/path.flac"].Artist = "bôa"
	model.library = populatedLibrary
	model.handleMetatagParseResultMessage(metatagParseResultMessage{
		generation: 4,
	})

	assert.Len(t, populatedLibrary.ByArtist["bôa"], 1, "expected artist index entry after the rebuild")
}

func TestHandleMetatagParseResultMessageStaleGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10
	model.libraryStatsModel.StartDiscovery()
	model.libraryStatsModel.SetProgress(testutil.NewDiscoveryProgress(7, 7, true))
	model.libraryDiscoveryCancel = func() {}

	command := model.handleMetatagParseResultMessage(metatagParseResultMessage{
		generation: 5,
	})

	assert.Nil(t, command)
	assert.NotNil(t, model.libraryDiscoveryCancel, "stale result cleared discoveryCancel")

	rendered := testutil.StripANSI(model.libraryStatsModel.Render(80, 24))
	assert.Contains(t, rendered, "audio files", "stale result should not clear discovery state")
}

func TestHandleMetatagParseResultMessageError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 4
	model.libraryStatsModel.StartDiscovery()
	model.libraryDiscoveryCancel = func() {}

	initialLogCount := len(model.initializationModel.LogLines())

	command := model.handleMetatagParseResultMessage(metatagParseResultMessage{
		err:        context.Canceled,
		generation: 4,
	})

	assert.Nil(t, command)
	assert.Len(t, model.initializationModel.LogLines(), initialLogCount, "cancellation should not log an error")
}

func TestHandleDiscoverFilesResultMessageKeepsParseActiveUntilDone(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10
	model.libraryStatsModel.StartDiscovery()
	sharedProgress := testutil.NewDiscoveryProgress(3, 0, false)
	model.libraryStatsModel.SetProgress(sharedProgress)

	library := audio.NewLibrary()
	library.Add("/this/path/does/not/exist.flac", 0)

	_, command := model.Update(discoverFilesResultMessage{
		library:    library,
		progress:   sharedProgress,
		err:        nil,
		generation: 10,
	})

	require.NotNil(t, command, "returned cmd should be the metatag parse start cmd")

	require.NotNil(t, model.libraryDiscoveryCancel, "discoveryCancel = nil, want the parse cancel func before dispatch")
	assert.Equal(
		t,
		uint64(10),
		model.libraryDiscoveryGeneration,
		"the parse phase continues under the pipeline's generation",
	)

	// The parse start message re-arms the tick chain and flips the status to the found+parsing pair.
	_, command = model.Update(metatagParseStartMessage{
		progress:   sharedProgress,
		generation: 10,
	})
	require.NotNil(t, command, "returned cmd = nil, want the progress tick cmd")

	rendered := testutil.StripANSI(model.libraryStatsModel.Render(80, 24))

	assert.Contains(t, rendered, "3 audio files have been found", "render output:\n%s", rendered)
	assert.Contains(t, rendered, "parsing 0/3 files", "render output:\n%s", rendered)
}

func TestHandleDiscoverFilesResultMessageDoesNotLogErrors(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.libraryDiscoveryGeneration = 10
	model.libraryStatsModel.StartDiscovery()

	_, command := model.Update(discoverFilesResultMessage{
		library:    nil,
		progress:   audio.NewDiscoveryProgress(),
		err:        context.Canceled,
		generation: 10,
	})

	assert.Nil(t, command)
	assert.Empty(t, model.initializationModel.LogLines(), "cancellation should not log an error")
}
