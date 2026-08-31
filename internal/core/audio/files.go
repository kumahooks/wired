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

var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".flac": {},
	".ogg":  {},
	".m4a":  {},
	".wav":  {},
}

// File is a single audio file discovered in a library.
type File struct {
	FileName  string
	FilePath  string
	SizeBytes int64
}

// FetchFiles goes through every file in the sub-trees at rootPaths and returns every found audio file. When audioFiles
// is non-nil, each discovered file is appended to it. When countChannel is non-nil, the running total is emitted every
// progressInterval files and once at the end of each root walk.
func FetchFiles(
	ctx context.Context,
	rootPaths []string,
	audioFiles *[]File,
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

			fileSize := int64(0)
			if info, infoErr := entry.Info(); infoErr == nil {
				fileSize = info.Size()
			}

			if audioFiles != nil {
				*audioFiles = append(
					*audioFiles,
					File{FileName: filepath.Base(path), FilePath: path, SizeBytes: fileSize},
				)
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
func ScanFiles(audioFiles []File) {
	return
}

func isAudio(path string) bool {
	fileExtension := strings.ToLower(filepath.Ext(path))
	_, ok := audioExtensions[fileExtension]

	return ok
}
