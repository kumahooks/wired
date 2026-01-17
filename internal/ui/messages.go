package ui

import "wired/internal/config"

type ConfigLoadedMsg struct {
	Config *config.Config
	Err    error
}
