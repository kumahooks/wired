package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

func isTeaQuit(command tea.Cmd) bool {
	if command == nil {
		return false
	}

	_, ok := command().(tea.QuitMsg)

	return ok
}

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

// runCmds drives the fetch drainer to completion. It executes the drainer command, feeds WaitProgress messages back
// into model.Update to get the next drainer, and returns the first fetchFilesResultMessage.
func runCmds(t *testing.T, model *UIModel, command tea.Cmd) tea.Msg {
	t.Helper()

	timeout := time.After(3 * time.Second)

	for command != nil {
		result := make(chan tea.Msg, 1)
		go func() { result <- command() }()

		select {
		case message := <-result:
			switch message := message.(type) {
			case fetchFilesResultMessage:
				return message
			case fetchFilesWaitProgressMessage:
				_, command = model.Update(message)
			default:
				t.Fatalf("unexpected message from drainer: %T", message)
			}
		case <-timeout:
			t.Fatal("timed out waiting for fetch result")
		}
	}

	return nil
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

func TestHandleInitializationLoadConfigResultError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadError := errors.New("[config:Load] parse config file: boom")
	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &config.Config{},
		isConfigDefaults: false,
		err:              loadError,
	})

	assert.Nil(t, command, "returned cmd should be nil on error")
	assert.True(t, initLogContains(model, loadError.Error()), "log missing error %q", loadError.Error())

	text, logType := initLastLog(model)
	assert.Equal(t, initializing.LogError, logType, "last log type = %v, want LogError (text: %q)", logType, text)
}

func TestHandleInitializationLoadConfigResultDefaults(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	defaultsConfig := config.Defaults()
	defaultsConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &defaultsConfig,
		isConfigDefaults: true,
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

func TestHandleInitializationLoadConfigResultInvalidLibraryPaths(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	customConfig := config.Defaults()
	customConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:              &customConfig,
		isConfigDefaults:    false,
		invalidLibraryPaths: []string{"/this/path/does/not/exist"},
		err:                 nil,
	})

	assert.True(t, initLogContains(model, "invalid path found (╥﹏╥): /this/path/does/not/exist"))

	assert.NotNil(t, command, "returned cmd = nil, want non-nil (libraries present)")
}

func TestHandleInitializationLoadConfigResultHappyPath(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	customConfig := config.Defaults()
	customConfig.Theme.Surface = "#ff0000"
	customConfig.Keybinds.MoveLeft = []string{"j"}
	customConfig.Keybinds.MoveRight = []string{"k"}
	customConfig.Keybinds.Select = []string{"space"}
	customConfig.Keybinds.Quit = []string{"q"}
	customConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &customConfig,
		isConfigDefaults: false,
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

func TestHandleInitializationLoadConfigResultKeymapParseFailure(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	initialKeyMap := model.keyMap

	badConfig := config.Defaults()
	badConfig.Keybinds.Select = []string{}
	badConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &badConfig,
		isConfigDefaults: false,
		err:              nil,
	})

	assert.True(t, initLogContains(model, "[keymap:New]"))
	assert.True(t, initLogContains(model, "falling back to default keybindings"))
	assert.True(t, initLogContains(model, "keybindings failed to load, using previously held bindings"))

	assert.Equal(t, initialKeyMap, model.keyMap, "model.keyMap changed on parse failure")
	assert.True(t, model.initializationModel.IsConfigError(), "expected modeConfigError on keymap parse failure")
	assert.Nil(t, command, "returned cmd = nil, want nil on keymap parse failure")
}

func TestHandleInitializationLoadConfigResultNoLibrariesErrorsOut(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	emptyConfig := config.Defaults()
	emptyConfig.LibrariesPaths = []string{}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &emptyConfig,
		isConfigDefaults: false,
		err:              nil,
	})

	require.NotNil(t, command, "returned cmd = nil, want a cache load command after config")

	_, command = model.Update(initializationLoadLibraryCacheResultMessage{
		library: Library{audioFiles: &[]audio.File{}},
		err:     nil,
	})

	assert.Nil(t, command, "returned cmd should be nil on empty cache with no library paths")
	assert.True(t, initLogContains(model, "no library paths found"))
	assert.True(t, model.initializationModel.IsConfigError(), "expected modeConfigError when no library paths")

	text, logType := initLastLog(model)
	assert.Equal(t, initializing.LogError, logType, "last log type = %v, want LogError (text: %q)", logType, text)
}

func TestHandleInitializationLoadConfigResultLibrariesEmitsCacheLoad(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		err:              nil,
	})

	require.NotNil(t, command, "returned cmd = nil, want a cache load command")

	message := executeCmd(t, command)
	_, ok := message.(initializationLoadLibraryCacheResultMessage)
	require.True(t, ok, "cmd produced %T, want initializationLoadLibraryCacheResultMessage", message)
}

func TestHandleEmptyLibraryCacheWarnsUser(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		err:              nil,
	})
	require.NotNil(t, command)

	message := executeCmd(t, command)
	_, command = model.Update(message)

	assert.Nil(t, command, "returned cmd should be nil on empty cache")
	assert.True(t, initLogContains(model, "no scanned songs found, you might want to scan them later"))
	assert.False(
		t,
		model.initializationModel.IsConfigError(),
		"empty cache with library paths should not be a config error",
	)
}

func TestHandleFetchFilesStartMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.library.scanGeneration = 5

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	progressChannel := make(chan int, 1)
	resultChannel := make(chan fetchFilesResultMessage, 1)

	message := fetchFilesStartMessage{
		progressChannel: progressChannel,
		resultChannel:   resultChannel,
		scanCancel:      cancel,
		generation:      7,
	}

	_, command := model.Update(message)

	require.NotNil(t, model.library.scanCancel, "scanCancel = nil, want the cancel func")

	assert.Equal(t, uint64(7), model.library.scanGeneration)

	require.NotNil(t, command, "returned cmd = nil, want the drainer cmd")
}

func TestHandleFetchFilesWaitProgressMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	progressChannel := make(chan int, 1)
	resultChannel := make(chan fetchFilesResultMessage, 1)

	_, command := model.Update(fetchFilesWaitProgressMessage{
		filesCount:      42,
		progressChannel: progressChannel,
		resultChannel:   resultChannel,
		generation:      3,
	})

	assert.NotNil(t, command, "returned cmd = nil, want the next drainer cmd")
}

func TestHandleFetchFilesResultMessageStaleGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.library.scanGeneration = 10

	sentinelCanceled := false
	model.library.scanCancel = func() { sentinelCanceled = true }

	initialLogCount := len(model.initializationModel.LogLines())

	_, command := model.Update(fetchFilesResultMessage{
		files:      nil,
		err:        nil,
		generation: 5,
	})

	assert.Nil(t, command, "returned cmd should be nil on stale generation")
	assert.False(t, sentinelCanceled, "stale result cleared scanCancel")

	assert.NotNil(
		t,
		model.library.scanCancel,
		"scanCancel = nil, want unchanged sentinel on stale generation",
	)

	assert.Len(t, model.initializationModel.LogLines(), initialLogCount, "logCount changed on stale")
}

func TestHandleFetchFilesResultMessageCurrentGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.library.scanGeneration = 10

	sentinelCanceled := false
	model.library.scanCancel = func() { sentinelCanceled = true }

	discoveredFiles := make([]audio.File, 137)

	_, command := model.Update(fetchFilesResultMessage{
		files:      discoveredFiles,
		err:        nil,
		generation: 10,
	})

	assert.NotNil(t, command, "returned cmd should be the next-step scan cmd on current generation result")
	assert.False(t, sentinelCanceled, "current result should set scanCancel to nil, not call it")
	assert.Nil(
		t,
		model.library.scanCancel,
		"scanCancel = non-nil, want nil after current generation result",
	)
	assert.Equal(t, discoveredFiles, *model.library.audioFiles)
}

func TestHandleFetchFilesResultMessageError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.library.scanGeneration = 10

	fetchError := errors.New("[audio:FetchFiles] walk failed")
	_, command := model.Update(fetchFilesResultMessage{
		files:      nil,
		err:        fetchError,
		generation: 10,
	})

	assert.Nil(t, command, "returned cmd should be nil on error result")

	text, logType := initLastLog(model)
	assert.Equal(t, initializing.LogError, logType, "last log type = %v, want LogError (text: %q)", logType, text)
	assert.True(t, strings.Contains(text, fetchError.Error()))
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
	model.library.scanCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'd'})

	require.True(t, isTeaQuit(command), "returned cmd is not tea.Quit")
	assert.True(t, sentinelCanceled, "cancelCurrentFileScan was not called on quit")
	assert.Nil(t, model.library.scanCancel, "scanCancel = non-nil, want nil after quit")
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

	sentinelCanceled := false
	model.library.scanCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, command, "returned cmd = nil, want non-nil for reload action")

	assert.True(t, sentinelCanceled, "cancelCurrentFileScan was not called on reload")
	assert.Equal(t, uiInitializing, model.state)
	assert.True(t, initLogContains(model, "reloading config..."))
}

func TestHandleKeyPressMsgForwardsToComponentProceed(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	model.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})

	sentinelCanceled := false
	model.library.scanCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, command, "returned cmd should be nil for proceed action")

	assert.True(t, sentinelCanceled, "cancelCurrentFileScan was not called on proceed")
	assert.Equal(t, uiIdle, model.state)
	assert.True(t, initLogContains(model, "proceeding without libraries"))
}

func TestHandleKeyPressMsgMoveLeftIsNoOp(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	assert.Nil(t, command, "returned cmd should be nil for move left (NoAction)")
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
			name:          "ReloadConfigAction returns load cmd and stays initializing",
			action:        action.ReloadConfigAction{},
			wantCmdNonNil: true,
			wantState:     uiInitializing,
			wantStateSet:  true,
			wantLogSubstr: "reloading config...",
			wantCanceled:  true,
			wantCleared:   true,
		},
		{
			name:          "ScanLibraryFullAction returns fetch start cmd and stays initializing",
			action:        action.ScanLibraryFullAction{},
			wantCmdNonNil: true,
			wantState:     uiInitializing,
			wantStateSet:  true,
			wantCanceled:  true,
			wantCleared:   true,
			skipLogCheck:  true,
		},
		{
			name:          "ProceedFromInitAction returns nil and goes idle",
			action:        action.ProceedFromInitAction{},
			wantState:     uiIdle,
			wantStateSet:  true,
			wantLogSubstr: "proceeding without libraries",
			wantCanceled:  true,
			wantCleared:   true,
		},
		{
			name:         "ActionCommand returns the carried cmd",
			action:       action.ActionCommand{Command: tea.Quit},
			wantQuit:     true,
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
			model.library.scanCancel = func() { sentinelCanceled = true }

			command := model.handleComponentAction(test.action)

			if test.wantQuit {
				require.True(t, isTeaQuit(command), "returned cmd is not tea.Quit")
			} else if test.wantCmdNonNil {
				require.NotNil(t, command, "returned cmd = nil, want non-nil")
			} else {
				assert.Nil(t, command, "returned cmd = non-nil, want nil")
			}

			if test.wantCanceled {
				assert.True(t, sentinelCanceled, "cancelCurrentFileScan was not called")
			}

			if test.wantCleared {
				assert.Nil(
					t,
					model.library.scanCancel,
					"scanCancel = non-nil, want nil after action",
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

func TestFetchFilesStartCommandAsync(t *testing.T) {
	t.Parallel()

	libraryDir := t.TempDir()
	plantAudioFiles(t, libraryDir, 5)

	model := newTestUI(t)

	contextForFetch, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startCommand := fetchFilesStartCommand(contextForFetch, 1, []string{libraryDir})
	startMessage := executeCmd(t, startCommand).(fetchFilesStartMessage)

	_, drainerCommand := model.Update(startMessage)

	resultMessage := runCmds(t, model, drainerCommand)
	result, ok := resultMessage.(fetchFilesResultMessage)
	require.True(t, ok, "runCmds returned %T, want fetchFilesResultMessage", resultMessage)

	require.NoError(t, result.err)
	assert.Len(t, result.files, 5, "result should carry the discovered audio files")

	_, _ = model.Update(result)

	assert.Len(t, *model.library.audioFiles, 5, "library.audioFiles should be populated from the result")
}
