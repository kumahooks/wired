// Package config handles config loading, parsing, and defaults.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	configDirPermission  = 0o755
	configFilePermission = 0o644
)

// ThemeConfig holds the color palette as #RRGGBB hex strings, read directly from TOML.
// TODO: I loaded all my color palette from my webserver. Eventually we should delete the colors we won't be using :)
type ThemeConfig struct {
	Surface           string `toml:"surface"`
	SurfaceAlt        string `toml:"surface_alt"`
	BorderPanel       string `toml:"border_panel"`
	BorderHairline    string `toml:"border_hairline"`
	TextPrimary       string `toml:"text_primary"`
	TextStrong        string `toml:"text_strong"`
	TextMuted         string `toml:"text_muted"`
	TextDim           string `toml:"text_dim"`
	TextFaint         string `toml:"text_faint"`
	TextPlaceholder   string `toml:"text_placeholder"`
	AccentInteractive string `toml:"accent_interactive"`
	AccentDeep        string `toml:"accent_deep"`
	AccentBright      string `toml:"accent_bright"`
	AccentConfirm     string `toml:"accent_confirm"`
	AccentLink        string `toml:"accent_link"`
	AccentPrompt      string `toml:"accent_prompt"`
	AccentDanger      string `toml:"accent_danger"`
	AccentError       string `toml:"accent_error"`
	Track             string `toml:"track"`
}

type ActionsMapping struct {
	Playlist     []string `toml:"playlist"`
	LibraryStats []string `toml:"library_stats"`
}

type KeybindMapping struct {
	MoveLeft    []string       `toml:"move_left"`
	MoveRight   []string       `toml:"move_right"`
	Select      []string       `toml:"select"`
	Quit        []string       `toml:"quit"`
	GoBack      []string       `toml:"go_back"`
	OpenActions []string       `toml:"open_actions"`
	Actions     ActionsMapping `toml:"actions"`
}

type Config struct {
	Title          string         `toml:"title"`
	LibrariesPaths []string       `toml:"libraries_paths"`
	Theme          ThemeConfig    `toml:"theme"`
	Keybinds       KeybindMapping `toml:"keybinds"`
}

// Load reads the config file from disk, validates it, prunes invalid library paths, and persists the pruned form back.
// The returned bool is true when no config file existed and one was just created from defaults.
func Load() (*Config, bool, []string, error) {
	var isConfigDefaults bool = false

	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] resolve config path: %w", err)
	}

	_, err = os.Stat(configFilePath)
	if err != nil && os.IsNotExist(err) {
		if err = ensureConfigExists(configFilePath); err != nil {
			return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] ensure config exists: %w", err)
		}

		isConfigDefaults = true
	}

	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] read config file: %w", err)
	}

	configData := Defaults()
	if err = toml.Unmarshal(fileData, &configData); err != nil {
		return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] parse config file: %w", err)
	}

	if err = configData.validateConfigValues(); err != nil {
		return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] validate config: %w", err)
	}

	if err = configData.saveFile(); err != nil {
		return nil, isConfigDefaults, []string{}, fmt.Errorf("[config:Load] persist config: %w", err)
	}

	invalidLibraryPaths := configData.getAndClearInvalidLibraryPaths()

	return &configData, isConfigDefaults, invalidLibraryPaths, nil
}

func (config *Config) validateConfigValues() error {
	var errs []error = []error{}

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

	hexColorPattern := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hexColor := func(field string, value string) {
		if !hexColorPattern.MatchString(value) {
			errs = append(
				errs,
				fmt.Errorf("[config:validateConfigValues] %s must be a #RRGGBB hex color, got %q", field, value),
			)
		}
	}

	nonEmptyString("title", config.Title)

	nonEmptyArray("keybinds.move_left", config.Keybinds.MoveLeft)
	nonEmptyArray("keybinds.move_right", config.Keybinds.MoveRight)
	nonEmptyArray("keybinds.select", config.Keybinds.Select)
	nonEmptyArray("keybinds.quit", config.Keybinds.Quit)
	nonEmptyArray("keybinds.go_back", config.Keybinds.GoBack)
	nonEmptyArray("keybinds.open_actions", config.Keybinds.OpenActions)

	nonEmptyArray("keybinds.actions.playlist", config.Keybinds.Actions.Playlist)
	nonEmptyArray("keybinds.actions.library_stats", config.Keybinds.Actions.LibraryStats)

	hexColor("theme.surface", config.Theme.Surface)
	hexColor("theme.surface_alt", config.Theme.SurfaceAlt)
	hexColor("theme.border_panel", config.Theme.BorderPanel)
	hexColor("theme.border_hairline", config.Theme.BorderHairline)
	hexColor("theme.text_primary", config.Theme.TextPrimary)
	hexColor("theme.text_strong", config.Theme.TextStrong)
	hexColor("theme.text_muted", config.Theme.TextMuted)
	hexColor("theme.text_dim", config.Theme.TextDim)
	hexColor("theme.text_faint", config.Theme.TextFaint)
	hexColor("theme.text_placeholder", config.Theme.TextPlaceholder)
	hexColor("theme.accent_interactive", config.Theme.AccentInteractive)
	hexColor("theme.accent_deep", config.Theme.AccentDeep)
	hexColor("theme.accent_bright", config.Theme.AccentBright)
	hexColor("theme.accent_confirm", config.Theme.AccentConfirm)
	hexColor("theme.accent_link", config.Theme.AccentLink)
	hexColor("theme.accent_prompt", config.Theme.AccentPrompt)
	hexColor("theme.accent_danger", config.Theme.AccentDanger)
	hexColor("theme.accent_error", config.Theme.AccentError)
	hexColor("theme.track", config.Theme.Track)

	return errors.Join(errs...)
}

func (config *Config) getAndClearInvalidLibraryPaths() []string {
	var invalidPaths []string = []string{}
	var validPaths []string = []string{}

	for _, path := range config.LibrariesPaths {
		expandedPath := expandPath(path)

		info, err := os.Stat(expandedPath)
		if err == nil && info.IsDir() {
			validPaths = append(validPaths, expandedPath)
			continue
		}

		invalidPaths = append(invalidPaths, expandedPath)
	}

	config.LibrariesPaths = validPaths

	return invalidPaths
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
