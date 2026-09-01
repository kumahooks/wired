package audio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const cacheFilePermission = 0o644

// WriteCache overwrites the cache file with the given fileMap.
func WriteCache(fileMap map[string]*AudioFile) error {
	cacheFilePath, err := getCachePath()
	if err != nil {
		return err
	}

	cacheData, err := json.MarshalIndent(fileMap, "", "\t")
	if err != nil {
		return fmt.Errorf("[audio:WriteCache] marshal cache: %w", err)
	}

	// The cache file lives alongside the user's config, which config.Load already created.
	if err = os.WriteFile(cacheFilePath, cacheData, cacheFilePermission); err != nil {
		return fmt.Errorf("[audio:WriteCache] write cache file: %w", err)
	}

	return nil
}

// LoadCache reads the cache file and returns the unmarshalled fileMap. A missing cache file returns a nil map and no error.
func LoadCache() (map[string]*AudioFile, error) {
	cacheFilePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	cacheData, err := os.ReadFile(cacheFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("[audio:LoadCache] read cache file: %w", err)
	}

	var cacheResult map[string]*AudioFile
	if err = json.Unmarshal(cacheData, &cacheResult); err != nil {
		return nil, fmt.Errorf("[audio:LoadCache] unmarshal cache file: %w", err)
	}

	// JSON null entries unmarshal as nil pointers, so we drop them here so consumers never see a nil *AudioFile.
	for filePath, audioFile := range cacheResult {
		if audioFile == nil {
			delete(cacheResult, filePath)
		}
	}

	return cacheResult, nil
}

func getCachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("[audio:getCachePath] user config dir: %w", err)
	}

	return filepath.Join(dir, "wire_d", "cache.json"), nil
}
