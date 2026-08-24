package ui

import (
	"context"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"

	"wired/internal/core/config"
	"wired/internal/core/keymap"
	"wired/internal/core/theme"
)

// defaultKeyMap builds the keymap from config.Defaults, fataling on a parse error.
func defaultKeyMap(t *testing.T) keymap.KeyMap {
	t.Helper()

	keyMap, err := keymap.New(config.Defaults().Keybinds)
	if err != nil {
		t.Fatalf("keymap.New(defaults) error: %v", err)
	}

	return keyMap
}

// newTestUI builds a *UIModel with the default keymap, default theme, and a *config.Config populated from Defaults, then
// applies a fixed WindowSizeMsg so dimensions are deterministic.
func newTestUI(t *testing.T) *UIModel {
	t.Helper()

	configValue := config.Defaults()
	model, err := New(context.Background(), defaultKeyMap(t), &configValue)
	if err != nil {
		t.Fatalf("ui.New error: %v", err)
	}

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	gotModel, ok := updatedModel.(*UIModel)
	if !ok {
		t.Fatalf("Update(WindowSizeMsg) returned %T, want *UIModel", updatedModel)
	}

	return gotModel
}

func TestNewSeedsDefaults(t *testing.T) {
	t.Parallel()

	configValue := config.Defaults()
	keyMap := defaultKeyMap(t)

	model, err := New(context.Background(), keyMap, &configValue)
	if err != nil {
		t.Fatalf("ui.New error: %v", err)
	}

	if model == nil {
		t.Fatal("ui.New returned nil model")
	}

	if model.state != uiInitializing {
		t.Errorf("state = %v, want uiInitializing", model.state)
	}

	if !reflect.DeepEqual(model.theme, theme.Default()) {
		t.Errorf("theme = %#v, want theme.Default()", model.theme)
	}

	if model.initializationModel == nil {
		t.Errorf("initializationModel is nil, want a non-nil *initializing.Model")
	}

	if model.config != &configValue {
		t.Errorf("config pointer = %p, want the one passed in (%p)", model.config, &configValue)
	}

	if !reflect.DeepEqual(model.keyMap, keyMap) {
		t.Errorf("keyMap = %#v, want the passed keyMap", model.keyMap)
	}
}

func TestInitReturnsNonNilCmd(t *testing.T) {
	t.Parallel()

	model, err := New(context.Background(), defaultKeyMap(t), func() *config.Config {
		configValue := config.Defaults()
		return &configValue
	}())
	if err != nil {
		t.Fatalf("ui.New error: %v", err)
	}

	// Init returns initializationLoadConfigCommand, which calls config.Load against the real user config dir. We do
	// not execute it here because that would touch the real filesystem.
	command := model.Init()
	if command == nil {
		t.Fatal("Init() returned nil cmd, want a non-nil tea.Cmd")
	}
}

func TestNewTestUIAppliesWindowSize(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	if model.windowWidth != 80 {
		t.Errorf("windowWidth = %d, want 80", model.windowWidth)
	}
	if model.windowHeight != 24 {
		t.Errorf("windowHeight = %d, want 24", model.windowHeight)
	}

	if model.state != uiInitializing {
		t.Errorf("state = %v, want uiInitializing", model.state)
	}

	if model.initializationModel == nil {
		t.Errorf("initializationModel is nil after newTestUI")
	}
}
