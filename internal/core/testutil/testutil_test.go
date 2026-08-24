package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/config"
	"wired/internal/core/theme"
)

func TestStripANSI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain string", input: "hello world", want: "hello world"},
		{name: "csi color", input: "\x1b[31mred\x1b[0m", want: "red"},
		{name: "csi with semicolons", input: "\x1b[1;33;40myellow\x1b[0m", want: "yellow"},
		{name: "empty string", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, StripANSI(test.input))
		})
	}
}

func TestDefaultKeyMap(t *testing.T) {
	t.Parallel()

	keyMap := DefaultKeyMap(t)

	require.NotEmpty(t, keyMap.MoveLeft.Keys())
	require.NotEmpty(t, keyMap.MoveRight.Keys())
	require.NotEmpty(t, keyMap.Select.Keys())
	require.NotEmpty(t, keyMap.Quit.Keys())

	// Default move_left binding is "h".
	assert.Contains(t, keyMap.MoveLeft.Keys(), "h")
	// Default select binding is "enter".
	assert.Contains(t, keyMap.Select.Keys(), "enter")
	// Default quit binding is "ctrl+d".
	assert.Contains(t, keyMap.Quit.Keys(), "ctrl+d")
}

func TestDefaultKeyMapReturnsSameDefaults(t *testing.T) {
	t.Parallel()

	keyMap := DefaultKeyMap(t)

	// Default move_left binding's primary key is "h".
	assert.Equal(t, "h", keyMap.MoveLeft.Help().Key)
	// Default select binding's primary key is "enter".
	assert.Equal(t, "enter", keyMap.Select.Help().Key)
	// Default quit binding's primary key is "ctrl+d".
	assert.Equal(t, "ctrl+d", keyMap.Quit.Help().Key)
}

func TestDefaultThemeMatchesConfigDefaults(t *testing.T) {
	t.Parallel()

	got := DefaultTheme()
	want := theme.New(config.Defaults().Theme)
	assert.Equal(t, want, got)
}

func TestCustomThemeDiffersFromDefault(t *testing.T) {
	t.Parallel()

	custom := CustomTheme()
	defaults := DefaultTheme()
	assert.NotEqual(t, defaults, custom)
}
