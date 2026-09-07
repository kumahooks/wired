// Package audio deals with audio file primitives.
package audio

import (
	"context"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kumahooks/wiretag"
)

// AudioFile is a single audio file within a library.
type AudioFile struct {
	Path       string
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
	library.File[path] = &AudioFile{Path: path, SizeBytes: sizeBytes}
}

// Reset wipes the library's in-memory state.
func (library *Library) Reset() {
	library.File = make(map[string]*AudioFile)
	library.ByAlbum = make(map[string][]*AudioFile)
	library.ByArtist = make(map[string][]*AudioFile)
}

// FilesCount returns the number of audio files currently held.
func (library *Library) FilesCount() int {
	return len(library.File)
}

// DiscoverFiles goes through every file in the sub-trees at rootPaths and inserts every found audio file into the library.
// When library is non-nil, each discovered file is added to it. When progress is non-nil, the running total is reported
// through it for the UI's progress reads.
func DiscoverFiles(
	ctx context.Context,
	rootPaths []string,
	library *Library,
	progress *DiscoveryProgress,
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

			if library != nil {
				if _, known := library.File[path]; known {
					return nil
				}
			}

			filesCount++

			var fileSizeBytes int64
			if info, infoErr := entry.Info(); infoErr == nil {
				fileSizeBytes = info.Size()
			}

			if library != nil {
				library.Add(path, fileSizeBytes)
			}

			if progress != nil {
				progress.AddDiscovered(1)
			}

			return nil
		}

		err := filepath.WalkDir(rootPath, walkDirectory)
		if err != nil {
			return filesCount, err
		}
	}

	if progress != nil {
		progress.SetDiscoveryDone()
	}

	return filesCount, nil
}

// ParseFiles goes through the already-discovered files and parses each one's metatags sequentially.
// TODO: this is currently sequentially parsing files. for HDD this is fine and preferable, but it's very sub-optimal
// for SSDs...
// For HDDS (experiments to see if improves speed):
// - It could first fetch each file's inode number (in bulk!) and walk sequentially through it;
// - Separate file reading and taglib parsing in two different threads, sending the file to taglib through RAM;
// - Potentially check if POSIX_FADV_SEQUENTIAL helps;
func ParseFiles(ctx context.Context, files []*AudioFile, progress *DiscoveryProgress) (int, error) {
	var parsedCount int

	for _, audioFile := range files {
		if err := ctx.Err(); err != nil {
			return parsedCount, err
		}

		parseAudioFileMetaTag(audioFile)
		parsedCount++

		if progress != nil {
			progress.AddParsed(1)
		}
	}

	return parsedCount, nil
}

// parseAudioFileMetaTag opens the file at audioFile's path and fills it with its metatag and audio properties.
func parseAudioFileMetaTag(audioFile *AudioFile) {
	if audioFile == nil {
		return
	}

	file, err := wiretag.Open(audioFile.Path)
	if err != nil {
		return
	}
	defer file.Close()

	// TODO: we made wiretag return idiomatic Go errors, but honestly this is kinda bad... can we improve there?
	title, _ := file.Title()
	artist, _ := file.Artist()
	album, _ := file.Album()
	comment, _ := file.Comment()
	genre, _ := file.Genre()
	year, _ := file.Year()
	track, _ := file.Track()
	length, _ := file.AudioLength()
	bitrate, _ := file.AudioBitrate()
	samplerate, _ := file.AudioSampleRate()
	channels, _ := file.AudioChannels()

	audioFile.Title = title
	audioFile.Artist = artist
	audioFile.Album = album
	audioFile.Comment = comment
	audioFile.Genre = genre
	audioFile.Year = strconv.Itoa(year)
	audioFile.Track = strconv.Itoa(track)
	audioFile.Length = length
	audioFile.Bitrate = bitrate
	audioFile.Samplerate = samplerate
	audioFile.Channels = channels
}

// UntaggedFiles returns the library's files that have no parsed metatags.
func (library *Library) UntaggedFiles() []*AudioFile {
	untaggedFiles := make([]*AudioFile, 0, len(library.File))

	for _, audioFile := range library.File {
		if audioFile == nil || audioFile.Artist != "" || audioFile.Title != "" || audioFile.Album != "" {
			continue
		}

		untaggedFiles = append(untaggedFiles, audioFile)
	}

	return untaggedFiles
}

// BuildLibraryIndexes creates the ByAlbum and ByArtist indexes from the parsed files.
func BuildLibraryIndexes(library *Library) {
	library.ByAlbum = make(map[string][]*AudioFile)
	library.ByArtist = make(map[string][]*AudioFile)

	for _, audioFile := range library.File {
		if audioFile == nil || audioFile.Artist == "" {
			continue
		}

		library.ByArtist[audioFile.Artist] = append(library.ByArtist[audioFile.Artist], audioFile)

		if audioFile.Album == "" {
			continue
		}

		albumKey := audioFile.Artist + ":" + audioFile.Album
		library.ByAlbum[albumKey] = append(library.ByAlbum[albumKey], audioFile)
	}
}

// TODO: can't I retrieve through taglib whether a file is an audiofile without having to map like this...?
var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".flac": {},
	".ogg":  {},
	".m4a":  {},
	".wav":  {},
}

func isAudio(path string) bool {
	fileExtension := strings.ToLower(filepath.Ext(path))
	_, ok := audioExtensions[fileExtension]

	return ok
}
