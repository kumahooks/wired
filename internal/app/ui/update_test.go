package ui

import (
	"context"
	"errors"
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
	assert.True(t, initLogContains(model, "keybindings failed to load, fallbacking to previous bindings"))

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
		library: audio.NewLibrary(),
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
	assert.True(t, initLogContains(model, "no songs found, you might want to discover them later~"))
	assert.False(
		t,
		model.initializationModel.IsConfigError(),
		"empty cache with library paths should not be a config error",
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

	sentinelCanceled := false
	model.libraryDiscoveryCancel = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, command, "returned cmd = nil, want non-nil for reload action")

	assert.True(t, sentinelCanceled, "cancelCurrentLibraryDiscovery was not called on reload")
	assert.Equal(t, uiInitializing, model.state)
	assert.True(t, initLogContains(model, "reloading config..."))
}

func TestHandleKeyPressMsgForwardsToComponentProceed(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

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
			sequence:     []tea.KeyPressMsg{{Code: ' '}, {Code: 'P'}},
			wantHidden:   true,
			wantState:    uiPlaylist,
			wantStateSet: true,
		},
		{
			name:         "open actions opens the card and maps library stats key to library stats state",
			sequence:     []tea.KeyPressMsg{{Code: ' '}, {Code: 'L'}},
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

	model := newTestUI(t)
	require.Equal(t, uiInitializing, model.state)

	_, command := model.Update(tea.KeyPressMsg{Code: ' ', Text: " "})

	assert.Nil(t, command, "returned cmd should be nil while initializing")
	assert.False(t, model.whichkeyModel.IsVisible(), "whichkey must not capture keys during initialization")
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
			name:          "DiscoverLibraryFullAction returns discovery start cmd and stays initializing",
			action:        action.DiscoverLibraryFullAction{},
			wantCmdNonNil: true,
			wantState:     uiInitializing,
			wantStateSet:  true,
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
	t.Parallel()

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
