package ui

import (
	"time"

	"wired/internal/config"
)

type LoadConfigMsg struct {
	Config *config.Config
	Errors []error
}

type HeartbeatMsg time.Time
