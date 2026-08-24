package config

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// configsEqual compares two Configs treating nil and empty []string slices as equal.
func configsEqual(want Config, got Config) bool {
	cloneConfig := func(config Config) Config {
		output := config

		normalizeEmptyToNil := func(values []string) []string {
			if len(values) == 0 {
				return nil
			}

			return values
		}

		output.LibrariesPaths = normalizeEmptyToNil(output.LibrariesPaths)
		output.Keybinds.MoveLeft = normalizeEmptyToNil(output.Keybinds.MoveLeft)
		output.Keybinds.MoveRight = normalizeEmptyToNil(output.Keybinds.MoveRight)
		output.Keybinds.Select = normalizeEmptyToNil(output.Keybinds.Select)
		output.Keybinds.Quit = normalizeEmptyToNil(output.Keybinds.Quit)

		return output
	}

	return reflect.DeepEqual(cloneConfig(want), cloneConfig(got))
}

func TestDefaults(t *testing.T) {
	t.Parallel()

	defaults := Defaults()

	assert.Equal(t, "wire_d", defaults.Title)
	assert.Empty(t, defaults.LibrariesPaths)

	themeValue := reflect.ValueOf(defaults.Theme)
	themeType := themeValue.Type()
	for index := 0; index < themeValue.NumField(); index++ {
		field := themeType.Field(index)
		value := themeValue.Field(index).String()

		require.NotEmpty(t, value, "Theme.%s = empty, want a color", field.Name)
		assert.True(
			t,
			hexColorPattern.MatchString(value),
			"Theme.%s = %q, want a #RRGGBB hex color",
			field.Name,
			value,
		)
	}

	keybindsValue := reflect.ValueOf(defaults.Keybinds)
	keybindsType := keybindsValue.Type()
	for index := 0; index < keybindsValue.NumField(); index++ {
		field := keybindsType.Field(index)
		bindings := keybindsValue.Field(index).Interface().([]string)

		require.NotEmpty(t, bindings, "Keybinds.%s = empty, want at least one binding", field.Name)

		for _, binding := range bindings {
			assert.NotEmpty(t, strings.TrimSpace(binding), "Keybinds.%s contains empty binding", field.Name)
		}
	}
}

func TestDefaultsTOMLRoundTrip(t *testing.T) {
	t.Parallel()

	data, err := DefaultsTOML()
	require.NoError(t, err)

	var got Config
	require.NoError(t, toml.Unmarshal(data, &got))

	assert.True(t, configsEqual(Defaults(), got))
}
