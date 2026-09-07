package audio

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildAudioTree plants a deterministic directory tree under root and returns the expected audio file count.
func buildAudioTree(t *testing.T, root string) int {
	t.Helper()

	entries := []struct {
		path  string
		isDir bool
	}{
		{"song1.mp3", false},
		{"song2.flac", false},
		{"song3.wav", false},
		{"song4.m4a", false},
		{"song5.ogg", false},
		{"readme.txt", false},
		{"noextfile", false},
		{"UPPER.MP3", false},
		{"album", true},
		{"album/track1.mp3", false},
		{"album/track2.flac", false},
		{"album/cover.png", false},
		{"album/deep", true},
		{"album/deep/track3.ogg", false},
		{"album/deep/track4.wav", false},
		{"album/deep/.hidden", false},
		{"double.mp3.bak", false},
	}

	audioCount := 0
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.path)

		if entry.isDir {
			require.NoError(t, os.MkdirAll(fullPath, 0o755))
			continue
		}

		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte("x"), 0o644))

		if isAudio(entry.path) {
			audioCount++
		}
	}

	return audioCount
}

// plantManyAudioFiles creates count audio files under root/sub/.
func plantManyAudioFiles(t *testing.T, root string, count int) {
	t.Helper()

	subDir := filepath.Join(root, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	for index := range count {
		path := filepath.Join(subDir, "track"+strconv.Itoa(index)+".mp3")
		require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	}
}

func TestDiscoverFilesTotal(t *testing.T) {
	t.Parallel()

	rootOne := t.TempDir()
	rootTwo := t.TempDir()

	wantOne := buildAudioTree(t, rootOne)
	wantTwo := buildAudioTree(t, rootTwo)

	got, err := DiscoverFiles(context.Background(), []string{rootOne, rootTwo}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, wantOne+wantTwo, got)
}

func TestDiscoverFilesPopulatesLibrary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	library := NewLibrary()
	got, err := DiscoverFiles(context.Background(), []string{root}, library, nil)
	require.NoError(t, err)

	assert.Equal(t, want, got)
	require.Len(t, library.File, want)

	// Every discovered path should carry an AudioFile with the written byte length.
	for path, file := range library.File {
		assert.True(t, isAudio(path), "non-audio file in library: %q", path)
		assert.Equal(t, int64(1), file.SizeBytes, "unexpected size for %q", path)
	}
}

func TestDiscoverFilesReportsProgress(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	plantManyAudioFiles(t, root, 10)

	progress := NewDiscoveryProgress()

	_, err := DiscoverFiles(context.Background(), []string{root}, nil, progress)
	require.NoError(t, err)

	assert.Equal(t, 10, progress.DiscoveredCount())
	assert.True(t, progress.DiscoveryDone(), "discoveryDone should flip once the walk completes")
}

func TestDiscoverFilesCancelMidWalk(t *testing.T) {
	root := t.TempDir()
	plantManyAudioFiles(t, root, 40)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DiscoverFiles(ctx, []string{root}, nil, NewDiscoveryProgress())
	require.Error(t, err)
	assert.ErrorIs(t, err, ctx.Err())
}

func TestIsAudio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"lowercase mp3", "track.mp3", true},
		{"lowercase flac", "track.flac", true},
		{"lowercase wav", "track.wav", true},
		{"lowercase m4a", "track.m4a", true},
		{"lowercase ogg", "track.ogg", true},
		{"uppercase mp3", "TRACK.MP3", true},
		{"uppercase flac", "TRACK.FLAC", true},
		{"no extension", "track", false},
		{"dotfile no ext", ".hidden", false},
		{"double extension", "track.mp3.bak", false},
		{"text file", "readme.txt", false},
		{"png file", "cover.png", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, isAudio(test.path))
		})
	}
}

func TestDiscoverFilesNonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := DiscoverFiles(context.Background(), []string{"/this/path/does/not/exist"}, nil, nil)
	require.Error(t, err)
}

func TestDiscoverFilesEmptyRoots(t *testing.T) {
	t.Parallel()

	got, err := DiscoverFiles(context.Background(), []string{}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, got)
}

func TestDiscoverFilesSkipsKnownPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	library := NewLibrary()

	firstCount, err := DiscoverFiles(context.Background(), []string{root}, library, nil)
	require.NoError(t, err)
	assert.Equal(t, want, firstCount)

	// A second walk over the same tree must find nothing new, without any flag from the caller.
	secondCount, err := DiscoverFiles(context.Background(), []string{root}, library, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, secondCount)
	assert.Len(t, library.File, want, "the library should be untouched by the second walk")
}

func TestLibraryUntaggedFiles(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	library.Add("/untagged.flac", 1)
	library.Add("/parsed.flac", 1)
	library.Add("/titleonly.flac", 1)
	library.Add("/nilentry.mp3", 1)
	library.File["/nilentry.mp3"] = nil

	library.File["/parsed.flac"].Title = "Fool"
	library.File["/parsed.flac"].Artist = "bôa"
	library.File["/parsed.flac"].Album = "Twilight"
	library.File["/titleonly.flac"].Title = "Fool"

	untaggedFiles := library.UntaggedFiles()

	require.Len(t, untaggedFiles, 1)
	assert.Equal(t, "/untagged.flac", untaggedFiles[0].Path)
}

func TestParseFilesParsesMetatags(t *testing.T) {
	t.Parallel()

	validDir := t.TempDir()
	validPath := filepath.Join(validDir, "boa.flac")
	testData, err := os.ReadFile(filepath.Join("testdata", "complete_boa_fool.flac"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(validPath, testData, 0o644))

	invalidPath := filepath.Join(validDir, "missing.mp3")

	library := NewLibrary()
	library.Add(validPath, int64(len(testData)))
	library.Add(invalidPath, 1)

	files := librarySnapshot(library)

	parsedCount, err := ParseFiles(context.Background(), files, NewDiscoveryProgress())
	require.NoError(t, err)
	assert.Equal(t, 2, parsedCount)

	parsedFile := library.File[validPath]
	require.NotNil(t, parsedFile)
	assert.Equal(t, "Fool", parsedFile.Title)
	assert.Equal(t, "bôa", parsedFile.Artist)
	assert.NotEmpty(t, parsedFile.Album)

	BuildLibraryIndexes(library)

	assert.Equal(
		t,
		"bôa:"+parsedFile.Album,
		firstAlbumKey(library.ByAlbum),
		"expected album index entry for the parsed file",
	)
	assert.Len(t, library.ByArtist["bôa"], 1, "expected artist index entry for the parsed file")

	// The invalid file stays unmetatagged.
	brokenFile := library.File[invalidPath]
	require.NotNil(t, brokenFile)
	assert.Empty(t, brokenFile.Title)
}

func TestParseFilesReportsProgress(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	validDir := t.TempDir()
	validPath := filepath.Join(validDir, "boa.flac")
	testData, err := os.ReadFile(filepath.Join("testdata", "complete_boa_fool.flac"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(validPath, testData, 0o644))

	for index := range 4 {
		copyPath := filepath.Join(validDir, "copy"+strconv.Itoa(index)+".flac")
		require.NoError(t, os.WriteFile(copyPath, testData, 0o644))
		library.Add(copyPath, int64(len(testData)))
	}

	discoveryProgress := NewDiscoveryProgress()

	_, err = ParseFiles(context.Background(), librarySnapshot(library), discoveryProgress)
	require.NoError(t, err)

	assert.Equal(t, 4, discoveryProgress.ParsedCount(), "all attempts should be reported")
}

func TestParseFilesProgressCountsFailures(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	validDir := t.TempDir()
	testData, err := os.ReadFile(filepath.Join("testdata", "complete_boa_fool.flac"))
	require.NoError(t, err)

	// 20 files: 16 parseable, 4 nonexistent paths that fail to parse. Every attempt must count towards progress.
	for index := range 20 {
		path := filepath.Join(validDir, "file"+strconv.Itoa(index)+".flac")
		if index%6 == 0 {
			path = filepath.Join(validDir, "missing"+strconv.Itoa(index)+".mp3")
		}

		library.Add(path, int64(len(testData)))
	}

	discoveryProgress := NewDiscoveryProgress()

	_, err = ParseFiles(context.Background(), librarySnapshot(library), discoveryProgress)
	require.NoError(t, err)

	assert.Equal(t, 20, discoveryProgress.ParsedCount(), "all attempts (including failed parses) should count")
}

func TestParseFilesCancel(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	testData, err := os.ReadFile(filepath.Join("testdata", "complete_boa_fool.flac"))
	require.NoError(t, err)

	for range 4 {
		library.Add(filepath.Join(t.TempDir(), "boa.flac"), int64(len(testData)))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = ParseFiles(ctx, librarySnapshot(library), NewDiscoveryProgress())
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLibraryResetWipesState(t *testing.T) {
	t.Parallel()

	library := NewLibrary()
	library.Add("/some/path.flac", 1)
	BuildLibraryIndexes(library)

	require.Len(t, library.File, 1)

	library.Reset()

	assert.Empty(t, library.File)
	assert.Empty(t, library.ByAlbum)
	assert.Empty(t, library.ByArtist)
}

// librarySnapshot returns the file slice the UI's discovery result handler would hand to ParseFiles.
func librarySnapshot(library *Library) []*AudioFile {
	files := make([]*AudioFile, 0, len(library.File))
	for _, audioFile := range library.File {
		files = append(files, audioFile)
	}

	return files
}

// firstAlbumKey returns the first album index key, for assertions.
func firstAlbumKey(byAlbum map[string][]*AudioFile) string {
	for key := range byAlbum {
		return key
	}

	return ""
}
