package ui

import (
	"time"

	"wired/internal/config"
)

type LoadConfigMsg struct {
	Config                  *config.Config
	Errors                  []error
	MusicLibraryPathCleared bool
}

type HeartbeatMsg time.Time
