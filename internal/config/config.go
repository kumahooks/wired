// Package config handles the program's settings and it's schema definitions
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var ErrInvalidMusicPath = errors.New("music path does not exist or is not a directory")

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

type Config struct {
	Title            string         `toml:"title"`
	MusicLibraryPath string         `toml:"music_library_path"`
	InputCharLimit   int            `toml:"input_char_limit"`
	Notification     Notification   `toml:"notification"`
	Colors           ColorPalette   `toml:"colors"`
	Keybinds         KeybindMapping `toml:"keybinds"`
}

type Notification struct {
	NotificationMaxWidth     int `toml:"notification_max_width"`
	NotificationMaxHeight    int `toml:"notification_max_height"`
	NotificationShownMax     int `toml:"notification_shown_max"`
	NotificationDurationSecs int `toml:"notification_duration_secs"`
}

type ColorPalette struct {
	Border              string `toml:"border"`
	TextInactive        string `toml:"text_inactive"`
	CursorForeground    string `toml:"cursor_fg"`
	NotificationInfo    string `toml:"notification_info"`
	NotificationError   string `toml:"notification_error"`
	NotificationSuccess string `toml:"notification_success"`
}

type KeybindMapping struct {
	MoveLeft []string `toml:"move_left"`
	MoveDown []string `toml:"move_down"`
	MoveUp   []string `toml:"move_up"`
	Select   []string `toml:"select"`
	Cancel   []string `toml:"cancel"`
	Quit     []string `toml:"quit"`
}

func (cfg *Config) Save() error {
	path, err := getPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, filePerm)
}

func (cfg *Config) SetAndSaveMusicLibraryPath(path string) error {
	expanded, err := cfg.IsMusicLibraryPathValid(path)
	if err != nil {
		return err
	}

	cfg.MusicLibraryPath = expanded

	return cfg.Save()
}

func (cfg *Config) IsMusicLibraryPathValid(path string) (string, error) {
	expanded := expandPath(path)

	info, err := os.Stat(expanded)
	if err != nil || !info.IsDir() {
		return "", ErrInvalidMusicPath
	}

	return expanded, nil
}

func Load() (*Config, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		if err = ensureExists(path); err != nil {
			return nil, err
		}
	}

	// TODO: there's a case where the file exists, but it's outdated, so it misses some configs
	// need to rewrite the file keeping the current config but adding the new ones

	// TODO: validate config file values, each value should have rules to validate

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// If music library is not correct, we clear it so the prompt shows up
	if cfg.MusicLibraryPath != "" {
		if _, err = cfg.IsMusicLibraryPathValid(cfg.MusicLibraryPath); err != nil {
			cfg.MusicLibraryPath = ""
		}
	}

	return &cfg, nil
}

func getPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "wired", "config.toml"), nil
}

func ensureExists(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(DefaultConfig), filePerm)
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if envHome := os.Getenv("HOME"); envHome != "" {
			home = envHome
		} else {
			return path
		}
	}

	return filepath.Join(home, path[2:])
}
