package whichkey

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestAnchorPositionsBottomLeft(t *testing.T) {
	t.Parallel()

	content := "uwu"
	windowWidth := 80
	windowHeight := 24

	layer := Anchor(content, windowWidth, windowHeight)

	assert.Equal(t, content, layer.GetContent())
	assert.Equal(t, 0, layer.GetX(), "X should pin the content to the left edge")
	assert.Equal(t, windowHeight-lipgloss.Height(content), layer.GetY(), "Y should pin the content to the bottom edge")
	assert.Equal(t, zIndex, layer.GetZ())
}

func TestAnchorAccountsForContentHeight(t *testing.T) {
	t.Parallel()

	content := "line one\nline two"
	windowHeight := 24

	layer := Anchor(content, 80, windowHeight)

	assert.Equal(t, windowHeight-2, layer.GetY(), "Y should offset by the content's line count")
}
