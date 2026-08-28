package whichkey

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
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

	model.flipIsVisible()
	require.True(t, model.IsVisible(), "flip from hidden must show the card")

	model.flipIsVisible()
	assert.False(t, model.IsVisible(), "flip from shown must hide the card")
}

func TestHandleMessage(t *testing.T) {
	t.Parallel()

	defaultKeyMap := testutil.DefaultKeyMap(t)

	tests := []struct {
		name       string
		message    tea.Msg
		wantAction action.Action
	}{
		{
			name:       "open actions key flips visibility and returns no action",
			message:    tea.KeyPressMsg{Code: ' '},
			wantAction: action.NoAction{},
		},
		{
			name:       "go back key flips visibility and returns no action",
			message:    tea.KeyPressMsg{Code: tea.KeyEscape},
			wantAction: action.NoAction{},
		},
		{
			name:       "playlist key flips visibility and returns open playlist",
			message:    tea.KeyPressMsg{Code: 'P'},
			wantAction: action.OpenPlaylistAction{},
		},
		{
			name:       "library stats key flips visibility and returns open library stats",
			message:    tea.KeyPressMsg{Code: 'L'},
			wantAction: action.OpenLibraryStatsAction{},
		},
		{
			name:       "unmapped key flips visibility and returns no action",
			message:    tea.KeyPressMsg{Code: 'x', Text: "x"},
			wantAction: action.NoAction{},
		},
		{
			name:       "non key press message is ignored and keeps the card hidden",
			message:    tea.WindowSizeMsg{Width: 80, Height: 24},
			wantAction: action.NoAction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap)
			require.False(t, model.IsVisible(), "a new model must start hidden")

			returnedAction := model.HandleMessage(test.message)

			assert.Equal(t, test.wantAction, returnedAction)
		})
	}
}

func TestHandleMessageSecondPressClosesCard(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	model.HandleMessage(tea.KeyPressMsg{Code: ' '})
	require.True(t, model.IsVisible(), "first open actions press must show the card")

	model.HandleMessage(tea.KeyPressMsg{Code: ' '})
	assert.False(t, model.IsVisible(), "second open actions press must hide the card")
}

func TestHandleMessageNonKeyPressKeepsVisibility(t *testing.T) {
	t.Parallel()

	model := New(testutil.DefaultKeyMap(t))

	model.HandleMessage(tea.KeyPressMsg{Code: ' '})
	require.True(t, model.IsVisible(), "card must be visible before the non key message")

	model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.True(t, model.IsVisible(), "a non key press message must not toggle the card")
}
