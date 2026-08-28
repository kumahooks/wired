package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnchorPositionsBottomRight(t *testing.T) {
	t.Parallel()

	content := "uwu"
	windowWidth := 80
	windowHeight := 24

	layer := Anchor(content, windowWidth, windowHeight)

	assert.Equal(t, content, layer.GetContent())
	assert.Equal(t, windowWidth-len(content), layer.GetX(), "X should pin the content to the right edge")
	assert.Equal(t, windowHeight-1, layer.GetY(), "Y should pin the content to the bottom edge")
	assert.Equal(t, zIndex, layer.GetZ())
}
