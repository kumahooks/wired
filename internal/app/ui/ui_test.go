package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/config"
	"wired/internal/core/testutil"
)

// newTestUI builds a *UIModel with the default keymap, default theme, and a *config.Config populated from Defaults, then
// applies a fixed WindowSizeMsg so dimensions are deterministic.
func newTestUI(t *testing.T) *UIModel {
	t.Helper()

	configValue := config.Defaults()
	model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue)
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

	model, err := New(context.Background(), keyMap, &configValue)
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
	model, err := New(context.Background(), testutil.DefaultKeyMap(t), &configValue)
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
