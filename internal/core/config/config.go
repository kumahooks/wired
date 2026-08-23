// Package config handles config loading, parsing, and defaults.
package config

import (
	"errors"
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

// Load reads the config file from disk, validates it, prunes invalid library paths, and persists the pruned form back.
func Load() (*Config, error) {
	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("[config:Load] resolve config path: %w", err)
	}

	_, err = os.Stat(configFilePath)
	if err != nil && os.IsNotExist(err) {
		if err = ensureConfigExists(configFilePath); err != nil {
			return nil, fmt.Errorf("[config:Load] ensure config exists: %w", err)
		}
	}

	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("[config:Load] read config file: %w", err)
	}

	configData := Defaults()
	if err = toml.Unmarshal(fileData, &configData); err != nil {
		return nil, fmt.Errorf("[config:Load] parse config file: %w", err)
	}

	if err = configData.validateConfigValues(); err != nil {
		return nil, fmt.Errorf("[config:Load] validate config: %w", err)
	}

	if len(configData.LibrariesPaths) > 0 {
		// TODO: in truth I do not like the idea of mutating the file here... we should improve this once we introduce
		// the initialization view.
		configData.clearInvalidLibraryPaths()
	}

	if err = configData.saveFile(); err != nil {
		return nil, fmt.Errorf("[config:Load] persist config: %w", err)
	}

	return &configData, nil
}

func (config *Config) validateConfigValues() error {
	var errs []error

	nonEmptyString := func(field string, value string) {
		if trimmed := strings.Trim(value, " "); trimmed == "" {
			errs = append(errs, fmt.Errorf("[config:validateConfigValues] %s must not contain empty strings", field))
		}
	}

	nonEmptyArray := func(field string, values []string) {
		if len(values) == 0 {
			errs = append(errs, fmt.Errorf("[config:validateConfigValues] %s must have at least one binding", field))
		}

		for _, value := range values {
			nonEmptyString(field, value)
		}
	}

	nonEmptyString("title", config.Title)

	nonEmptyArray("keybinds.quit", config.Keybinds.Quit)

	return errors.Join(errs...)
}

func (config *Config) clearInvalidLibraryPaths() {
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

func (config *Config) saveFile() error {
	configFilePath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("[config:saveFile] resolve config path: %w", err)
	}

	configData, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("[config:saveFile] marshal config: %w", err)
	}

	if err = os.WriteFile(configFilePath, configData, configFilePermission); err != nil {
		return fmt.Errorf("[config:saveFile] write config file: %w", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("[config:getConfigPath] user config dir: %w", err)
	}

	return filepath.Join(dir, "wire_d", "config.toml"), nil
}

func ensureConfigExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), configDirPermission); err != nil {
		return fmt.Errorf("[config:ensureConfigExists] create config dir: %w", err)
	}

	defaultsByteData, err := DefaultsTOML()
	if err != nil {
		return fmt.Errorf("[config:ensureConfigExists] marshal defaults: %w", err)
	}

	if err = os.WriteFile(path, defaultsByteData, configFilePermission); err != nil {
		return fmt.Errorf("[config:ensureConfigExists] write default config: %w", err)
	}

	return nil
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
