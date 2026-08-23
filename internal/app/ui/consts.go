package ui

// The view is drawn based on these states below.
type uiState uint8

const (
	uiInitializing uiState = iota
	uiIdle
)

// Constants for the UI dimensions.
const (
	minWindowHeight = 20
	minWindowWidth  = 20
)

// countFilesProgressChannelBuffer is the size of the count progress channel.
const countFilesProgressChannelBuffer = 64
