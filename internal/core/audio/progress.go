package audio

import "sync/atomic"

// DiscoveryProgress is the live progress of a library discovery.
type DiscoveryProgress struct {
	discoveredCount atomic.Int64 // discoveredCount is the current total of audio files seen by DiscoverFiles.
	parsedCount     atomic.Int64 // parsedCount is the current total of files seen by ParseFiles.
	discoveryDone   atomic.Bool  // discoveryDone flips to 1 when DiscoverFiles completes.
}

// NewDiscoveryProgress returns a zeroed DiscoveryProgress.
func NewDiscoveryProgress() *DiscoveryProgress {
	return &DiscoveryProgress{}
}

// AddDiscovered bumps the discovered-files counter by n.
func (progress *DiscoveryProgress) AddDiscovered(n int) {
	progress.discoveredCount.Add(int64(n))
}

// AddParsed bumps the attempted-files counter by n.
func (progress *DiscoveryProgress) AddParsed(n int) {
	progress.parsedCount.Add(int64(n))
}

// SetDiscoveryDone marks the discovery phase as complete.
func (progress *DiscoveryProgress) SetDiscoveryDone() {
	progress.discoveryDone.Store(true)
}

// DiscoveredCount returns the current discovered-files count.
func (progress *DiscoveryProgress) DiscoveredCount() int {
	return int(progress.discoveredCount.Load())
}

// ParsedCount returns the current attempted-files count.
func (progress *DiscoveryProgress) ParsedCount() int {
	return int(progress.parsedCount.Load())
}

// DiscoveryDone reports whether the discovery phase has completed.
func (progress *DiscoveryProgress) DiscoveryDone() bool {
	return progress.discoveryDone.Load()
}
