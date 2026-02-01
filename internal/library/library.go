// Package library handles everything related to the library of files, from reading and writing
// to disk, to cache loading, storing and processing
package library

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/dhowden/tag"
)

var audioExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".m4a":  true,
	".wav":  true,
}

type songScanResult struct {
	path string
	song *Song
}

type FileScanningResult struct {
	Library *Library
	Error   error
}

type SongMetadata struct {
	SongName   string
	ArtistName string
	AlbumName  string
}

type Song struct {
	FileName string
	Metadata SongMetadata
}

type Album struct {
	AlbumName  string
	ArtistName string
	CoverImage string
	Songs      []*Song
}

type Artist struct {
	Albums []*Album
}

type Library struct {
	Songs   map[string]*Song   // file_path -> song structure
	Artists map[string]*Artist // artist_name -> artist structure
}

func New() *Library {
	return &Library{
		Songs:   map[string]*Song{},
		Artists: map[string]*Artist{},
	}
}

func (library *Library) AddSong(filePath string, song *Song) {
	library.Songs[filePath] = song

	artistName := song.Metadata.ArtistName
	artist, ok := library.Artists[artistName]
	if !ok {
		artist = &Artist{}
		library.Artists[artistName] = artist
	}

	albumName := song.Metadata.AlbumName
	var album *Album

	for _, a := range artist.Albums {
		if a.AlbumName == albumName {
			album = a
			break
		}
	}

	if album == nil {
		album = &Album{
			AlbumName:  albumName,
			ArtistName: artistName,
		}

		artist.Albums = append(artist.Albums, album)
	}

	album.Songs = append(album.Songs, song)
}

func LoadLibrary() *Library {
	return LoadCache()
}

func CountFiles(ctx context.Context, libraryPath string) (int, error) {
	count := 0

	err := filepath.WalkDir(libraryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == libraryPath {
				return err
			}

			return nil
		}

		if d.IsDir() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !isAudioFile(path) {
			return nil
		}

		count++

		return nil
	})

	return count, err
}

// Scan goes through every file in the path, returning every song file scanned
// there's a minor case where CountFiles and Scan values differ if a file is modified
// in the middle of the scan... I don't think I care about this since it won't matter much
func Scan(ctx context.Context, libraryPath string, channel chan<- int) (*Library, error) {
	var paths []string

	// we technically already walk the files with CountFiles
	// but honestly this doesn't matter that much, this is very fast
	err := filepath.WalkDir(libraryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == libraryPath {
				return err
			}

			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !isAudioFile(path) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		paths = append(paths, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	workersTotal := min(runtime.NumCPU()*4, len(paths))
	if workersTotal == 0 {
		return New(), nil
	}

	pathsChannel := make(chan string, workersTotal)
	resultsChannel := make(chan songScanResult, workersTotal)

	var waitGroup sync.WaitGroup

	for range workersTotal {
		waitGroup.Go(func() {
			for path := range pathsChannel {
				select {
				case <-ctx.Done():
					return
				default:
				}

				metadata := readMetadata(path)

				select {
				case resultsChannel <- songScanResult{
					path: path,
					song: &Song{
						FileName: filepath.Base(path),
						Metadata: metadata,
					},
				}:
				case <-ctx.Done():
					return
				}
			}
		})
	}

	go func() {
		defer close(pathsChannel)

		for _, p := range paths {
			select {
			case <-ctx.Done():
				return
			case pathsChannel <- p:
			}
		}
	}()

	go func() {
		waitGroup.Wait()
		close(resultsChannel)
	}()

	library := New()
	count := 0

	for result := range resultsChannel {
		select {
		case <-ctx.Done():
			return library, ctx.Err()
		default:
		}

		library.AddSong(result.path, result.song)

		count++

		if count == len(paths) {
			channel <- count
		} else {
			select {
			case channel <- count:
			default:
			}
		}
	}

	if ctx.Err() != nil {
		return library, ctx.Err()
	}

	return library, nil
}

func isAudioFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return audioExtensions[extension]
}

func readMetadata(path string) SongMetadata {
	resultMetadata := SongMetadata{}

	// TODO: do we want to save/show/do anything to errors?

	f, err := os.Open(path)
	if err == nil {
		defer f.Close()

		// TODO: tag doesn't support .wav
		// not sure how to proceed - maybe implementing the reader myself?
		// tag itself is quite old and might not be the ideal solution here
		metadata, err := tag.ReadFrom(f)

		if err == nil {
			resultMetadata.SongName = metadata.Title()
			resultMetadata.ArtistName = metadata.Artist()
			resultMetadata.AlbumName = metadata.Album()
		}
	}

	if resultMetadata.SongName == "" {
		name := filepath.Base(path)
		resultMetadata.SongName = strings.TrimSuffix(name, filepath.Ext(name))
	}

	if resultMetadata.ArtistName == "" {
		resultMetadata.ArtistName = "Unknown Artist"
	}

	if resultMetadata.AlbumName == "" {
		resultMetadata.AlbumName = "Unknown Album"
	}

	return resultMetadata
}
