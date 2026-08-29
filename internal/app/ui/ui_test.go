package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/core/audio"
	"wired/internal/core/config"
	"wired/internal/core/testutil"
)

// newTestUI builds a *UIModel with the default keymap, default theme, and a *config.Config populated from Defaults, then
// applies a fixed WindowSizeMsg so dimensions are deterministic.
func newTestUI(t *testing.T) *UIModel {
	t.Helper()

	configValue := config.Defaults()
	audioFiles := &[]audio.File{}

	model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue, audioFiles)
	require.NoError(t, err)

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	gotModel, ok := updatedModel.(*UIModel)
	require.True(t, ok, "Update(WindowSizeMsg) returned %T, want *UIModel", updatedModel)

	return gotModel
}

func TestNewSeedsDefaults(t *testing.T) {
	t.Parallel()

	configValue := config.Defaults()
	keyMap := testutil.DefaultKeyMap(t)
	audioFiles := &[]audio.File{}

	model, err := New(context.Background(), keyMap, &configValue, audioFiles)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, uiInitializing, model.state)
	assert.Equal(t, testutil.DefaultTheme(), model.theme)
	require.NotNil(t, model.initializationModel, "initializationModel is nil, want a non-nil *initializing.Model")
	assert.Equal(t, &configValue, model.config, "config pointer mismatch")
	assert.Equal(t, keyMap, model.keyMap)
}

func TestInitReturnsNonNilCmd(t *testing.T) {
	t.Parallel()

	configValue := config.Defaults()
	audioFiles := &[]audio.File{}
	model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue, audioFiles)
	require.NoError(t, err)

	// Init returns initializationLoadConfigCommand, which calls config.Load against the real user config dir. We do
	// not execute it here because that would touch the real filesystem.
	command := model.Init()
	require.NotNil(t, command, "Init() returned nil cmd, want a non-nil tea.Cmd")
}

func TestNewTestUIAppliesWindowSize(t *testing.T) {
	t.Parallel()

	model := newTestUI(t)

	assert.Equal(t, 80, model.windowWidth)
	assert.Equal(t, 24, model.windowHeight)
	assert.Equal(t, uiInitializing, model.state)
	require.NotNil(t, model.initializationModel, "initializationModel is nil after newTestUI")
}

func TestBindingsForOmitsCurrentStateAction(t *testing.T) {
	t.Parallel()

	// TODO: this will not scale well.. ;x
	tests := []struct {
		name         string
		state        uiState
		wantPlaylist bool
		wantLibStats bool
	}{
		{
			name:         "initializing has no bindings",
			state:        uiInitializing,
			wantPlaylist: false,
			wantLibStats: false,
		},
		{
			name:         "playlist omits open playlist",
			state:        uiPlaylist,
			wantPlaylist: false,
			wantLibStats: true,
		},
		{
			name:         "library stats omits open library stats",
			state:        uiLibraryStats,
			wantPlaylist: true,
			wantLibStats: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			configValue := config.Defaults()
			model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue, &[]audio.File{})
			require.NoError(t, err)

			bindings := model.commandBindingsFor(test.state)

			var gotPlaylist, gotLibStats bool
			for _, binding := range bindings {
				switch binding.Action.(type) {
				case action.OpenPlaylistAction:
					gotPlaylist = true
				case action.OpenLibraryStatsAction:
					gotLibStats = true
				}
			}

			assert.Equal(t, test.wantPlaylist, gotPlaylist, "playlist binding presence mismatch")
			assert.Equal(t, test.wantLibStats, gotLibStats, "library stats binding presence mismatch")
		})
	}
}

func TestSetStateRefreshesWhichKeyBindings(t *testing.T) {
	t.Parallel()

	configValue := config.Defaults()
	model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue, &[]audio.File{})
	require.NoError(t, err)
	model.setState(uiLibraryStats)

	// Leader then 'P' from the stats state navigates to the playlist.
	model.Update(tea.KeyPressMsg{Code: ' '})
	updatedModel, _ := model.Update(tea.KeyPressMsg{Code: 'P'})
	model = updatedModel.(*UIModel)
	assert.Equal(t, uiPlaylist, model.state, "'P' through the whichkey card must navigate to the playlist")

	// Without leaving the playlist, leader then 'P' again must be a no-op.
	model.Update(tea.KeyPressMsg{Code: ' '})
	updatedModel, _ = model.Update(tea.KeyPressMsg{Code: 'P'})
	model = updatedModel.(*UIModel)
	assert.Equal(t, uiPlaylist, model.state, "'P' in the playlist state must not toggle away from the playlist")
}
