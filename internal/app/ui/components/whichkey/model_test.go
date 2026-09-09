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

// testBindings builds a two-entry binding list.
func testBindings(t *testing.T) []action.Binding {
	t.Helper()

	keyMap := testutil.DefaultKeyMap(t)

	return []action.Binding{
		{Keys: keyMap.Actions.Playlist, Action: action.OpenPlaylistAction{}},
		{Keys: keyMap.Actions.LibraryStats, Action: action.OpenLibraryStatsAction{}},
	}
}

// newTestModel builds a Model with the default style, close key, and the test binding list applied.
func newTestModel(t *testing.T) *Model {
	t.Helper()

	model := New()
	model.ApplyCloseKeybinding(testutil.DefaultKeyMap(t).GoBack)
	model.SetBindings(testBindings(t))

	return model
}

func TestNew(t *testing.T) {
	t.Parallel()

	model := New()

	assert.False(t, model.isVisible, "a new model must start hidden")
	assert.Empty(t, model.bindings, "a new model must start with no bindings")
	assert.Equal(t, newStyle(testutil.DefaultTheme()), model.style)
}

func TestSetBindingsReplacesBindings(t *testing.T) {
	t.Parallel()

	model := New()
	require.Empty(t, model.bindings, "a new model must start with no bindings")

	updated := []action.Binding{{
		Keys:   key.NewBinding(key.WithKeys("Q"), key.WithHelp("Q", "custom")),
		Action: action.QuitAction{},
	}}
	model.SetBindings(updated)

	assert.Equal(t, updated, model.bindings)
}

func TestApplyKeyMapStoresCloseKey(t *testing.T) {
	t.Parallel()

	model := New()
	model.ApplyCloseKeybinding(testutil.DefaultKeyMap(t).GoBack)

	goBackKey := testutil.DefaultKeyMap(t).GoBack.Help().Key
	assert.Equal(t, goBackKey, model.closeKey)
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New()

	customTheme := testutil.CustomTheme()

	model.ApplyTheme(customTheme)

	assert.Equal(t, newStyle(customTheme), model.style)
}

func TestFlipIsVisible(t *testing.T) {
	t.Parallel()

	model := New()

	model.flipIsVisible()
	require.True(t, model.IsVisible(), "flip from hidden must show the card")

	model.flipIsVisible()
	assert.False(t, model.IsVisible(), "flip from shown must hide the card")
}

func TestHandleMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		message    tea.Msg
		wantAction action.Action
	}{
		{
			name:       "playlist key flips visibility and returns open playlist",
			message:    tea.KeyPressMsg{Code: 'p'},
			wantAction: action.OpenPlaylistAction{},
		},
		{
			name:       "library stats key flips visibility and returns open library stats",
			message:    tea.KeyPressMsg{Code: 's'},
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

			model := newTestModel(t)
			require.False(t, model.IsVisible(), "a new model must start hidden")

			returnedAction := model.HandleMessage(test.message)

			assert.Equal(t, test.wantAction, returnedAction)
		})
	}
}

func TestHandleMessageSecondPressClosesCard(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)
	model.HandleMessage(tea.KeyPressMsg{Code: ' '})
	require.True(t, model.IsVisible(), "first open actions press must show the card")

	model.HandleMessage(tea.KeyPressMsg{Code: ' '})
	assert.False(t, model.IsVisible(), "second open actions press must hide the card")
}

func TestHandleMessageNonKeyPressKeepsVisibility(t *testing.T) {
	t.Parallel()

	model := newTestModel(t)
	model.HandleMessage(tea.KeyPressMsg{Code: 'P'})
	require.True(t, model.IsVisible(), "card must be visible before the non key message")

	model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.True(t, model.IsVisible(), "a non key press message must not toggle the card")
}

func TestHandleMessageWithoutBindingsNeverDispatches(t *testing.T) {
	t.Parallel()

	model := New()
	model.flipIsVisible()
	require.True(t, model.IsVisible())

	returnedAction := model.HandleMessage(tea.KeyPressMsg{Code: 'p'})

	assert.Equal(t, action.NoAction{}, returnedAction)
}
