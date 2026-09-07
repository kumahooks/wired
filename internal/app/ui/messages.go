package ui

import (
	"context"

	"wired/internal/core/audio"
	"wired/internal/core/config"
)

// configLoadOrigin tags which surface triggered a config load, deciding the feedback strategy in handleConfigLoadedMessage.
type configLoadOrigin int

const (
	configLoadOriginInit configLoadOrigin = iota
	configLoadOriginUser
)

// configLoadedMessage is produced by configLoadCommand when config.Load completes, for any origin.
type configLoadedMessage struct {
	config              *config.Config
	isConfigDefaults    bool
	invalidLibraryPaths []string
	err                 error
	origin              configLoadOrigin
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
	onlyNew         bool
}

// discoverFilesResultMessage is produced when the discovery finishes, carrying the discovered files and the shared
// progress reporter the metatag parse continues on.
type discoverFilesResultMessage struct {
	library    *audio.Library
	progress   *audio.DiscoveryProgress
	err        error
	generation uint64
	onlyNew    bool
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

// metatagParseResultMessage is produced when the metatag parse finishes, carrying the count of files it attempted.
type metatagParseResultMessage struct {
	parsedCount int
	err         error
	generation  uint64
}

// notificationExpireMessage is produced by notificationExpireCommand after a notification's lifetime ends.
type notificationExpireMessage struct{}
