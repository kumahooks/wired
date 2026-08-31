// Package audio deals with audio file primitives.
package audio

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

// progressInterval is how many files we count between two progress emissions.
const progressInterval = 32

// TODO: can't I retrieve through taglib whether a file is an audiofile without having to map like this...?
var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".flac": {},
	".ogg":  {},
	".m4a":  {},
	".wav":  {},
}

// AudioFile is a single audio file within a library.
type AudioFile struct {
	Title      string
	Artist     string
	Album      string
	Comment    string
	Genre      string
	Year       string
	Track      string
	Length     int
	Bitrate    int
	Samplerate int
	Channels   int
	SizeBytes  int64
}

// Library holds every known audio file, indexed by their path, and by their artist and album.
type Library struct {
	File     map[string]*AudioFile   // all files are loaded in a map keyed by their full filepath.
	ByAlbum  map[string][]*AudioFile // keyed by "artist:album", references a file within File.
	ByArtist map[string][]*AudioFile // keyed by "artist", references a file within File.
}

func NewLibrary() *Library {
	return &Library{
		File:     make(map[string]*AudioFile),
		ByAlbum:  make(map[string][]*AudioFile),
		ByArtist: make(map[string][]*AudioFile),
	}
}

// Add inserts a file under the given path as an unmetatagged AudioFile.
func (library *Library) Add(path string, sizeBytes int64) {
	library.File[path] = &AudioFile{SizeBytes: sizeBytes}
}

// FilesCount returns the number of audio files currently held.
func (library *Library) FilesCount() int {
	return len(library.File)
}

// FetchFiles goes through every file in the sub-trees at rootPaths and inserts every found audio file into the library.
// When library is non-nil, each discovered file is added to it. When progressChannel is non-nil, the running total is
// emitted every progressInterval files and once at the end of each root walk.
func FetchFiles(
	ctx context.Context,
	rootPaths []string,
	library *Library,
	progressChannel chan<- int,
) (int, error) {
	filesCount := 0

	for _, rootPath := range rootPaths {
		currentRoot := rootPath

		walkDirectory := func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				// If the very first directory errors, we should return it. Otherwise we just continue.
				if path == currentRoot {
					return err
				}

				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if entry.IsDir() {
				return nil
			}

			if !isAudio(path) {
				return nil
			}

			filesCount++

			var fileSizeBytes int64
			if info, infoErr := entry.Info(); infoErr == nil {
				fileSizeBytes = info.Size()
			}

			if library != nil {
				library.Add(path, fileSizeBytes)
			}

			if progressChannel != nil && filesCount%progressInterval == 0 {
				select {
				case progressChannel <- filesCount:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		}

		err := filepath.WalkDir(rootPath, walkDirectory)
		if err != nil {
			return filesCount, err
		}

		// Emit a final progress value for this root so the live counter is accurate at the moment of completion.
		if progressChannel != nil {
			select {
			case progressChannel <- filesCount:
			case <-ctx.Done():
				return filesCount, ctx.Err()
			}
		}
	}

	return filesCount, nil
}

// ScanFiles goes through every fetched path, retrieving and updating each audio file's metatag information.
// TODO: finish this
func ScanFiles(library *Library) {
}

func isAudio(path string) bool {
	fileExtension := strings.ToLower(filepath.Ext(path))
	_, ok := audioExtensions[fileExtension]

	return ok
}
