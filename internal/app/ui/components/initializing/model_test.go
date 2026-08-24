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

	require.Len(t, model.buttons, 2)
	assert.Equal(t, actionReload, model.buttons[0].action)
	assert.Equal(t, actionProceed, model.buttons[1].action)
	assert.Zero(t, model.cursorPosition)
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

func TestSetCountFilesProgress(t *testing.T) {
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

			model.SetCountFilesProgress(test.value)

			assert.Equal(t, test.value, model.countFilesProgress)
		})
	}
}

func TestHandleMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		message       tea.Msg
		setupButtons  func() []button
		setupCursor   int
		wantAction    action.Action
		wantCursor    int
		wantCursorSet bool
	}{
		{
			name:         "non-keypress returns NoAction",
			message:      tea.WindowSizeMsg{},
			setupButtons: func() []button { return New(defaultKeyMap(t)).buttons },
			wantAction:   action.NoAction{},
		},
		{
			name:          "MoveLeft wraps from zero to last",
			message:       tea.KeyPressMsg{Code: 'h', Text: "h"},
			setupButtons:  func() []button { return New(defaultKeyMap(t)).buttons },
			setupCursor:   0,
			wantAction:    action.NoAction{},
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:          "MoveRight advances from zero to one",
			message:       tea.KeyPressMsg{Code: 'l', Text: "l"},
			setupButtons:  func() []button { return New(defaultKeyMap(t)).buttons },
			setupCursor:   0,
			wantAction:    action.NoAction{},
			wantCursor:    1,
			wantCursorSet: true,
		},
		{
			name:          "Select on reload returns ReloadConfigAction",
			message:       tea.KeyPressMsg{Code: tea.KeyEnter},
			setupButtons:  func() []button { return New(defaultKeyMap(t)).buttons },
			setupCursor:   0,
			wantAction:    action.ReloadConfigAction{},
			wantCursorSet: false,
		},
		{
			name:          "Select on proceed returns ProceedFromInitAction",
			message:       tea.KeyPressMsg{Code: tea.KeyEnter},
			setupButtons:  func() []button { return New(defaultKeyMap(t)).buttons },
			setupCursor:   1,
			wantAction:    action.ProceedFromInitAction{},
			wantCursorSet: false,
		},
		{
			name:         "Select with empty buttons returns NoAction",
			message:      tea.KeyPressMsg{Code: tea.KeyEnter},
			setupButtons: func() []button { return nil },
			wantAction:   action.NoAction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t))
			model.buttons = test.setupButtons()
			model.cursorPosition = test.setupCursor

			gotAction := model.HandleMessage(test.message)

			assert.Equal(t, reflect.TypeOf(test.wantAction), reflect.TypeOf(gotAction))

			if test.wantCursorSet {
				assert.Equal(t, test.wantCursor, model.cursorPosition)
			}
		})
	}
}
