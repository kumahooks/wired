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
				*audioFiles = append(*audioFiles, File{FileName: filepath.Base(path), SizeBytes: fileSize})
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

// ScanFiles goes through every file in the sub-trees at rootPaths and saves every file name in a slice. It emits the
// final count on countChannel once at completion.
//
// TODO: eventually, it will be necessary to implement a metatag retrieval here for each audio file. This should be
// done using sync.WaitGroup with at least runtime.NumCPU()*4 workers, updating countChannel every processed file.
// also, doing FetchFiles twice is not necessary. we just need to retrieve the []audio.File here.
func ScanFiles(ctx context.Context, rootPaths []string, countChannel chan<- int) ([]File, error) {
	var audioFiles []File

	_, err := FetchFiles(ctx, rootPaths, &audioFiles, nil)
	if err != nil {
		return nil, err
	}

	select {
	case countChannel <- len(audioFiles):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return audioFiles, nil
}

func isAudio(path string) bool {
	fileExtension := strings.ToLower(filepath.Ext(path))
	_, ok := audioExtensions[fileExtension]

	return ok
}
