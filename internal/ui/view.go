package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"wired/internal/ui/notification"
)

func (model Model) View() string {
	contentHeight := max(model.height-1, 0)

	var base string

	if model.Dialog.Visible() {
		base = model.Dialog.View()
	} else if model.Config == nil {
		base = ""
	} else if model.Modal.Visible() {
		base = model.Modal.View()
	} else {
		// TODO: player view
		base = model.debugLibraryView()
	}

	// Pad base to exactly contentHeight lines
	baseLines := strings.Split(base, "\n")
	for len(baseLines) < contentHeight {
		baseLines = append(baseLines, "")
	}
	if len(baseLines) > contentHeight {
		baseLines = baseLines[:contentHeight]
	}

	base = strings.Join(baseLines, "\n")

	if model.Config != nil {
		visibleNotifications := model.Notifications.Visible(model.Config.Notification.NotificationShownMax)

		if len(visibleNotifications) > 0 {
			notifications := model.renderNotifications(visibleNotifications)
			base = overlayBottomRight(base, notifications, model.width, contentHeight)
		}
	}

	return base + "\n" + model.Footer.View()
}

// TODO: dont forget to delete this :P
// or maybe not? do we want a statistics page?
func (model Model) debugLibraryView() string {
	if model.Library == nil || len(model.Library.Songs) == 0 {
		scanKeys := strings.Join(model.Config.Keybinds.ScanFiles, ", ")
		return fmt.Sprintf("Library is empty. Press [%s] to scan.", scanKeys)
	}

	var b strings.Builder

	albumCount := 0
	for _, artist := range model.Library.Artists {
		albumCount += len(artist.Albums)
	}

	type formatStats struct {
		total      int
		incomplete int
	}

	formats := map[string]*formatStats{}
	totalIncomplete := 0

	for _, song := range model.Library.Songs {
		ext := ""
		if dot := strings.LastIndex(song.FileName, "."); dot >= 0 {
			ext = strings.ToLower(song.FileName[dot:])
		}

		stats, ok := formats[ext]
		if !ok {
			stats = &formatStats{}
			formats[ext] = stats
		}

		stats.total++

		if song.Metadata.ArtistName == "Unknown Artist" || song.Metadata.AlbumName == "Unknown Album" {
			stats.incomplete++
			totalIncomplete++
		}
	}

	totalValid := len(model.Library.Songs) - totalIncomplete

	fmt.Fprintf(&b, "Library Statistics\n\n")
	fmt.Fprintf(&b, "  Songs:   %d\n", len(model.Library.Songs))
	fmt.Fprintf(&b, "  Artists: %d\n", len(model.Library.Artists))
	fmt.Fprintf(&b, "  Albums:  %d\n", albumCount)
	fmt.Fprintf(&b, "\n  Metadata:\n")
	fmt.Fprintf(&b, "    Valid:      %d\n", totalValid)
	fmt.Fprintf(&b, "    Incomplete: %d\n", totalIncomplete)
	fmt.Fprintf(&b, "\n  Formats:\n")

	exts := make([]string, 0, len(formats))
	for ext := range formats {
		exts = append(exts, ext)
	}

	sort.Slice(exts, func(i, j int) bool {
		return formats[exts[i]].total > formats[exts[j]].total
	})

	for _, ext := range exts {
		stats := formats[ext]
		valid := stats.total - stats.incomplete
		fmt.Fprintf(&b, "    %-6s %d\n", ext, stats.total)
		fmt.Fprintf(&b, "      valid: %d / incomplete: %d\n", valid, stats.incomplete)
	}

	// TODO: remove this
	fmt.Fprintf(&b, "\n  Incomplete .flac files:\n")
	for filePath, song := range model.Library.Songs {
		if strings.HasSuffix(strings.ToLower(song.FileName), ".flac") &&
			(song.Metadata.ArtistName == "Unknown Artist" || song.Metadata.AlbumName == "Unknown Album") {
			fmt.Fprintf(&b, "    %s\n", filePath)
		}
	}

	scanKeys := strings.Join(model.Config.Keybinds.ScanFiles, ", ")
	fmt.Fprintf(&b, "\n  Press [%s] to rescan.", scanKeys)

	return b.String()
}

func (model Model) renderNotifications(notifications []notification.Notification) string {
	bubbles := make([]string, 0, len(notifications))

	for _, n := range notifications {
		bubble := model.Notifications.Render(n)
		bubbles = append(bubbles, bubble)
	}

	return lipgloss.JoinVertical(lipgloss.Right, bubbles...)
}

func overlayBottomRight(base string, overlay string, width int, height int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	if len(baseLines) > height {
		baseLines = baseLines[:height]
	}

	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := len(overlayLines)
	if overlayHeight > height {
		overlayLines = overlayLines[overlayHeight-height:]
		overlayHeight = height
	}

	startRow := height - overlayHeight
	startCol := max(width-overlayWidth, 0)

	// Merge overlay onto base, preserving base content to the left
	for i, overlayLine := range overlayLines {
		rowIdx := startRow + i
		if rowIdx < 0 || rowIdx >= height {
			continue
		}

		baseLine := baseLines[rowIdx]
		baseVisualWidth := lipgloss.Width(baseLine)

		if baseVisualWidth <= startCol {
			// Base line is shorter than where overlay starts, just pad and append
			padding := strings.Repeat(" ", startCol-baseVisualWidth)
			baseLines[rowIdx] = baseLine + padding + overlayLine
		} else {
			truncated := ansi.Truncate(baseLine, startCol, "")

			truncatedWidth := lipgloss.Width(truncated)
			paddingLength := max(startCol-truncatedWidth, 0)
			padding := strings.Repeat(" ", paddingLength)

			baseLines[rowIdx] = truncated + padding + overlayLine
		}
	}

	return strings.Join(baseLines, "\n")
}
