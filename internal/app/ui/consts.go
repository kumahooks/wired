package ui

// The view is drawn based on these states below.
type uiState uint8

const (
	uiInitializing uiState = iota
	uiPlaylist
	uiLibraryStats
)

// Constants for the UI dimensions.
const (
	minWindowHeight = 20
	minWindowWidth  = 20
)

// fetchFilesProgressChannelBuffer is the size of the fetch progress channel.
const fetchFilesProgressChannelBuffer = 64
