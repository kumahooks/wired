package dialog

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestAnchorCentersContent(t *testing.T) {
	t.Parallel()

	content := "uwu"
	windowWidth := 80
	windowHeight := 24
	contentWidth := lipgloss.Width(content)

	layer := Anchor(content, windowWidth, windowHeight)

	assert.Equal(t, content, layer.GetContent())
	assert.Equal(t, (windowWidth-contentWidth)/2, layer.GetX(), "X should center the content horizontally")
	assert.Equal(t, (windowHeight-lipgloss.Height(content))/2, layer.GetY(), "Y should center the content vertically")
	assert.Equal(t, zIndex, layer.GetZ())
}

func TestAnchorClampsNegativeOffsets(t *testing.T) {
	t.Parallel()

	// The sample content is 46 columns wide and a single line tall.
	content := "a card far wider and taller than this window"
	contentWidth := lipgloss.Width(content)

	tests := []struct {
		name         string
		windowWidth  int
		windowHeight int
		wantX        int
		wantY        int
	}{
		{
			name:         "window narrower than the content clamps X only",
			windowWidth:  4,
			windowHeight: 24,
			wantX:        0,
			wantY:        (24 - lipgloss.Height(content)) / 2,
		},
		{
			name:         "window as wide as the content centers X",
			windowWidth:  contentWidth,
			windowHeight: 24,
			wantX:        0,
			wantY:        (24 - lipgloss.Height(content)) / 2,
		},
		{
			name:         "window shorter than the content clamps Y only",
			windowWidth:  80,
			windowHeight: 0,
			wantX:        (80 - contentWidth) / 2,
			wantY:        0,
		},
		{
			name:         "window smaller than the content on both axes clamps both",
			windowWidth:  4,
			windowHeight: 2,
			wantX:        0,
			wantY:        0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			layer := Anchor(content, test.windowWidth, test.windowHeight)

			assert.Equal(t, test.wantX, layer.GetX(), "X offset mismatch")
			assert.Equal(t, test.wantY, layer.GetY(), "Y offset mismatch")
		})
	}
}
