package ui

import (
	"charm.land/lipgloss/v2"

	"wired/internal/app/ui/components/dialog"
	"wired/internal/app/ui/components/notification"
	"wired/internal/app/ui/components/whichkey"
)

// composeOverlays stacks every visible overlay on top of the base view.
//
// TODO: a new Compositor flattens and z-sorts the layer tree on every call, and View calls this once per frame.
// this *could* be problematic... if wired starts getting performance issues later on, it will be necessary to revisit this.
// for reference, crush avoids this by drawing into a uv.ScreenBuffer.
func (model *UIModel) composeOverlays(base string) string {
	baseLayer := lipgloss.NewLayer(base)

	// whichkey is the "command palette" overlay. it basically maps all the currently available commands.
	if model.state != uiBootstrapping && model.state != uiInitializing && model.whichkeyModel.IsVisible() {
		baseLayer.AddLayers(whichkey.Anchor(
			model.whichkeyModel.Render(model.windowWidth, model.windowHeight),
			model.windowWidth,
			model.windowHeight,
		))
	}

	// notifications is a FIFO queue where UIModel can push notifications to and show to the user in an overlay.
	if model.notificationModel.HasActiveNotifications() {
		baseLayer.AddLayers(notification.Anchor(
			model.notificationModel.Render(model.windowWidth, model.windowHeight),
			model.windowWidth,
			model.windowHeight,
		))
	}

	// the confirm dialog is a modal question rendered on top of every other overlay.
	if model.dialogModel.IsOpen() {
		baseLayer.AddLayers(dialog.Anchor(
			model.dialogModel.Render(),
			model.windowWidth,
			model.windowHeight,
		))
	}

	return lipgloss.NewCompositor(baseLayer).Render()
}
