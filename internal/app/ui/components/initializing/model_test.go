package initializing

import (
	"reflect"
	"strconv"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/testutil"
)

func defaultKeyMap(t *testing.T) keymap.KeyMap {
	t.Helper()

	return testutil.DefaultKeyMap(t)
}

func TestNewSeedsButtonsAndDefaults(t *testing.T) {
	t.Parallel()

	keyMap := defaultKeyMap(t)
	model := New(keyMap)

	require.Len(t, model.buttons, 3)
	assert.Equal(t, actionScan, model.buttons[0].action)
	assert.Equal(t, actionReload, model.buttons[1].action)
	assert.Equal(t, actionProceed, model.buttons[2].action)
	assert.Equal(t, modeLoading, model.mode)
	assert.Equal(t, 2, model.cursorPosition)
	assert.Equal(t, keyMap, model.keyMap)

	wantStyle := newStyle(testutil.DefaultTheme())
	assert.Equal(t, wantStyle, model.style)
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))

	customTheme := testutil.CustomTheme()

	model.ApplyTheme(customTheme)

	wantStyle := newStyle(customTheme)
	assert.Equal(t, wantStyle, model.style)
}

func TestApplyKeyMapStoresKeyMap(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))

	alternateBindings := config.KeybindMapping{
		MoveLeft:  []string{"j"},
		MoveRight: []string{"k"},
		Select:    []string{"space"},
		Quit:      []string{"q"},
	}
	alternateKeyMap, err := keymap.New(alternateBindings)
	require.NoError(t, err)

	model.ApplyKeyMap(alternateKeyMap)

	assert.Equal(t, alternateKeyMap, model.keyMap)
}

func TestAppendLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		line    string
		logType LogType
	}{
		{name: "normal line", line: "config loaded successfully", logType: LogNormal},
		{name: "error line", line: "theme.surface must be a hex color", logType: LogError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t))
			model.AppendLog(test.line, test.logType)

			assert.Equal(t, 1, model.logCount)
			require.Len(t, model.logLines, 1)

			wantText := "[1] " + test.line
			assert.Equal(t, wantText, model.logLines[0].text)
			assert.Equal(t, test.logType, model.logLines[0].logType)
		})
	}
}

func TestAppendLogAccumulatesInOrder(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))

	lines := []struct {
		text    string
		logType LogType
	}{
		{"first", LogNormal},
		{"second", LogError},
		{"third", LogNormal},
	}

	for _, line := range lines {
		model.AppendLog(line.text, line.logType)
	}

	assert.Equal(t, len(lines), model.logCount)
	require.Len(t, model.logLines, len(lines))

	for index, line := range lines {
		wantText := "[" + strconv.Itoa(index+1) + "] " + line.text
		assert.Equal(t, wantText, model.logLines[index].text)
		assert.Equal(t, line.logType, model.logLines[index].logType)
	}
}

func TestSetFetchFilesProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int
	}{
		{name: "positive", value: 42},
		{name: "zero", value: 0},
		{name: "negative hides progress", value: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t))
			model.SetFetchFilesProgress(test.value)

			assert.Equal(t, test.value, model.fetchFilesProgress)
		})
	}
}

func TestButtonsForMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      initMode
		wantOrder []buttonAction
	}{
		{name: "loading shows only proceed", mode: modeLoading, wantOrder: []buttonAction{actionProceed}},
		{
			name:      "config error shows reload and proceed",
			mode:      modeConfigError,
			wantOrder: []buttonAction{actionReload, actionProceed},
		},
		{
			name:      "empty library shows scan and proceed",
			mode:      modeEmptyLibrary,
			wantOrder: []buttonAction{actionScan, actionProceed},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.wantOrder, buttonsForMode(test.mode))
		})
	}
}

func TestSetConfigError(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))
	model.SetConfigError()

	assert.Equal(t, modeConfigError, model.Mode())
	assert.Equal(t, 1, model.cursorPosition, "cursor should land on the first visible button (reload)")
}

func TestSetEmptyLibrary(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))

	model.SetEmptyLibrary()

	assert.Equal(t, modeEmptyLibrary, model.Mode())
	assert.Equal(t, 0, model.cursorPosition, "cursor should land on the first visible button (scan)")
}

func TestHandleMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		message       tea.Msg
		setupMode     initMode
		setupCursor   int
		clearButtons  bool
		wantAction    action.Action
		wantCursor    int
		wantCursorSet bool
	}{
		{
			name:        "non-keypress returns NoAction",
			message:     tea.WindowSizeMsg{},
			setupMode:   modeConfigError,
			setupCursor: 1,
			wantAction:  action.NoAction{},
		},
		{
			name:          "MoveLeft wraps from reload to proceed in config error",
			message:       tea.KeyPressMsg{Code: 'h', Text: "h"},
			setupMode:     modeConfigError,
			setupCursor:   1,
			wantAction:    action.NoAction{},
			wantCursor:    2,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight advances from reload to proceed in config error",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupMode:     modeConfigError,
			setupCursor:   1,
			wantAction:    action.NoAction{},
			wantCursor:    2,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight wraps from proceed to reload in config error",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupMode:     modeConfigError,
			setupCursor:   2,
			wantAction:    action.NoAction{},
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight in loading is a no-op with a single visible button",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupMode:     modeLoading,
			setupCursor:   2,
			wantAction:    action.NoAction{},
			wantCursor:    2,
			wantCursorSet: true,
		},
		{
			name:        "Select on reload returns ReloadConfigAction",
			message:     tea.KeyPressMsg{Code: tea.KeyEnter},
			setupMode:   modeConfigError,
			setupCursor: 1,
			wantAction:  action.ReloadConfigAction{},
		},
		{
			name:        "Select on proceed returns ProceedFromInitAction",
			message:     tea.KeyPressMsg{Code: tea.KeyEnter},
			setupMode:   modeConfigError,
			setupCursor: 2,
			wantAction:  action.ProceedFromInitAction{},
		},
		{
			name:        "Select on scan returns ScanLibraryFullAction",
			message:     tea.KeyPressMsg{Code: tea.KeyEnter},
			setupMode:   modeEmptyLibrary,
			setupCursor: 0,
			wantAction:  action.ScanLibraryFullAction{},
		},
		{
			name:         "Select with empty buttons returns NoAction",
			message:      tea.KeyPressMsg{Code: tea.KeyEnter},
			setupMode:    modeLoading,
			setupCursor:  0,
			clearButtons: true,
			wantAction:   action.NoAction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t))
			if test.clearButtons {
				model.buttons = nil
			}

			model.mode = test.setupMode
			model.cursorPosition = test.setupCursor

			gotAction := model.HandleMessage(test.message)

			assert.Equal(t, reflect.TypeOf(test.wantAction), reflect.TypeOf(gotAction))

			if test.wantCursorSet {
				assert.Equal(t, test.wantCursor, model.cursorPosition)
			}
		})
	}
}
