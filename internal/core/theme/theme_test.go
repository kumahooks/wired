package theme

import (
	"image/color"
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wired/internal/core/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	configTheme := config.Defaults().Theme
	resolved := New(configTheme)

	themeValue := reflect.ValueOf(resolved)
	themeType := themeValue.Type()

	for index := 0; index < themeValue.NumField(); index++ {
		field := themeType.Field(index)
		fieldColor, ok := themeValue.Field(index).Interface().(color.Color)
		require.True(t, ok, "Theme.%s is not a color.Color", field.Name)
		require.NotNil(t, fieldColor, "Theme.%s = nil, want a non-nil color.Color", field.Name)

		configValue := reflect.ValueOf(configTheme).FieldByName(field.Name).String()
		wantColor := lipgloss.Color(configValue)

		gotR, gotG, gotB, gotA := fieldColor.RGBA()
		wantR, wantG, wantB, wantA := wantColor.RGBA()

		assert.Equal(t, wantR, gotR, "Theme.%s red mismatch", field.Name)
		assert.Equal(t, wantG, gotG, "Theme.%s green mismatch", field.Name)
		assert.Equal(t, wantB, gotB, "Theme.%s blue mismatch", field.Name)
		assert.Equal(t, wantA, gotA, "Theme.%s alpha mismatch", field.Name)
	}
}

func TestNewAllBlack(t *testing.T) {
	t.Parallel()

	configTheme := config.ThemeConfig{}
	configThemeValue := reflect.ValueOf(&configTheme).Elem()

	themeType := reflect.TypeOf(configTheme)
	for index := 0; index < themeType.NumField(); index++ {
		field := themeType.Field(index)
		configThemeValue.FieldByName(field.Name).SetString("#000000")
	}

	resolved := New(configTheme)

	themeValue := reflect.ValueOf(resolved)
	for index := 0; index < themeValue.NumField(); index++ {
		fieldColor, ok := themeValue.Field(index).Interface().(color.Color)
		require.True(t, ok, "field %d is not a color.Color", index)
		require.NotNil(t, fieldColor, "nil color at field %d", index)

		red, green, blue, _ := fieldColor.RGBA()
		assert.Zero(t, red, "field %d red = %d, want 0", index, red)
		assert.Zero(t, green, "field %d green = %d, want 0", index, green)
		assert.Zero(t, blue, "field %d blue = %d, want 0", index, blue)
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()

	defaultTheme := Default()
	fromDefaults := New(config.Defaults().Theme)

	assert.Equal(t, fromDefaults, defaultTheme)
}
