package ui

import (
	"context"

	"wired/internal/core/audio"
	"wired/internal/core/config"
)

// initializationLoadConfigResultMessage is produced by initializationLoadConfigCommand when config.Load completes.
type initializationLoadConfigResultMessage struct {
	config              *config.Config
	isConfigDefaults    bool
	invalidLibraryPaths []string
	err                 error
}

// initializationLoadLibraryCacheResultMessage is produced by initializationLoadLibraryCacheCommand when the library
// cache lookup completes.
type initializationLoadLibraryCacheResultMessage struct {
	library *audio.Library
	err     error
}

// discoverFilesStartMessage is produced on demand by the user. It carries the cancel func that the cancel path shares,
// the DiscoveryProgress the discovery reports through, and the result channel a waiter command blocks on.
type discoverFilesStartMessage struct {
	progress        *audio.DiscoveryProgress
	result          <-chan discoverFilesResultMessage
	discoveryCancel context.CancelFunc
	generation      uint64
}

// discoverFilesResultMessage is produced when the discovery finishes, carrying the discovered files and the shared
// progress reporter the metatag parse continues on.
type discoverFilesResultMessage struct {
	library    *audio.Library
	progress   *audio.DiscoveryProgress
	err        error
	generation uint64
}

// metatagParseStartMessage is produced on demand by the user. It carries the DiscoveryProgress the parse reports
// through and the result channel a waiter command blocks on.
type metatagParseStartMessage struct {
	progress   *audio.DiscoveryProgress
	result     <-chan metatagParseResultMessage
	generation uint64
}

// discoveryProgressTickMessage is produced by discoveryProgressTickCommand on each tick while a discovery phase runs,
// carrying the progress reporter to read, and the generation of the discovery phase it belongs to.
type discoveryProgressTickMessage struct {
	progress   *audio.DiscoveryProgress
	generation uint64
}

// metatagParseResultMessage is produced when the metatag parse finishes, carrying the parsed counts.
type metatagParseResultMessage struct {
	err        error
	generation uint64
}

// notificationExpireMessage is produced by notificationExpireCommand after a notification's lifetime ends.
type notificationExpireMessage struct{}
