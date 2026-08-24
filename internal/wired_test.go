package wired

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"wired/internal/core/config"
)

func TestNewReturnsOrchestrator(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if orchestrator == nil {
		t.Fatal("New returned nil orchestrator")
	}

	if orchestrator.Config() == nil {
		t.Fatal("Config() = nil, want non-nil")
	}
	if got := orchestrator.Config().Title; got != config.Defaults().Title {
		t.Errorf("Config().Title = %q, want %q", got, config.Defaults().Title)
	}

	if orchestrator.cancelContext == nil {
		t.Error("cancelContext = nil, want non-nil")
	}
	if orchestrator.cancelFunction == nil {
		t.Error("cancelFunction = nil, want non-nil")
	}
	if orchestrator.teaProgram != nil {
		t.Error("teaProgram = non-nil, want nil before Run")
	}
	if orchestrator.uiModel == nil {
		t.Error("uiModel = nil, want non-nil")
	}
}

func TestNewDoesNotCallConfigLoad(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	_, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	configFile := filepath.Join(configHome, "wire_d", "config.toml")
	if _, statErr := os.Stat(configFile); statErr == nil {
		t.Errorf("config file was created at %q; New should not call config.Load", configFile)
	}
}

func TestShutdownCancelsContext(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	orchestrator.Shutdown()

	if err := orchestrator.cancelContext.Err(); err != context.Canceled {
		t.Errorf("cancelContext.Err() = %v, want context.Canceled", err)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	orchestrator.Shutdown()
	orchestrator.Shutdown()

	if err := orchestrator.cancelContext.Err(); err != context.Canceled {
		t.Errorf("cancelContext.Err() after double shutdown = %v, want context.Canceled", err)
	}
}

func TestNotifyTeaNilProgramIsNoOp(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("NotifyTea panicked with nil teaProgram: %v", recovered)
		}
	}()

	orchestrator.NotifyTea(nil)
}

func TestConfigReturnsSharedPointer(t *testing.T) {
	t.Parallel()

	orchestrator, err := New(context.Background())
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	if orchestrator.Config() != orchestrator.config {
		t.Error("Config() does not return the shared config pointer")
	}
}
