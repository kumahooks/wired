package audio

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
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
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				t.Fatalf("mkdir %q: %v", fullPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir parent for %q: %v", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %q: %v", fullPath, err)
		}

		if isAudio(entry.path) {
			audioCount++
		}
	}

	return audioCount
}

// testCountChannelBuffer is large enough that the walk never blocks on the drainer during tests.
const testCountChannelBuffer = 64

// plantManyAudioFiles creates count audio files under root/sub/ so the walk crosses progressInterval at least once.
func plantManyAudioFiles(t *testing.T, root string, count int) {
	t.Helper()

	subDir := filepath.Join(root, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", subDir, err)
	}

	for index := range count {
		path := filepath.Join(subDir, "track"+pad(index)+".mp3")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %q: %v", path, err)
		}
	}
}

func pad(index int) string {
	if index < 10 {
		return "0" + itoa(index)
	}

	return itoa(index)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var buf [20]byte

	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}

	return string(buf[pos:])
}

func TestCountFilesTotal(t *testing.T) {
	t.Parallel()

	rootOne := t.TempDir()
	rootTwo := t.TempDir()

	wantOne := buildAudioTree(t, rootOne)
	wantTwo := buildAudioTree(t, rootTwo)

	got, err := CountFiles(context.Background(), []string{rootOne, rootTwo}, nil, nil)
	if err != nil {
		t.Fatalf("CountFiles error: %v", err)
	}

	if want := wantOne + wantTwo; got != want {
		t.Errorf("CountFiles total = %d, want %d", got, want)
	}
}

func TestCountFilesPopulatesAudioFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	var files []File
	got, err := CountFiles(context.Background(), []string{root}, &files, nil)
	if err != nil {
		t.Fatalf("CountFiles error: %v", err)
	}

	if got != want {
		t.Errorf("CountFiles count = %d, want %d", got, want)
	}
	if len(files) != want {
		t.Fatalf("len(files) = %d, want %d", len(files), want)
	}

	// Every populated file name should be an audio file by base name.
	for _, file := range files {
		if !isAudio(file.FileName) {
			t.Errorf("non-audio file in slice: %q", file.FileName)
		}
	}
}

func TestCountChannelEmissions(t *testing.T) {
	t.Parallel()

	rootOne := t.TempDir()
	rootTwo := t.TempDir()

	plantManyAudioFiles(t, rootOne, 40)
	plantManyAudioFiles(t, rootTwo, 10)

	countChannel := make(chan int, testCountChannelBuffer)

	_, err := CountFiles(context.Background(), []string{rootOne, rootTwo}, nil, countChannel)
	if err != nil {
		t.Fatalf("CountFiles error: %v", err)
	}

	// CountFiles does not close the channel.
	close(countChannel)

	// Drain the channel with a timeout.
	var values []int
	for {
		select {
		case value, ok := <-countChannel:
			if !ok {
				goto done
			}

			values = append(values, value)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for countChannel, got %d values", len(values))
		}
	}

done:
	if len(values) < 3 {
		t.Errorf("expected at least 3 emissions, got %d: %v", len(values), values)
	}

	// The last value should be the grand total.
	if got := values[len(values)-1]; got != 50 {
		t.Errorf("last emission = %d, want 50", got)
	}
}

func TestCountFilesCancelMidWalk(t *testing.T) {
	root := t.TempDir()
	plantManyAudioFiles(t, root, 40)

	context, cancel := context.WithCancel(context.Background())
	cancel()

	countChannel := make(chan int, testCountChannelBuffer)

	_, err := CountFiles(context, []string{root}, nil, countChannel)
	if !errors.Is(err, context.Err()) {
		t.Errorf("CountFiles with canceled context err = %v, want %v", err, context.Err())
	}

	// Drain any stray progress values so the sender is not blocked.
	for {
		select {
		case <-countChannel:
		default:
			return
		}
	}
}

func TestScanFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := buildAudioTree(t, root)

	countChannel := make(chan int, 1)

	files, err := ScanFiles(context.Background(), []string{root}, countChannel)
	if err != nil {
		t.Fatalf("ScanFiles error: %v", err)
	}

	if got := len(files); got != want {
		t.Errorf("ScanFiles len = %d, want %d", got, want)
	}

	select {
	case value := <-countChannel:
		if value != want {
			t.Errorf("countChannel value = %d, want %d", value, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for countChannel")
	}

	// Exactly one value: the channel should now be empty.
	select {
	case value := <-countChannel:
		t.Errorf("unexpected extra countChannel value %d", value)
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

			if got := isAudio(test.path); got != test.want {
				t.Errorf("isAudio(%q) = %v, want %v", test.path, got, test.want)
			}
		})
	}
}

func TestCountFilesNonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := CountFiles(context.Background(), []string{"/this/path/does/not/exist"}, nil, nil)
	if err == nil {
		t.Fatal("CountFiles with non-existent root want error, got nil")
	}
}

func TestCountFilesEmptyRoots(t *testing.T) {
	t.Parallel()

	got, err := CountFiles(context.Background(), []string{}, nil, nil)
	if err != nil {
		t.Fatalf("CountFiles error: %v", err)
	}
	if got != 0 {
		t.Errorf("CountFiles empty roots = %d, want 0", got)
	}
}
