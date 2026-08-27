package keymap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/config"
)

func TestNewHappyPath(t *testing.T) {
	t.Parallel()

	bindings := config.Defaults().Keybinds
	keyMap, err := New(bindings)
	require.NoError(t, err)

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
		{
			name:     "go back",
			binding:  func() string { return keyMap.GoBack.Help().Key },
			keys:     func() []string { return keyMap.GoBack.Keys() },
			wantKey:  bindings.GoBack[0],
			wantKeys: bindings.GoBack,
		},
		{
			name:     "open actions",
			binding:  func() string { return keyMap.OpenActions.Help().Key },
			keys:     func() []string { return keyMap.OpenActions.Keys() },
			wantKey:  bindings.OpenActions[0],
			wantKeys: bindings.OpenActions,
		},
		{
			name:     "actions library stats",
			binding:  func() string { return keyMap.Actions.LibraryStats.Help().Key },
			keys:     func() []string { return keyMap.Actions.LibraryStats.Keys() },
			wantKey:  bindings.Actions.LibraryStats[0],
			wantKeys: bindings.Actions.LibraryStats,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.wantKey, test.binding())
			assert.Equal(t, test.wantKeys, test.keys())
		})
	}
}

func TestNewEmptyBindingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		emptyFunc func(config.KeybindMapping) config.KeybindMapping
		wantError string
	}{
		{
			name: "empty move_left fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.MoveLeft = []string{}
				return bindings
			},
			wantError: "[keymap:New] move_left must have at least one binding",
		},
		{
			name: "empty move_right fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.MoveRight = []string{}
				return bindings
			},
			wantError: "[keymap:New] move_right must have at least one binding",
		},
		{
			name: "empty select fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.Select = []string{}
				return bindings
			},
			wantError: "[keymap:New] select must have at least one binding",
		},
		{
			name: "empty quit fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.Quit = []string{}
				return bindings
			},
			wantError: "[keymap:New] quit must have at least one binding",
		},
		{
			name: "empty go_back fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.GoBack = []string{}
				return bindings
			},
			wantError: "[keymap:New] go_back must have at least one binding",
		},
		{
			name: "empty open_actions fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.OpenActions = []string{}
				return bindings
			},
			wantError: "[keymap:New] open_actions must have at least one binding",
		},
		{
			name: "empty actions.library_stats fails",
			emptyFunc: func(bindings config.KeybindMapping) config.KeybindMapping {
				bindings.Actions.LibraryStats = []string{}
				return bindings
			},
			wantError: "[keymap:New] library_stats must have at least one binding",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			bindings := test.emptyFunc(config.Defaults().Keybinds)

			_, err := New(bindings)
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), test.wantError))
		})
	}
}
