package keymap

import (
	"reflect"
	"strings"
	"testing"

	"wired/internal/core/config"
)

func TestNewHappyPath(t *testing.T) {
	t.Parallel()

	bindings := config.Defaults().Keybinds
	keyMap, err := New(bindings)
	if err != nil {
		t.Fatalf("New(defaults) error: %v", err)
	}

	tests := []struct {
		name     string
		binding  func() string
		keys     func() []string
		wantKey  string
		wantKeys []string
	}{
		{
			name:     "move left",
			binding:  func() string { return keyMap.MoveLeft.Help().Key },
			keys:     func() []string { return keyMap.MoveLeft.Keys() },
			wantKey:  bindings.MoveLeft[0],
			wantKeys: bindings.MoveLeft,
		},
		{
			name:     "move right",
			binding:  func() string { return keyMap.MoveRight.Help().Key },
			keys:     func() []string { return keyMap.MoveRight.Keys() },
			wantKey:  bindings.MoveRight[0],
			wantKeys: bindings.MoveRight,
		},
		{
			name:     "select",
			binding:  func() string { return keyMap.Select.Help().Key },
			keys:     func() []string { return keyMap.Select.Keys() },
			wantKey:  bindings.Select[0],
			wantKeys: bindings.Select,
		},
		{
			name:     "quit",
			binding:  func() string { return keyMap.Quit.Help().Key },
			keys:     func() []string { return keyMap.Quit.Keys() },
			wantKey:  bindings.Quit[0],
			wantKeys: bindings.Quit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.binding(); got != test.wantKey {
				t.Errorf("Help().Key = %q, want %q", got, test.wantKey)
			}
			if got := test.keys(); !reflect.DeepEqual(got, test.wantKeys) {
				t.Errorf("Keys() = %v, want %v", got, test.wantKeys)
			}
		})
	}
}

func TestNewEmptyBindingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldName string
		wantError string
	}{
		{
			name:      "empty move_left fails",
			fieldName: "MoveLeft",
			wantError: "[keymap:New] move_left must have at least one binding",
		},
		{
			name:      "empty move_right fails",
			fieldName: "MoveRight",
			wantError: "[keymap:New] move_right must have at least one binding",
		},
		{
			name:      "empty select fails",
			fieldName: "Select",
			wantError: "[keymap:New] select must have at least one binding",
		},
		{
			name:      "empty quit fails",
			fieldName: "Quit",
			wantError: "[keymap:New] quit must have at least one binding",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			bindings := config.Defaults().Keybinds

			keybindsValue := reflect.ValueOf(&bindings).Elem()
			keybindsValue.FieldByName(test.fieldName).Set(reflect.ValueOf([]string{}))

			_, err := New(bindings)
			switch {
			case err == nil:
				t.Fatalf("New() want error containing %q, got nil", test.wantError)
			case !strings.Contains(err.Error(), test.wantError):
				t.Fatalf("New() error = %q, want substring %q", err.Error(), test.wantError)
			}
		})
	}
}
