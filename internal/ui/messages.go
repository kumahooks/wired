package ui

import (
	"time"

	"wired/internal/config"
)

type LoadConfigMsg struct {
	Config *config.Config
	Err    error
}

type HeartbeatMsg time.Time
