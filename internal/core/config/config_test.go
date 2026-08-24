package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde expands to home",
			path: "~/music",
			want: filepath.Join(home, "music"),
		},
		{
			name: "non-tilde returns as-is",
			path: "/var/lib/music",
			want: "/var/lib/music",
		},
		{
			name: "relative path returns as-is",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "tilde-only stays as separator check",
			path: "~/",
			want: home,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := expandPath(test.path)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpandPathEmptyHome(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	got := expandPath("~/music")
	assert.Equal(t, "~/music", got, "expandPath with empty home should return path unchanged")
}

func TestValidateConfigValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    Config
		wantError string
	}{
		{
			name:   "defaults pass",
			config: Defaults(),
		},
		{
			name: "empty title fails",
			config: func() Config {
				config := Defaults()
				config.Title = "  "
				return config
			}(),
			wantError: "title must not contain empty strings",
		},
		{
			name: "empty move_left fails",
			config: func() Config {
				config := Defaults()
				config.Keybinds.MoveLeft = []string{}
				return config
			}(),
			wantError: "keybinds.move_left must have at least one binding",
		},
		{
			name: "empty move_right fails",
			config: func() Config {
				config := Defaults()
				config.Keybinds.MoveRight = []string{}
				return config
			}(),
			wantError: "keybinds.move_right must have at least one binding",
		},
		{
			name: "empty select fails",
			config: func() Config {
				config := Defaults()
				config.Keybinds.Select = []string{}
				return config
			}(),
			wantError: "keybinds.select must have at least one binding",
		},
		{
			name: "empty quit fails",
			config: func() Config {
				config := Defaults()
				config.Keybinds.Quit = []string{}
				return config
			}(),
			wantError: "keybinds.quit must have at least one binding",
		},
		{
			name: "empty surface color fails",
			config: func() Config {
				config := Defaults()
				config.Theme.Surface = ""
				return config
			}(),
			wantError: "theme.surface must be a #RRGGBB hex color",
		},
		{
			name: "malformed accent_error fails",
			config: func() Config {
				config := Defaults()
				config.Theme.AccentError = "not-a-color"
				return config
			}(),
			wantError: "theme.accent_error must be a #RRGGBB hex color",
		},
		{
			name: "malformed text_strong fails",
			config: func() Config {
				config := Defaults()
				config.Theme.TextStrong = "not-a-color"
				return config
			}(),
			wantError: "theme.text_strong must be a #RRGGBB hex color",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.config.validateConfigValues()
			switch {
			case test.wantError == "":
				require.NoError(t, err)
			case test.wantError != "":
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantError)
			}
		})
	}
}

func TestValidateConfigValuesMultipleErrorsJoined(t *testing.T) {
	t.Parallel()

	config := Defaults()
	config.Title = ""
	config.Theme.Surface = "bad"
	config.Keybinds.Quit = []string{}

	err := config.validateConfigValues()
	require.Error(t, err)

	joined := err.Error()
	for _, want := range []string{
		"title must not contain empty strings",
		"theme.surface must be a #RRGGBB hex color",
		"keybinds.quit must have at least one binding",
	} {
		assert.Contains(t, joined, want)
	}
}

func TestClearInvalidLibraryPaths(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	nestedFile := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(nestedFile, []byte("file"), 0o644))

	config := Config{
		LibrariesPaths: []string{
			realDir,
			"/this/path/does/not/exist",
			nestedFile,
		},
	}

	config.clearInvalidLibraryPaths()

	if assert.Len(t, config.LibrariesPaths, 1) {
		assert.Equal(t, realDir, config.LibrariesPaths[0])
	}
}

func TestEnsureConfigExists(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "wire_d", "config.toml")

	require.NoError(t, ensureConfigExists(configPath))

	dirInfo, err := os.Stat(filepath.Dir(configPath))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(configDirPermission), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(configFilePermission), fileInfo.Mode().Perm())

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var got Config
	require.NoError(t, toml.Unmarshal(data, &got))
	assert.True(t, configsEqual(Defaults(), got))
}

func TestSaveFile(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	// saveFile does not mkdir, it's Load that does that via ensureConfigExists.
	require.NoError(t, os.MkdirAll(filepath.Join(configRoot, "wire_d"), configDirPermission))

	config := Defaults()
	config.LibrariesPaths = []string{"/tmp/library"}

	require.NoError(t, config.saveFile())

	configPath, err := getConfigPath()
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var got Config
	require.NoError(t, toml.Unmarshal(data, &got))

	// LibrariesPaths may come back nil after the round-trip.
	assert.True(t, configsEqual(config, got))
}

func TestGetConfigPath(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	got, err := getConfigPath()
	require.NoError(t, err)

	wantSuffix := filepath.Join("wire_d", "config.toml")
	assert.True(t, strings.HasSuffix(got, wantSuffix), "getConfigPath = %q, want suffix %q", got, wantSuffix)
}
