// Package config handles config loading, parsing, and defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	configDirPermission  = 0o755
	configFilePermission = 0o644
)

type KeybindMapping struct {
	Quit []string `toml:"quit"`
}

type Config struct {
	Title          string         `toml:"title"`
	LibrariesPaths []string       `toml:"libraries_paths"`
	Keybinds       KeybindMapping `toml:"keybinds"`
}

func Load() (*Config, []error) {
	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, []error{err}
	}

	_, err = os.Stat(configFilePath)
	if err != nil && os.IsNotExist(err) {
		if err = ensureConfigExists(configFilePath); err != nil {
			return nil, []error{err}
		}
	}

	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, []error{err}
	}

	configData := Defaults()
	if err = toml.Unmarshal(fileData, &configData); err != nil {
		return nil, []error{err}
	}

	if errs := configData.ValidateConfigValues(); errs != nil {
		return nil, errs
	}

	if len(configData.LibrariesPaths) > 0 {
		// TODO: in truth I do not like the idea of mutating the file here... we should improve this once we introduce
		// the initialization view.
		configData.ClearInvalidLibraryPaths()
	}

	if err = configData.SaveFile(); err != nil {
		return nil, []error{err}
	}

	return &configData, nil
}

func (config *Config) ValidateConfigValues() []error {
	var errs []error

	nonEmptyString := func(field string, value string) {
		if trimmed := strings.Trim(value, " "); trimmed == "" {
			errs = append(errs, fmt.Errorf("%s must not contain empty strings", field))
		}
	}

	nonEmptyArray := func(field string, values []string) {
		if len(values) == 0 {
			errs = append(errs, fmt.Errorf("%s must have at least one binding", field))
		}

		for _, value := range values {
			nonEmptyString(field, value)
		}
	}

	nonEmptyString("title", config.Title)

	nonEmptyArray("keybinds.quit", config.Keybinds.Quit)

	return errs
}

func (config *Config) ClearInvalidLibraryPaths() {
	var validPaths []string

	for _, path := range config.LibrariesPaths {
		expandedPath := expandPath(path)

		info, err := os.Stat(expandedPath)
		if err != nil || !info.IsDir() {
			continue
		}

		validPaths = append(validPaths, expandedPath)
	}

	config.LibrariesPaths = validPaths
}

func (config *Config) SaveFile() error {
	configFilePath, err := getConfigPath()
	if err != nil {
		return err
	}

	configData, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, configData, configFilePermission)
}

func getConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "wire_d", "config.toml"), nil
}

func ensureConfigExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), configDirPermission); err != nil {
		return err
	}

	defaultsByteData, err := DefaultsTOML()
	if err != nil {
		return err
	}

	return os.WriteFile(path, defaultsByteData, configFilePermission)
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	homePath, err := os.UserHomeDir()
	if err != nil || homePath == "" {
		if homeEnvironment := os.Getenv("HOME"); homeEnvironment != "" {
			homePath = homeEnvironment
		} else {
			return path
		}
	}

	return filepath.Join(homePath, path[2:])
}
