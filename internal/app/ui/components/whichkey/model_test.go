package whichkey

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/testutil"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaultKeyMap := testutil.DefaultKeyMap(t)

	model := New(defaultKeyMap)

	assert.False(t, model.isVisible, "a new model must start hidden")
	assert.Equal(t, defaultKeyMap, model.keyMap)
	assert.Equal(t, newStyle(testutil.DefaultTheme()), model.style)
}

func TestApplyKeyMapReplacesKeyMap(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	customKeyMap := testutil.DefaultKeyMap(t)
	customKeyMap.Actions.LibraryStats = key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "custom stats"),
	)

	model.ApplyKeyMap(customKeyMap)

	assert.Equal(t, customKeyMap, model.keyMap)
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	customTheme := testutil.CustomTheme()

	model.ApplyTheme(customTheme)

	assert.Equal(t, newStyle(customTheme), model.style)
}

func TestFlipIsVisible(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	model.FlipIsVisible()
	require.True(t, model.IsVisible(), "flip from hidden must show the card")

	model.FlipIsVisible()
	assert.False(t, model.IsVisible(), "flip from shown must hide the card")
}
