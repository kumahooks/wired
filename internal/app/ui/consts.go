package ui

import "time"

// The view is drawn based on these states below.
type uiState uint8

const (
	uiBootstrapping uiState = iota
	uiInitializing
	uiLibrary
	uiPlaylist
	uiLibraryStats
)

// Constants for the UI dimensions.
const (
	minWindowHeight = 20
	minWindowWidth  = 20
)

// discoveryProgressTickInterval is how often the UI ticks the discovery progress reporter while a discovery is running.
const discoveryProgressTickInterval = 100 * time.Millisecond
