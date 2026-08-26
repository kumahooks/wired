package audio

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

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

// testFetchChannelBuffer is large enough that the walk never blocks on the drainer during tests.
const testFetchChannelBuffer = 64

// plantManyAudioFiles creates count audio files under root/sub/ so the walk crosses progressInterval at least once.
func plantManyAudioFiles(t *testing.T, root string, count int) {
	t.Helper()

	subDir := filepath.Join(root, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	for index := range count {
		path := filepath.Join(subDir, "track"+strconv.Itoa(index)+".mp3")
		require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	}
}

func TestFetchFilesTotal(t *testing.T) {
	t.Parallel()

	rootOne := t.TempDir()
	rootTwo := t.TempDir()

	wantOne := buildAudioTree(t, rootOne)
	wantTwo := buildAudioTree(t, rootTwo)

	got, err := FetchFiles(context.Background(), []string{rootOne, rootTwo}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, wantOne+wantTwo, got)
}

func TestFetchFilesPopulatesAudioFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	var files []File
	got, err := FetchFiles(context.Background(), []string{root}, &files, nil)
	require.NoError(t, err)

	assert.Equal(t, want, got)
	require.Len(t, files, want)

	// Every populated file name should be an audio file by base name.
	for _, file := range files {
		assert.True(t, isAudio(file.FileName), "non-audio file in slice: %q", file.FileName)
	}
}

func TestFetchChannelEmissions(t *testing.T) {
	t.Parallel()

	rootOne := t.TempDir()
	rootTwo := t.TempDir()

	plantManyAudioFiles(t, rootOne, 40)
	plantManyAudioFiles(t, rootTwo, 10)

	fetchChannel := make(chan int, testFetchChannelBuffer)

	_, err := FetchFiles(context.Background(), []string{rootOne, rootTwo}, nil, fetchChannel)
	require.NoError(t, err)

	// FetchFiles does not close the channel.
	close(fetchChannel)

	// Drain the channel with a timeout.
	var values []int
	for {
		select {
		case value, ok := <-fetchChannel:
			if !ok {
				goto done
			}

			values = append(values, value)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for fetchChannel, got %d values", len(values))
		}
	}

done:
	require.GreaterOrEqual(t, len(values), 3, "expected at least 3 emissions, got %d: %v", len(values), values)

	// The last value should be the grand total.
	assert.Equal(t, 50, values[len(values)-1])
}

func TestFetchFilesCancelMidWalk(t *testing.T) {
	root := t.TempDir()
	plantManyAudioFiles(t, root, 40)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fetchChannel := make(chan int, testFetchChannelBuffer)

	_, err := FetchFiles(ctx, []string{root}, nil, fetchChannel)
	require.Error(t, err)
	assert.ErrorIs(t, err, ctx.Err())

	// Drain any stray progress values so the sender is not blocked.
	for {
		select {
		case <-fetchChannel:
		default:
			return
		}
	}
}

func TestScanFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	fetchChannel := make(chan int, 1)

	files, err := ScanFiles(context.Background(), []string{root}, fetchChannel)
	require.NoError(t, err)

	assert.Len(t, files, want)

	select {
	case value := <-fetchChannel:
		assert.Equal(t, want, value)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fetchChannel")
	}

	// Exactly one value: the channel should now be empty.
	select {
	case value := <-fetchChannel:
		t.Errorf("unexpected extra fetchChannel value %d", value)
	default:
	}
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

func TestFetchFilesNonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := FetchFiles(context.Background(), []string{"/this/path/does/not/exist"}, nil, nil)
	require.Error(t, err)
}

func TestFetchFilesEmptyRoots(t *testing.T) {
	t.Parallel()

	got, err := FetchFiles(context.Background(), []string{}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, got)
}
