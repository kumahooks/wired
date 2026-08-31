package librarystats

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"wired/internal/core/audio"
)

// Render draws the full-screen library stats view with the card grid centered in it.
func (model *Model) Render(windowWidth int, windowHeight int) string {
	// TODO: we will need to save this isntead of recomputing everytime. this can then be updated when we change the library.
	stats := model.computeStats()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		model.renderHeader(),
		model.renderTopCard(stats),
		model.renderBottomCards(stats),
		model.renderButton(),
		model.renderScanStatus(),
	)

	return lipgloss.Place(windowWidth, windowHeight, lipgloss.Center, lipgloss.Center, content)
}

// renderButton draws the focused rescan button.
func (model *Model) renderButton() string {
	return model.style.buttonFocused.Render(rescanButtonLabel)
}

// renderScanStatus draws the live scan progress line below the button, only while a scan is running.
func (model *Model) renderScanStatus() string {
	if !model.isScanning {
		return ""
	}

	return model.style.scanStatus.Render(fmt.Sprintf(scanStatusText, model.scannedFilesCount))
}

// computeStats computes the current library stats.
func (model *Model) computeStats() audio.Stats {
	if model.library == nil {
		return audio.Stats{FormatCounts: map[string]int{}}
	}

	return audio.Stats{FormatCounts: map[string]int{}}
}

// renderHeader draws the screen title, separator, and subtitle.
func (model *Model) renderHeader() string {
	headerTitle := model.style.header.Render(headerTitle)
	headerSeparator := model.style.headerSeparator.Render(headerSeparator)
	headerSubtitle := model.style.headerSubtitle.Render(headerSubtitle)

	return headerTitle + headerSeparator + headerSubtitle
}

// renderTopCard draws the "library size" stats card.
func (model *Model) renderTopCard(stats audio.Stats) string {
	return model.renderCard(
		model.drawLibrarySizeContent(stats),
		librarySizeCardTitle,
		librarySizeCardHeight,
		bigCardWidth,
	)
}

// drawLibrarySizeContent draws the files count, total size, and avg/track value.
func (model *Model) drawLibrarySizeContent(stats audio.Stats) string {
	averageBytes := int64(0)
	if stats.FilesCount > 0 {
		averageBytes = stats.TotalBytes / int64(stats.FilesCount)
	}

	filesCountString := fmt.Sprintf("%d", stats.FilesCount)
	averageBytesString := audio.GetReadableByteSize(averageBytes)
	totalBytesString := audio.GetReadableByteSize(stats.TotalBytes)

	rows := []string{
		model.renderStatRow("files", model.formatValue(filesCountString)),
		model.renderStatRow("total", model.formatValue(totalBytesString)),
		model.renderStatRow("avg/track", model.formatValue(averageBytesString)),
	}

	return strings.Join(rows, "\n")
}

// formatValue renders a row value.
func (model *Model) formatValue(value string) string {
	if value == "0" {
		return model.style.dash.Render(dashPlaceholder)
	}

	return model.style.value.Render(value)
}

// renderStatRow draws a "label value" pair.
func (model *Model) renderStatRow(rowLabel string, rowValue string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		model.style.label.Render(fmt.Sprintf("%-*s", labelWidth, rowLabel)), // we align every label up to labelWidth.
		model.style.value.Render(rowValue),
	)
}

// renderBottomCards joins the "files by format" and "library paths" cards horizontally.
func (model *Model) renderBottomCards(stats audio.Stats) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		model.renderCard(model.drawFilesByFormatContent(stats), formatsCardTitle, smallCardHeight, smallCardWidth),
		model.renderCard(model.drawLibraryPathsContent(), pathsCardTitle, smallCardHeight, smallCardWidth),
	)
}

// renderCard wraps drawn content in the bordered box with the title set as the first line.
func (model *Model) renderCard(content string, title string, height int, width int) string {
	innerWidth := width - borderWidth
	content = lipgloss.NewStyle().
		Width(innerWidth).
		Height(height).
		AlignVertical(lipgloss.Top).
		Render(content)

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		model.style.cardTitle.Render(title),
		content,
	)

	return model.style.card.Width(width).Render(inner)
}

// drawFilesByFormatContent draws one share bar per format, capped at maxVisibleFormatRows.
func (model *Model) drawFilesByFormatContent(stats audio.Stats) string {
	if stats.FilesCount == 0 {
		return model.style.dash.Render(dashPlaceholder)
	}

	formatBars := model.formatBars(stats)
	visibleCount := min(len(formatBars), maxVisibleFormatRows)

	rows := make([]string, 0, visibleCount)
	for _, bar := range formatBars[:visibleCount] {
		rows = append(rows, model.renderFormatBar(bar))
	}

	if len(formatBars) > visibleCount {
		rows = append(
			rows,
			model.style.label.Render(fmt.Sprintf("...and %d more formats~", len(formatBars)-visibleCount)),
		)
	}

	return strings.Join(rows, "\n")
}

// formatBars derives the per-format share bars from the stats, ordered by count descending.
func (model *Model) formatBars(stats audio.Stats) []formatBar {
	if stats.FilesCount == 0 {
		return nil
	}

	formatBars := make([]formatBar, 0, len(stats.FormatCounts))
	for format, count := range stats.FormatCounts {
		if format == "" {
			format = unknownFormatText
		}

		formatBars = append(formatBars, formatBar{
			format:   format,
			count:    count,
			fraction: float64(count) / float64(stats.FilesCount),
		})
	}

	// Sort by count descending, then by format name.
	slices.SortFunc(formatBars, func(left formatBar, right formatBar) int {
		if left.count != right.count {
			return right.count - left.count
		}

		return strings.Compare(left.format, right.format)
	})

	return formatBars
}

// renderFormatBar draws a single format row with a name, count, and a share bar.
func (model *Model) renderFormatBar(bar formatBar) string {
	name := model.style.label.Render(fmt.Sprintf("%-6s", bar.format))
	count := model.style.value.Render(fmt.Sprintf("%4d", bar.count))
	share := model.renderShareBar(bar.fraction, barWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, name, share, count)
}

// renderShareBar builds a filled/empty glyph bar of the given width.
func (model *Model) renderShareBar(fraction float64, width int) string {
	var shareBarLine strings.Builder
	filledTotal := int(min(max(fraction, 0), 1) * float64(width))

	shareBarLine.WriteString(model.style.formatShareBar.Render(strings.Repeat("█", filledTotal)))
	if width-filledTotal > 0 {
		shareBarLine.WriteString(model.style.formatShareEmptyBar.Render(strings.Repeat("░", width-filledTotal)))
	}

	return shareBarLine.String()
}

// drawLibraryPathsContent draws one indexed row per configured library path, capped at maxVisiblePathRows.
func (model *Model) drawLibraryPathsContent() string {
	if len(model.libraryPaths) == 0 {
		return model.style.dash.Render(noPathsText)
	}

	visibleCount := min(len(model.libraryPaths), maxVisiblePathRows)
	libraryPathRows := make([]string, 0, visibleCount)

	for index, path := range model.libraryPaths[:visibleCount] {
		libraryPathRows = append(libraryPathRows, model.renderLibraryPathRow(index, path))
	}

	remainder := len(model.libraryPaths) - visibleCount
	if remainder > 0 {
		libraryPathRows = append(
			libraryPathRows,
			model.style.label.Render(fmt.Sprintf("...and %d more", remainder)),
		)
	}

	return strings.Join(libraryPathRows, "\n")
}

// renderLibraryPathRow draws an index prefixed ("{index} {libraryPath}") row, truncating long paths.
func (model *Model) renderLibraryPathRow(index int, libraryPath string) string {
	innerWidth := smallCardWidth - borderWidth
	indexLabel := model.style.libraryPathIndex.Render(fmt.Sprintf("%02d ", index))
	availableInnerSpace := max(innerWidth-lipgloss.Width(indexLabel), 0)

	if lipgloss.Width(libraryPath) > availableInnerSpace && availableInnerSpace > 3 {
		libraryPath = fmt.Sprintf("%s...", ansi.Truncate(libraryPath, availableInnerSpace-3, ""))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, indexLabel, model.style.libraryPath.Render(libraryPath))
}
