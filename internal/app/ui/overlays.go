package ui

import (
	"charm.land/lipgloss/v2"

	"wired/internal/app/ui/components/notification"
	"wired/internal/app/ui/components/whichkey"
)

// composeOverlays stacks every visible overlay on top of the base view.
//
// TODO: a new Compositor flattens and z-sorts the layer tree on every call, and View calls this once per frame.
// this *could* be problematic... if we have performance issues later we should definitely revisit this.
// for reference, crush avoids this by drawing into a uv.ScreenBuffer.
func (model *UIModel) composeOverlays(base string) string {
	baseLayer := lipgloss.NewLayer(base)

	if model.state == uiIdle && model.whichkeyModel.IsVisible() {
		baseLayer.AddLayers(whichkey.Anchor(
			model.whichkeyModel.Render(model.windowWidth, model.windowHeight),
			model.windowWidth,
			model.windowHeight,
		))
	}

	if model.notificationModel.HasActiveNotifications() {
		baseLayer.AddLayers(notification.Anchor(
			model.notificationModel.Render(model.windowWidth, model.windowHeight),
			model.windowWidth,
			model.windowHeight,
		))
	}

	return lipgloss.NewCompositor(baseLayer).Render()
}
