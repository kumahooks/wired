package librarystats

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/app/ui/action"
	"wired/internal/core/audio"
	"wired/internal/core/keymap"
	"wired/internal/core/testutil"
)

func defaultKeyMap(t *testing.T) keymap.KeyMap {
	t.Helper()

	return testutil.DefaultKeyMap(t)
}

func TestNewSeedsButtonsAndDefaults(t *testing.T) {
	t.Parallel()

	keyMap := defaultKeyMap(t)
	library := audio.NewLibrary()
	model := New(keyMap, library)

	require.Len(t, model.buttons, 2)
	assert.Equal(t, scanFullLibraryAction, model.buttons[0].action)
	assert.Equal(t, scanNewLibraryAction, model.buttons[1].action)
	assert.Empty(t, model.libraryPaths)
	assert.Equal(t, audio.Stats{}, model.libraryStats)
	assert.Equal(t, library, model.library)
	assert.Equal(t, keyMap, model.keyMap)

	wantStyle := newStyle(testutil.DefaultTheme())
	assert.Equal(t, wantStyle, model.style)
}

func TestApplyThemeRebuildsStyle(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())

	customTheme := testutil.CustomTheme()

	model.ApplyTheme(customTheme)

	wantStyle := newStyle(customTheme)
	assert.Equal(t, wantStyle, model.style)
}

func TestApplyKeyMapStoresKeyMap(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())

	replacementKeyMap := defaultKeyMap(t)

	model.ApplyKeyMap(replacementKeyMap)

	assert.Equal(t, replacementKeyMap, model.keyMap)
}

func TestSetLibraryPaths(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())

	paths := []string{"/mnt/music", "/mnt/podcasts"}

	model.SetLibraryPaths(paths)

	assert.Equal(t, paths, model.libraryPaths)
}

func TestComputeStats(t *testing.T) {
	t.Parallel()

	library := audio.NewLibrary()
	library.Add("/music/one.flac", 100)
	library.Add("/music/two.flac", 300)

	model := New(defaultKeyMap(t), library)

	model.ComputeStats()

	assert.Equal(t, 2, model.libraryStats.FilesCount)
	assert.Equal(t, int64(400), model.libraryStats.TotalBytes)
}

func TestStartDiscoveryResetsProgressState(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())
	model.isDiscovering = false
	model.isDiscoveryDone = true
	model.discoveredFilesCount = 42
	model.parsedMetatagCount = 42

	model.StartDiscovery()

	assert.True(t, model.isDiscovering)
	assert.False(t, model.isDiscoveryDone)
	assert.Zero(t, model.discoveredFilesCount)
	assert.Zero(t, model.parsedMetatagCount)
}

func TestSetProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		progress             *audio.DiscoveryProgress
		wantDiscoveredCount  int
		wantDiscoveryDone    bool
		wantParsedMetatag    int
		wantParsedMetatagSet bool
	}{
		{
			name:                "nil reporter is a no-op",
			progress:            nil,
			wantDiscoveredCount: 0,
		},
		{
			name:                "running discovery only updates discovered count",
			progress:            testutil.NewDiscoveryProgress(7, 0, false),
			wantDiscoveredCount: 7,
		},
		{
			name:                 "finished discovery also updates parsed count",
			progress:             testutil.NewDiscoveryProgress(7, 5, true),
			wantDiscoveredCount:  7,
			wantDiscoveryDone:    true,
			wantParsedMetatag:    5,
			wantParsedMetatagSet: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := New(defaultKeyMap(t), audio.NewLibrary())

			model.SetProgress(test.progress)

			assert.Equal(t, test.wantDiscoveredCount, model.discoveredFilesCount)
			assert.Equal(t, test.wantDiscoveryDone, model.isDiscoveryDone)
			if test.wantParsedMetatagSet {
				assert.Equal(t, test.wantParsedMetatag, model.parsedMetatagCount)
			}
		})
	}
}

func TestDiscoveredFilesCount(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())
	model.discoveredFilesCount = 12

	assert.Equal(t, 12, model.DiscoveredFilesCount())
}

func TestFinishDiscoveryClearsState(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())
	model.isDiscovering = true
	model.isDiscoveryDone = true
	model.parsedMetatagCount = 9

	model.FinishDiscovery()

	assert.False(t, model.isDiscovering)
	assert.False(t, model.isDiscoveryDone)
	assert.Zero(t, model.parsedMetatagCount)
}

func TestHandleMessageNonKeyPressIsIgnored(t *testing.T) {
	t.Parallel()

	model := New(defaultKeyMap(t), audio.NewLibrary())

	gotAction := model.HandleMessage(tea.WindowSizeMsg{})

	assert.Equal(t, action.NoAction{}, gotAction)
}
