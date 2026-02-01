package ui

import (
	"context"
	"time"

	"wired/internal/config"
	"wired/internal/library"
)

type LoadConfigMsg struct {
	Config                  *config.Config
	Errors                  []error
	MusicLibraryPathCleared bool
}

type HeartbeatMsg time.Time

type FileScanningState struct {
	Total           int
	Current         int
	ProgressChannel <-chan int
	ResultChannel   <-chan library.FileScanningResult
	CancelContext   context.CancelFunc
}

type ScanStartMsg struct {
	Total           int
	ProgressChannel <-chan int
	ResultChannel   <-chan library.FileScanningResult
}

type ScanProgressMsg struct {
	Current int
}

type ScanCompleteMsg struct {
	Library *library.Library
	Error   error
}

type LoadLibraryMsg struct {
	Library *library.Library
}
