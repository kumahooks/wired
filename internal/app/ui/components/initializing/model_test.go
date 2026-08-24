package initializing

import (
	"reflect"
	"strconv"
	"testing"

	tea "charm.land/bubbletea/v2"

	"wired/internal/app/ui/action"
	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

func defaultKeyMap(t *testing.T) keymap.KeyMap {
	t.Helper()

	keyMap, err := keymap.New(config.Defaults().Keybinds)
	if err != nil {
		t.Fatalf("keymap.New(defaults) error: %v", err)
	}

	return keyMap
}

func TestNewSeedsButtonsAndDefaults(t *testing.T) {
	t.Parallel()

	keyMap := defaultKeyMap(t)
	model := New(keyMap)

	if len(model.buttons) != 2 {
		t.Fatalf("len(buttons) = %d, want 2", len(model.buttons))
	}
	if model.buttons[0].action != actionReload {
		t.Errorf("buttons[0].action = %v, want actionReload", model.buttons[0].action)
	}
	if model.buttons[1].action != actionProceed {
		t.Errorf("buttons[1].action = %v, want actionProceed", model.buttons[1].action)
	}

	if model.cursorPosition != 0 {
		t.Errorf("cursorPosition = %d, want 0", model.cursorPosition)
	}

	if !reflect.DeepEqual(model.keyMap, keyMap) {
		t.Errorf("keyMap = %#v, want the passed keyMap", model.keyMap)
	}

	wantStyle := newStyle(theme.Default())
	if !reflect.DeepEqual(model.style, wantStyle) {
		t.Errorf("style = %#v, want newStyle(theme.Default())", model.style)
	}
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t))

	customTheme := theme.New(config.ThemeConfig{
		Surface:           "#0a0a0a",
		SurfaceAlt:        "#0b0b0b",
		BorderPanel:       "#1a1a1a",
		BorderHairline:    "#0c0c0c",
		TextPrimary:       "#aaaaaa",
		TextStrong:        "#ffffff",
		TextMuted:         "#888888",
		TextDim:           "#999999",
		TextFaint:         "#777777",
		TextPlaceholder:   "#666666",
		AccentInteractive: "#b8748a",
		AccentDeep:        "#ad7084",
		AccentBright:      "#d94ea0",
		AccentConfirm:     "#b25c83",
		AccentLink:        "#a65e6e",
		AccentPrompt:      "#b90074",
		AccentDanger:      "#ff6b6b",
		AccentError:       "#112233",
		Track:             "#4a4a4a",
	})

	model.ApplyTheme(customTheme)

	wantStyle := newStyle(customTheme)
	if !reflect.DeepEqual(model.style, wantStyle) {
		t.Errorf("style after ApplyTheme = %#v, want newStyle(customTheme)", model.style)
	}
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
	if err != nil {
		t.Fatalf("keymap.New(alternate) error: %v", err)
	}

	model.ApplyKeyMap(alternateKeyMap)

	if !reflect.DeepEqual(model.keyMap, alternateKeyMap) {
		t.Errorf("keyMap after ApplyKeyMap = %#v, want alternateKeyMap", model.keyMap)
	}
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

			if model.logCount != 1 {
				t.Fatalf("logCount = %d, want 1", model.logCount)
			}

			if len(model.logLines) != 1 {
				t.Fatalf("len(logLines) = %d, want 1", len(model.logLines))
			}

			wantText := "[1] " + test.line
			if got := model.logLines[0].text; got != wantText {
				t.Errorf("logLines[0].text = %q, want %q", got, wantText)
			}

			if model.logLines[0].logType != test.logType {
				t.Errorf("logLines[0].logType = %v, want %v", model.logLines[0].logType, test.logType)
			}
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

	if model.logCount != len(lines) {
		t.Fatalf("logCount = %d, want %d", model.logCount, len(lines))
	}

	if len(model.logLines) != len(lines) {
		t.Fatalf("len(logLines) = %d, want %d", len(model.logLines), len(lines))
	}

	for index, line := range lines {
		wantText := "[" + strconv.Itoa(index+1) + "] " + line.text
		if got := model.logLines[index].text; got != wantText {
			t.Errorf("logLines[%d].text = %q, want %q", index, got, wantText)
		}

		if model.logLines[index].logType != line.logType {
			t.Errorf("logLines[%d].logType = %v, want %v", index, model.logLines[index].logType, line.logType)
		}
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

			if model.countFilesProgress != test.value {
				t.Errorf("countFilesProgress = %d, want %d", model.countFilesProgress, test.value)
			}
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

			if reflect.TypeOf(gotAction) != reflect.TypeOf(test.wantAction) {
				t.Fatalf("HandleMessage action = %T, want %T", gotAction, test.wantAction)
			}

			if test.wantCursorSet && model.cursorPosition != test.wantCursor {
				t.Errorf("cursorPosition = %d, want %d", model.cursorPosition, test.wantCursor)
			}
		})
	}
}
