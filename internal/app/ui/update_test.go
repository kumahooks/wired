package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/app/ui/components/initializing"
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

// runCmds drives the count drainer to completion. It executes the drainer command, feeds WaitProgress messages back
// into model.Update to get the next drainer, and returns the first initializationCountFilesResultMessage.
func runCmds(t *testing.T, model *UIModel, command tea.Cmd) tea.Msg {
	t.Helper()

	timeout := time.After(3 * time.Second)

	for command != nil {
		result := make(chan tea.Msg, 1)
		go func() { result <- command() }()

		select {
		case message := <-result:
			switch message := message.(type) {
			case initializationCountFilesResultMessage:
				return message
			case initializationCountFilesWaitProgressMessage:
				_, command = model.Update(message)
			default:
				t.Fatalf("unexpected message from drainer: %T", message)
			}
		case <-timeout:
			t.Fatal("timed out waiting for count result")
		}
	}

	return nil
}

func plantAudioFiles(t *testing.T, dir string, count int) {
	t.Helper()

	for index := range count {
		path := filepath.Join(dir, "track"+strconv.Itoa(index)+".mp3")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %q: %v", path, err)
		}
	}
}

// initLogTexts reads the unexported logLines slice of initializationModel via reflect.
func initLogTexts(model *UIModel) []string {
	modelValue := reflect.ValueOf(model.initializationModel).Elem()
	logLinesField := modelValue.FieldByName("logLines")

	texts := make([]string, logLinesField.Len())
	for index := range texts {
		texts[index] = logLinesField.Index(index).FieldByName("text").String()
	}

	return texts
}

func initLogContains(model *UIModel, substring string) bool {
	for _, text := range initLogTexts(model) {
		if strings.Contains(text, substring) {
			return true
		}
	}

	return false
}

func initLastLog(model *UIModel) (string, initializing.LogType) {
	modelValue := reflect.ValueOf(model.initializationModel).Elem()
	logLinesField := modelValue.FieldByName("logLines")
	last := logLinesField.Index(logLinesField.Len() - 1)

	return last.FieldByName("text").String(), initializing.LogType(last.FieldByName("logType").Int())
}

func initProgress(model *UIModel) int {
	return int(reflect.ValueOf(model.initializationModel).Elem().FieldByName("countFilesProgress").Int())
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

	if command != nil {
		t.Errorf("returned cmd = non-nil, want nil on error")
	}

	if !initLogContains(model, loadError.Error()) {
		t.Errorf("log missing error %q", loadError.Error())
	}

	text, logType := initLastLog(model)
	if logType != initializing.LogError {
		t.Errorf("last log type = %v, want LogError (text: %q)", logType, text)
	}
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

	if !initLogContains(model, "no config file found, loading one using defaults") {
		t.Error("missing 'no config file found' log line")
	}

	for _, text := range initLogTexts(model) {
		if strings.Contains(text, "[keymap:New]") || strings.Contains(text, "falling back") {
			t.Errorf("unexpected keymap error log on defaults: %q", text)
		}
	}

	if command == nil {
		t.Error("returned cmd = nil, want non-nil (libraries present)")
	}
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

	if !reflect.DeepEqual(*model.config, customConfig) {
		t.Errorf("model.config = %#v, want %#v", *model.config, customConfig)
	}

	wantTheme := theme.New(customConfig.Theme)
	if !reflect.DeepEqual(model.theme, wantTheme) {
		t.Errorf("model.theme = %#v, want %#v", model.theme, wantTheme)
	}

	wantKeyMap, err := keymap.New(customConfig.Keybinds)
	if err != nil {
		t.Fatalf("keymap.New(custom) error: %v", err)
	}
	if !reflect.DeepEqual(model.keyMap, wantKeyMap) {
		t.Errorf("model.keyMap = %#v, want %#v", model.keyMap, wantKeyMap)
	}

	if reflect.DeepEqual(model.keyMap, initialKeyMap) {
		t.Error("model.keyMap unchanged after loading custom keybinds")
	}

	if !initLogContains(model, "config loaded successfully") {
		t.Error("missing 'config loaded successfully' log")
	}
	if !initLogContains(model, "theme loaded successfully") {
		t.Error("missing 'theme loaded successfully' log")
	}
	if !initLogContains(model, "keybindings loaded successfully") {
		t.Error("missing 'keybindings loaded successfully' log")
	}
	if initLogContains(model, "no config file found") {
		t.Error("'no config file found' should not appear on isConfigDefaults=false")
	}

	if command == nil {
		t.Error("returned cmd = nil, want non-nil (libraries present)")
	}
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

	if !initLogContains(model, "[keymap:New]") {
		t.Error("missing [keymap:New] error log")
	}
	if !initLogContains(model, "falling back to default keybindings") {
		t.Error("missing 'falling back to default keybindings' log")
	}

	if !reflect.DeepEqual(model.keyMap, initialKeyMap) {
		t.Errorf("model.keyMap changed on parse failure = %#v, want %#v", model.keyMap, initialKeyMap)
	}

	if !initLogContains(model, "falling back to default keybindings") {
		t.Error("ApplyKeyMap(default) should have produced the fallback log")
	}

	if command == nil {
		t.Error("returned cmd = nil, want non-nil (libraries present, count still starts)")
	}
}

func TestHandleInitializationLoadConfigResultNoLibrariesNoCountCmd(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	emptyConfig := config.Defaults()
	emptyConfig.LibrariesPaths = []string{}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &emptyConfig,
		isConfigDefaults: false,
		err:              nil,
	})

	if command != nil {
		t.Errorf("returned cmd = non-nil, want nil when no library paths")
	}

	if !initLogContains(model, "no library paths found") {
		t.Error("missing 'no library paths found' log line")
	}

	text, logType := initLastLog(model)
	if logType != initializing.LogError {
		t.Errorf("last log type = %v, want LogError (text: %q)", logType, text)
	}
}

func TestHandleInitializationLoadConfigResultLibrariesEmitsCountStart(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	loadedConfig := config.Defaults()
	loadedConfig.LibrariesPaths = []string{t.TempDir()}

	_, command := model.Update(initializationLoadConfigResultMessage{
		config:           &loadedConfig,
		isConfigDefaults: false,
		err:              nil,
	})

	if command == nil {
		t.Fatal("returned cmd = nil, want a count start command")
	}

	if model.countGeneration != 1 {
		t.Errorf("countGeneration = %d, want 1", model.countGeneration)
	}

	message := executeCmd(t, command)
	startMessage, ok := message.(initializationCountFilesStartMessage)
	if !ok {
		t.Fatalf("cmd produced %T, want initializationCountFilesStartMessage", message)
	}

	if startMessage.generation != model.countGeneration {
		t.Errorf("startMessage.generation = %d, want %d", startMessage.generation, model.countGeneration)
	}

	if startMessage.countCancel != nil {
		startMessage.countCancel()
	}
}

func TestHandleInitializationCountFilesStartMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.countGeneration = 5

	_, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	progressChannel := make(chan int, 1)
	resultChannel := make(chan initializationCountFilesResultMessage, 1)

	message := initializationCountFilesStartMessage{
		progressChannel: progressChannel,
		resultChannel:   resultChannel,
		countCancel:     cancel,
		generation:      7,
	}

	_, command := model.Update(message)

	if model.cancelInitializationCount == nil {
		t.Fatal("cancelInitializationCount = nil, want the cancel func")
	}
	if model.countGeneration != 7 {
		t.Errorf("countGeneration = %d, want 7", model.countGeneration)
	}
	if initProgress(model) != 0 {
		t.Errorf("countFilesProgress = %d, want 0", initProgress(model))
	}
	if !initLogContains(model, "counting total library files") {
		t.Error("missing 'counting total library files' log")
	}
	if command == nil {
		t.Fatal("returned cmd = nil, want the drainer cmd")
	}
}

func TestHandleInitializationCountFilesWaitProgressMessage(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	progressChannel := make(chan int, 1)
	resultChannel := make(chan initializationCountFilesResultMessage, 1)

	_, command := model.Update(initializationCountFilesWaitProgressMessage{
		filesCount:      42,
		progressChannel: progressChannel,
		resultChannel:   resultChannel,
		generation:      3,
	})

	if initProgress(model) != 42 {
		t.Errorf("countFilesProgress = %d, want 42", initProgress(model))
	}
	if command == nil {
		t.Error("returned cmd = nil, want the next drainer cmd")
	}
}

func TestHandleInitializationCountFilesResultMessageStaleGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.countGeneration = 10
	model.initializationModel.SetCountFilesProgress(50)

	sentinelCanceled := false
	model.cancelInitializationCount = func() { sentinelCanceled = true }

	initialLogCount := len(initLogTexts(model))

	_, command := model.Update(initializationCountFilesResultMessage{
		filesCount: 100,
		err:        nil,
		generation: 5,
	})

	if command != nil {
		t.Error("returned cmd = non-nil, want nil on stale generation")
	}
	if sentinelCanceled {
		t.Error("stale result cleared cancelInitializationCount")
	}
	if model.cancelInitializationCount == nil {
		t.Error("cancelInitializationCount = nil, want unchanged sentinel on stale generation")
	}
	if initProgress(model) != 50 {
		t.Errorf("countFilesProgress = %d, want 50 (unchanged on stale)", initProgress(model))
	}
	if len(initLogTexts(model)) != initialLogCount {
		t.Errorf("logCount = %d, want %d (unchanged on stale)", len(initLogTexts(model)), initialLogCount)
	}
}

func TestHandleInitializationCountFilesResultMessageCurrentGeneration(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.countGeneration = 10

	sentinelCanceled := false
	model.cancelInitializationCount = func() { sentinelCanceled = true }

	_, command := model.Update(initializationCountFilesResultMessage{
		filesCount: 137,
		err:        nil,
		generation: 10,
	})

	if command != nil {
		t.Error("returned cmd = non-nil, want nil on current generation result")
	}
	if sentinelCanceled {
		t.Error("current result should set cancelInitializationCount to nil, not call it")
	}
	if model.cancelInitializationCount != nil {
		t.Error("cancelInitializationCount = non-nil, want nil after current generation result")
	}
	if initProgress(model) != -1 {
		t.Errorf("countFilesProgress = %d, want -1", initProgress(model))
	}

	text, logType := initLastLog(model)
	if logType != initializing.LogNormal {
		t.Errorf("last log type = %v, want LogNormal (text: %q)", logType, text)
	}
	if !strings.Contains(text, "counted a total of 137 audio files successfully") {
		t.Errorf("last log text = %q, want it to contain the count", text)
	}
}

func TestHandleInitializationCountFilesResultMessageError(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)
	model.countGeneration = 10
	model.initializationModel.SetCountFilesProgress(50)

	countError := errors.New("[audio:CountFiles] walk failed")
	_, command := model.Update(initializationCountFilesResultMessage{
		filesCount: 0,
		err:        countError,
		generation: 10,
	})

	if command != nil {
		t.Error("returned cmd = non-nil, want nil on error result")
	}
	if initProgress(model) != 50 {
		t.Errorf("countFilesProgress = %d, want 50 (unchanged on error)", initProgress(model))
	}

	text, logType := initLastLog(model)
	if logType != initializing.LogError {
		t.Errorf("last log type = %v, want LogError (text: %q)", logType, text)
	}
	if !strings.Contains(text, countError.Error()) {
		t.Errorf("last log text = %q, want it to contain %q", text, countError.Error())
	}
}

func TestHandleWindowResize(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	if model.windowWidth != 100 {
		t.Errorf("windowWidth = %d, want 100", model.windowWidth)
	}
	if model.windowHeight != 40 {
		t.Errorf("windowHeight = %d, want 40", model.windowHeight)
	}
	if command != nil {
		t.Error("returned cmd = non-nil, want nil for window resize")
	}
}

func TestHandleKeyPressMsgQuit(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	sentinelCanceled := false
	model.cancelInitializationCount = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'd'})

	if !isTeaQuit(command) {
		t.Fatal("returned cmd is not tea.Quit")
	}
	if !sentinelCanceled {
		t.Error("cancelInitializationCount was not called on quit")
	}
	if model.cancelInitializationCount != nil {
		t.Error("cancelInitializationCount = non-nil, want nil after quit")
	}
}

func TestHandleKeyPressMsgQuitDoesNotMatchNonQuitKey(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	if command != nil {
		t.Errorf("returned cmd = non-nil, want nil for unmatched key")
	}
}

func TestHandleKeyPressMsgForwardsToComponentReload(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	sentinelCanceled := false
	model.cancelInitializationCount = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if command == nil {
		t.Fatal("returned cmd = nil, want non-nil for reload action")
	}
	if !sentinelCanceled {
		t.Error("cancelInitializationCount was not called on reload")
	}
	if model.state != uiInitializing {
		t.Errorf("state = %v, want uiInitializing", model.state)
	}
	if !initLogContains(model, "reloading config...") {
		t.Error("missing 'reloading config...' log line")
	}
}

func TestHandleKeyPressMsgForwardsToComponentProceed(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	model.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})

	sentinelCanceled := false
	model.cancelInitializationCount = func() { sentinelCanceled = true }

	_, command := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if command != nil {
		t.Errorf("returned cmd = non-nil, want nil for proceed action")
	}
	if !sentinelCanceled {
		t.Error("cancelInitializationCount was not called on proceed")
	}
	if model.state != uiIdle {
		t.Errorf("state = %v, want uiIdle", model.state)
	}
	if !initLogContains(model, "proceeding without libraries") {
		t.Error("missing 'proceeding without libraries' log line")
	}
}

func TestHandleKeyPressMsgMoveLeftIsNoOp(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	_, command := model.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	if command != nil {
		t.Errorf("returned cmd = non-nil, want nil for move left (NoAction)")
	}
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
			model.cancelInitializationCount = func() { sentinelCanceled = true }

			command := model.handleComponentAction(test.action)

			if test.wantQuit {
				if !isTeaQuit(command) {
					t.Fatalf("returned cmd is not tea.Quit")
				}
			} else if test.wantCmdNonNil {
				if command == nil {
					t.Fatal("returned cmd = nil, want non-nil")
				}
			} else {
				if command != nil {
					t.Errorf("returned cmd = non-nil, want nil")
				}
			}

			if test.wantCanceled && !sentinelCanceled {
				t.Error("cancelInitializationCount was not called")
			}

			if test.wantCleared && model.cancelInitializationCount != nil {
				t.Error("cancelInitializationCount = non-nil, want nil after action")
			}

			if test.wantStateSet && model.state != test.wantState {
				t.Errorf("state = %v, want %v", model.state, test.wantState)
			}

			if !test.skipLogCheck {
				if !initLogContains(model, test.wantLogSubstr) {
					t.Errorf("missing log line containing %q", test.wantLogSubstr)
				}
			}
		})
	}
}

func TestInitializationCountFilesStartCommandAsync(t *testing.T) {
	t.Parallel()

	libraryDir := t.TempDir()
	plantAudioFiles(t, libraryDir, 5)

	model := newTestUI(t)

	contextForCount, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startCommand := initializationCountFilesStartCommand(contextForCount, 1, []string{libraryDir})
	startMessage := executeCmd(t, startCommand).(initializationCountFilesStartMessage)

	_, drainerCommand := model.Update(startMessage)

	resultMessage := runCmds(t, model, drainerCommand)
	result, ok := resultMessage.(initializationCountFilesResultMessage)
	if !ok {
		t.Fatalf("runCmds returned %T, want initializationCountFilesResultMessage", resultMessage)
	}

	if result.err != nil {
		t.Fatalf("count error: %v", result.err)
	}
	if result.filesCount != 5 {
		t.Errorf("filesCount = %d, want 5", result.filesCount)
	}

	_, _ = model.Update(result)

	if initProgress(model) != -1 {
		t.Errorf("countFilesProgress = %d, want -1 after result", initProgress(model))
	}

	if !initLogContains(model, "counted a total of 5 audio files successfully") {
		t.Error("missing success log line")
	}
}
