package wired

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/config"
)

func TestNewReturnsOrchestrator(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	require.NoError(t, err)
	require.NotNil(t, orchestrator)

	require.NotNil(t, orchestrator.Config())
	assert.Equal(t, config.Defaults().Title, orchestrator.Config().Title)

	assert.NotNil(t, orchestrator.cancelContext, "cancelContext = nil, want non-nil")
	assert.NotNil(t, orchestrator.cancelFunction, "cancelFunction = nil, want non-nil")
	assert.Nil(t, orchestrator.teaProgram, "teaProgram = non-nil, want nil before Run")
	assert.NotNil(t, orchestrator.uiModel, "uiModel = nil, want non-nil")
}

func TestNewDoesNotCallConfigLoad(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	_, err := New(context.Background())
	require.NoError(t, err)

	configFile := filepath.Join(configHome, "wire_d", "config.toml")
	_, statErr := os.Stat(configFile)
	assert.True(t, os.IsNotExist(statErr), "config file was created at %q; New should not call config.Load", configFile)
}

func TestShutdownCancelsContext(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	require.NoError(t, err)

	orchestrator.Shutdown()

	assert.ErrorIs(t, orchestrator.cancelContext.Err(), context.Canceled)
}

func TestShutdownIsIdempotent(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	require.NoError(t, err)

	orchestrator.Shutdown()
	orchestrator.Shutdown()

	assert.ErrorIs(t, orchestrator.cancelContext.Err(), context.Canceled)
}

func TestNotifyTeaNilProgramIsNoOp(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	require.NoError(t, err)

	assert.NotPanics(t, func() { orchestrator.NotifyTea(nil) })
}

func TestConfigReturnsSharedPointer(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	require.NoError(t, err)

	assert.Equal(t, orchestrator.config, orchestrator.Config())
}
