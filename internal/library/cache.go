package library

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	dirPerm      = 0o755
	filePerm     = 0o644
	cacheVersion = 7
)

type songCache struct {
	FileName   string `json:"file_name"`
	SongName   string `json:"song_name"`
	ArtistName string `json:"artist_name"`
	AlbumName  string `json:"album_name"`
}

type libraryCache struct {
	Version int                   `json:"version"`
	Songs   map[string]*songCache `json:"songs"` // file_path -> song metadata
}

func (library *Library) SaveCache() error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	cache := libraryCache{
		Version: cacheVersion,
		Songs:   make(map[string]*songCache, len(library.Songs)),
	}

	for filePath, song := range library.Songs {
		cache.Songs[filePath] = &songCache{
			FileName:   song.FileName,
			SongName:   song.Metadata.SongName,
			ArtistName: song.Metadata.ArtistName,
			AlbumName:  song.Metadata.AlbumName,
		}
	}

	data, err := json.MarshalIndent(cache, "", "\t")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	return os.WriteFile(path, data, filePerm)
}

func LoadCache() *Library {
	path, err := getCachePath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cache libraryCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	if cache.Version != cacheVersion {
		return nil
	}

	if len(cache.Songs) == 0 {
		return nil
	}

	library := New()

	for filePath, cached := range cache.Songs {
		song := &Song{
			FileName: cached.FileName,
			Metadata: SongMetadata{
				SongName:   cached.SongName,
				ArtistName: cached.ArtistName,
				AlbumName:  cached.AlbumName,
			},
		}

		library.AddSong(filePath, song)
	}

	return library
}

func getCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "wired", "library.json"), nil
}
