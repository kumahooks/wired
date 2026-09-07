package librarystats

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"wired/internal/core/audio"
)

// screenCards is the fixed card set of the stats screen.
var screenCards = renderedCards{
	librarySize: card{
		title:      sizeCardTitle,
		draw:       (*Model).drawLibrarySizeContent,
		fixedLines: librarySizeCardLines,
		fixedWidth: librarySizeCardWidth,
	},
	libraryPaths: card{
		title:      pathsCardTitle,
		draw:       (*Model).drawLibraryPathsContent,
		fixedLines: pathsCardLines,
		fixedWidth: librarySizeCardWidth,
	},
	metadataHealth: card{
		title:      metadataCardTitle,
		draw:       (*Model).drawMetadataContent,
		fixedLines: metadataCardLines,
		fixedWidth: metadataFormatGroupWidth,
	},
	filesByFormat: card{
		title:      formatCardTitle,
		draw:       (*Model).drawFilesByFormatContent,
		fixedLines: formatCardLines,
		fixedWidth: metadataFormatGroupWidth,
	},
	topArtists: card{
		title:      topArtistsCardTitle,
		draw:       (*Model).drawTopArtistsContent,
		fixedLines: topArtistsCardLines,
		fixedWidth: topArtistsCardWidth,
	},
	placeholder: card{
		title:      placeholderCardTitle,
		draw:       (*Model).drawPlaceholderContent,
		fixedLines: 1,
		fixedWidth: topArtistsCardWidth,
	},
	trackLengths: card{
		title:      trackLengthsCardTitle,
		draw:       (*Model).drawTrackLengthsContent,
		fixedLines: lengthsGroupLines,
		fixedWidth: lengthsGroupWidth,
	},
	albumLengths: card{
		title:      albumLengthsCardTitle,
		draw:       (*Model).drawAlbumLengthsContent,
		fixedLines: lengthsGroupLines,
		fixedWidth: lengthsGroupWidth,
	},
}

// Render draws the library stats view centered in the terminal.
func (model *Model) Render(windowWidth int, windowHeight int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		model.renderHeader(),
		model.renderGrid(max(windowWidth, 0), max(windowHeight, 0)),
		model.renderDiscoveryStatusLines(max(windowWidth, 0)),
	)

	placed := lipgloss.Place(windowWidth, windowHeight, lipgloss.Center, lipgloss.Center, content)
	return lipgloss.NewStyle().MaxWidth(windowWidth).MaxHeight(windowHeight).Render(placed)
}

// renderHeader draws the screen title, separator, and subtitle.
func (model *Model) renderHeader() string {
	headerTitle := model.style.header.Render(headerTitle)
	headerSeparator := model.style.headerSeparator.Render(headerSeparator)
	headerSubtitle := model.style.headerSubtitle.Render(headerSubtitle)

	return headerTitle + headerSeparator + headerSubtitle
}

// renderGrid composes the card groups, and the button line, dropping groups the window cannot fit.
// - On width: below fullGridWidth, the "Top Artists" column (and its placeholder) is dropped.
// - On height: below fullGridLines, the lengths row ("Track Lengths"+"Album Lengths") is dropped.
//
// When even the compacted grid cannot fit, a message about the window size is rendered instead.
func (model *Model) renderGrid(windowWidth int, windowHeight int) string {
	if windowWidth < compactGridWidth || windowHeight < compactGridLines {
		wrappedMessage := ansi.Wordwrap(smallWindowText, max(windowWidth, 0), " ")

		return model.style.muted.
			Width(max(windowWidth, 0)).
			Align(lipgloss.Center).
			Render(wrappedMessage)
	}

	cards := screenCards

	libraryColumnGroup := lipgloss.JoinVertical(
		lipgloss.Left,
		model.renderCard(cards.librarySize),
		model.renderCard(cards.libraryPaths),
	)

	metadataFormatColumnGroup := lipgloss.JoinVertical(
		lipgloss.Left,
		model.renderCard(cards.metadataHealth),
		model.renderCard(cards.filesByFormat),
	)

	rows := make([]string, 0, 3)

	if windowWidth < fullGridWidth {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, libraryColumnGroup, metadataFormatColumnGroup))
	} else {
		topArtistsColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			model.renderCard(cards.topArtists),
			model.renderCard(cards.placeholder),
		)

		rows = append(
			rows,
			lipgloss.JoinHorizontal(lipgloss.Top, libraryColumnGroup, metadataFormatColumnGroup, topArtistsColumn),
		)
	}

	// on width clamping the lengths row vanishes along with the top artists column.
	if windowWidth >= lengthsRowWidth && windowHeight >= fullGridLines {
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.renderCard(cards.trackLengths),
			model.renderCard(cards.albumLengths),
		))
	}

	rows = append(rows, lipgloss.NewStyle().PaddingLeft(buttonRowLeftPadding).Render(model.renderButtonsRow()))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderCard wraps drawn content in the bordered box with the title set as the first line.
func (model *Model) renderCard(gridCard card) string {
	layout := cardLayout{
		innerWidth:  gridCard.fixedWidth - borderWidth,
		innerHeight: gridCard.fixedLines,
	}

	content := lipgloss.NewStyle().
		Width(layout.innerWidth).
		Height(layout.innerHeight).
		AlignVertical(lipgloss.Top).
		Render(gridCard.draw(model, layout))

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		model.style.cardTitle.Render(gridCard.title),
		content,
	)

	return model.style.card.Width(gridCard.fixedWidth).Render(inner)
}

// renderButtonsRow draws the action buttons.
func (model *Model) renderButtonsRow() string {
	buttonsRendering := make([]string, 0, len(model.buttons)*2)

	for position, button := range model.buttons {
		if position > 0 {
			buttonsRendering = append(buttonsRendering, buttonSpacing)
		}

		if position == model.cursorPosition {
			buttonsRendering = append(buttonsRendering, model.style.buttonFocused.Render(button.label))
		} else {
			buttonsRendering = append(buttonsRendering, model.style.buttonBlurred.Render(button.label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, buttonsRendering...)
}

// renderDiscoveryStatusLines draws the discovery progress lines, centered horizontally in the window.
func (model *Model) renderDiscoveryStatusLines(windowWidth int) string {
	lines := make([]string, discoveryStatusLines)

	switch {
	case !model.isDiscovering:
	case !model.isDiscoveryDone:
		lines[0] = fmt.Sprintf(scanStatusText, model.discoveredFilesCount)
	default:
		lines[0] = fmt.Sprintf(scanFoundText, model.discoveredFilesCount)
		lines[1] = fmt.Sprintf(scanParsingText, model.parsedMetatagCount, model.discoveredFilesCount)
	}

	centeredStyle := model.style.discoveryStatus.Width(max(windowWidth, 0)).Align(lipgloss.Center)
	renderedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		renderedLines = append(renderedLines, centeredStyle.Render(line))
	}

	return strings.Join(renderedLines, "\n")
}

// renderLabelValueRow draws a "{label} {value}" pair with the value right-aligned to the card's inner border.
func (model *Model) renderLabelValueRow(label string, value string, innerWidth int) string {
	valueColumn := model.style.rowValue.Width(max(innerWidth-labelWidth, 0)).Align(lipgloss.Right)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		model.style.rowLabel.Render(fmt.Sprintf("%-*s", labelWidth, label)),
		valueColumn.Render(model.renderStatValue(value)),
	)
}

// renderStatValue styles a formatted value with faint dashes when absent (empty string), and a bright style otherwise.
func (model *Model) renderStatValue(value string) string {
	if value == "" {
		return model.style.empty.Render(emptyPlaceholder)
	}

	return model.style.rowValue.Render(value)
}

// drawLibrarySizeContent draws the "Library Size" card's stat rows.
func (model *Model) drawLibrarySizeContent(layout cardLayout) string {
	stats := model.libraryStats
	averageBytes := int64(0)
	if stats.FilesCount > 0 {
		averageBytes = stats.TotalBytes / int64(stats.FilesCount)
	}

	rows := []string{
		model.renderLabelValueRow(sizeCardFilesTotalLabel, formatCountValue(stats.FilesCount), layout.innerWidth),
		model.renderLabelValueRow(sizeCardBytesTotalLabel, formatBytesValue(stats.TotalBytes), layout.innerWidth),
		model.renderLabelValueRow(sizeCardAvgBytesLabel, formatBytesValue(averageBytes), layout.innerWidth),
		model.renderLabelValueRow(sizeCardHeaviestLabel, formatBytesValue(stats.BiggestFileBytes), layout.innerWidth),
	}

	return strings.Join(rows, "\n")
}

// drawLibraryPathsContent draws one indexed row per configured library path, up to pathsCardVisibleRows.
func (model *Model) drawLibraryPathsContent(layout cardLayout) string {
	if len(model.libraryPaths) == 0 {
		return model.style.empty.Render(pathsCardEmptyText)
	}

	visibleCount := min(len(model.libraryPaths), pathsCardVisibleRows)
	libraryPathRows := make([]string, 0, visibleCount+1)

	for index, path := range model.libraryPaths[:visibleCount] {
		libraryPathRows = append(libraryPathRows, model.renderLibraryPathRow(index, path, layout.innerWidth))
	}

	remainder := len(model.libraryPaths) - visibleCount
	if remainder > 0 {
		libraryPathRows = append(
			libraryPathRows,
			model.style.muted.Render(fmt.Sprintf(pathsCardHintText, remainder)),
		)
	}

	return strings.Join(libraryPathRows, "\n")
}

// renderLibraryPathRow draws an index prefixed ("{index} {libraryPath}") row, truncating long paths.
func (model *Model) renderLibraryPathRow(index int, libraryPath string, innerWidth int) string {
	indexLabel := model.style.libraryPathIndex.Render(fmt.Sprintf("%02d ", index))
	availableInnerSpace := max(innerWidth-lipgloss.Width(indexLabel), 0)

	if lipgloss.Width(libraryPath) > availableInnerSpace && availableInnerSpace > 3 {
		libraryPath = fmt.Sprintf("%s...", ansi.Truncate(libraryPath, availableInnerSpace-3, ""))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, indexLabel, model.style.libraryPath.Render(libraryPath))
}

// drawMetadataContent draws rows regarding the metadata health of the files (missing titles, artist, album, or duplicates).
func (model *Model) drawMetadataContent(layout cardLayout) string {
	stats := model.libraryStats

	rows := []string{
		model.renderLabelValueRow(
			metadataCardTitleLabel,
			formatTrackCountValue(stats.MissingTitleCount),
			layout.innerWidth,
		),
		model.renderLabelValueRow(
			metadataCardArtistLabel,
			formatTrackCountValue(stats.MissingArtistCount),
			layout.innerWidth,
		),
		model.renderLabelValueRow(
			metadataCardAlbumLabel,
			formatTrackCountValue(stats.MissingAlbumCount),
			layout.innerWidth,
		),
		model.renderLabelValueRow(
			metadataCardDupesLabel,
			formatTrackCountValue(stats.DuplicatedTrackCount),
			layout.innerWidth,
		),
	}

	return strings.Join(rows, "\n")
}

// drawFilesByFormatContent draws a list of files formats, drawing a share bar and the total size of files with that format.
func (model *Model) drawFilesByFormatContent(layout cardLayout) string {
	stats := model.libraryStats

	if stats.FilesCount == 0 {
		return model.style.empty.Render(emptyPlaceholder)
	}

	formatBars := model.formatBars(stats)
	visibleCount := min(len(formatBars), formatCardVisibleRows)

	rows := make([]string, 0, visibleCount+1)
	for _, bar := range formatBars[:visibleCount] {
		rows = append(rows, model.renderFilesByFormatRow(bar))
	}

	if len(formatBars) > visibleCount {
		rows = append(
			rows,
			model.style.muted.Render(fmt.Sprintf(formatCardMoreText, len(formatBars)-visibleCount)),
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
		rowFormat := format
		if rowFormat == "" {
			rowFormat = formatCardUnknownValue
		}

		formatBars = append(formatBars, formatBar{
			format:   rowFormat,
			count:    count,
			bytes:    stats.BytesPerFormat[format],
			fraction: float64(count) / float64(stats.FilesCount),
		})
	}

	// sort by count descending, then by format name.
	slices.SortStableFunc(formatBars, func(left formatBar, right formatBar) int {
		return cmp.Or(
			cmp.Compare(right.count, left.count),
			strings.Compare(left.format, right.format),
		)
	})

	return formatBars
}

// renderFilesByFormatRow draws a single format row with a name, count, size, and a share bar.
func (model *Model) renderFilesByFormatRow(bar formatBar) string {
	name := model.style.rowLabel.Render(fmt.Sprintf("%-*s", formatCardKeyWidth, bar.format))
	shareBar := model.renderShareBar(bar.fraction, formatCardBarWidth)
	count := model.style.rowValue.Render(fmt.Sprintf("%*d", formatCardCountWidth, bar.count))
	size := model.style.muted.Render(
		fmt.Sprintf(" %*s", formatCardBytesColumnWidth, audio.GetReadableByteSize(bar.bytes)),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, name, shareBar+" ", count, size)
}

// renderShareBar builds a filled/empty glyph bar of the given width.
func (model *Model) renderShareBar(fraction float64, width int) string {
	var shareBarLine strings.Builder
	filledTotal := int(min(max(fraction, 0), 1) * float64(width))

	shareBarLine.WriteString(model.style.formatShareBar.Render(strings.Repeat("█", filledTotal)))
	if width-filledTotal > 0 {
		shareBarLine.WriteString(
			model.style.formatShareEmptyBar.Render(strings.Repeat("░", width-filledTotal)),
		)
	}

	return shareBarLine.String()
}

// drawTopArtistsContent draws the artists rankings: by files, by albums, and by playtime.
func (model *Model) drawTopArtistsContent(layout cardLayout) string {
	stats := model.libraryStats

	return model.renderTopArtistsCard([]topArtistsSection{
		{title: topArtistsByFilesTitle, counts: stats.TopArtistsByFiles, formatValue: formatCountValue},
		{title: topArtistsByAlbumsTitle, counts: stats.TopArtistsByAlbums, formatValue: formatCountValue},
		{
			title:       topArtistsByDurationTitle,
			counts:      stats.TopArtistsByDuration,
			formatValue: formatDurationValue,
		},
	}, layout)
}

// renderTopArtistsCard draws named sub-sections of formatted "name value" rows.
func (model *Model) renderTopArtistsCard(sections []topArtistsSection, layout cardLayout) string {
	entryLinesCount := topArtistsCardSectionEntries

	// find out what is gonna be the longest rendered value to allocate the proper width "budget"
	valueColumnWidth := 0
	for _, section := range sections {
		for _, count := range section.counts {
			valueColumnWidth = max(valueColumnWidth, lipgloss.Width(section.formatValue(count.Value)))
		}
	}
	nameWidth := max(layout.innerWidth-valueColumnWidth, 0)

	rows := make([]string, 0, len(sections)*(entryLinesCount+1)) // each section is lines + subtitle
	for _, section := range sections {
		rows = append(rows, model.style.cardTitle.Render(section.title))

		if len(section.counts) == 0 {
			rows = append(rows, model.style.empty.Render(emptyPlaceholder))
			continue
		}

		visibleCounts := section.counts[:min(len(section.counts), entryLinesCount)]
		for _, count := range visibleCounts {
			rows = append(
				rows,
				model.renderTopArtistRow(count.Name, section.formatValue(count.Value), nameWidth, valueColumnWidth),
			)
		}
	}

	return strings.Join(rows, "\n")
}

// renderTopArtistRow draws a leaderboard-y "{name} {value}" row, truncating long names.
func (model *Model) renderTopArtistRow(name string, value string, nameWidth int, valueWidth int) string {
	truncatedName := ansi.Truncate(name, nameWidth, "...")

	nameColumn := model.style.rowLabel.Width(nameWidth).Render(truncatedName)
	valueColumn := model.style.rowValue.Width(valueWidth).Align(lipgloss.Right).Render(model.renderStatValue(value))

	return lipgloss.JoinHorizontal(lipgloss.Top, nameColumn, valueColumn)
}

// drawPlaceholderContent draws nothing: the placeholder card is an empty reserved slot.
// TODO: maybe some glitch animation? or something else?
func (model *Model) drawPlaceholderContent(_ cardLayout) string {
	return ""
}

// drawTrackLengthsContent draws the track lengths card rows.
func (model *Model) drawTrackLengthsContent(layout cardLayout) string {
	stats := model.libraryStats

	if !stats.HasTrackLengths {
		return model.style.empty.Render(noTracksWithLengthText)
	}

	rows := model.renderAnnotatedLengthRows([]lengthsGroupEntry{
		{
			label:      lengthsCardLongestLabel,
			value:      formatDurationValue(stats.LongestTrack.Value),
			annotation: stats.LongestTrack.Name,
		},
		{
			label:      lengthsCardShortestLabel,
			value:      formatDurationValue(stats.ShortestTrack.Value),
			annotation: stats.ShortestTrack.Name,
		},
		{
			label:      lengthsCardAverageLabel,
			value:      formatDurationValue(stats.AverageLengthTrack.Value),
			annotation: trackLengthsAverageValue,
		},
	}, layout)

	return strings.Join(rows, "\n")
}

// drawAlbumLengthsContent draws the album lengths card rows.
func (model *Model) drawAlbumLengthsContent(layout cardLayout) string {
	stats := model.libraryStats

	if !stats.HasAlbumLengths {
		return model.style.empty.Render(noTracksWithLengthText)
	}

	rows := model.renderAnnotatedLengthRows([]lengthsGroupEntry{
		{
			label:      lengthsCardLongestLabel,
			value:      formatDurationValue(stats.LongestAlbum.Value),
			annotation: stats.LongestAlbum.Name,
		},
		{
			label:      lengthsCardShortestLabel,
			value:      formatDurationValue(stats.ShortestAlbum.Value),
			annotation: stats.ShortestAlbum.Name,
		},
		{
			label:      lengthsCardAverageLabel,
			value:      formatDurationValue(stats.AverageLengthAlbum.Value),
			annotation: albumLengthsAverageValue,
		},
	}, layout)

	return strings.Join(rows, "\n")
}

// renderAnnotatedLengthRows draws "label value (annotation)" rows for a length card.
func (model *Model) renderAnnotatedLengthRows(rows []lengthsGroupEntry, layout cardLayout) []string {
	// the duration column fits the widest value, while the the annotation column takes whatever is left.
	durationColumnWidth := 0
	for _, annotatedRow := range rows {
		durationColumnWidth = max(durationColumnWidth, lipgloss.Width(annotatedRow.value))
	}
	annotationColumnWidth := max(layout.innerWidth-lengthsGroupLabelWidth-durationColumnWidth, 0)

	durationColumn := model.style.rowValue.Width(durationColumnWidth).Align(lipgloss.Right)
	annotationColumn := model.style.muted.Width(annotationColumnWidth).Align(lipgloss.Right)

	renderedRows := make([]string, 0, len(rows))
	for _, annotatedRow := range rows {
		annotationText := ""
		if annotatedRow.annotation != "" {
			annotationText = ansi.Truncate(
				fmt.Sprintf(annotatedValueText, annotatedRow.annotation),
				annotationColumnWidth,
				"...",
			)
		}

		renderedRows = append(renderedRows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.style.rowLabel.Render(fmt.Sprintf("%-*s", lengthsGroupLabelWidth, annotatedRow.label)),
			durationColumn.Render(model.renderStatValue(annotatedRow.value)),
			annotationColumn.Render(annotationText),
		))
	}

	return renderedRows
}

// formatTrackCountValue formats a track-count value (suffixing with " tracks"), empty on zero.
func formatTrackCountValue(count int) string {
	if count == 0 {
		return ""
	}

	return fmt.Sprintf("%d tracks", count)
}

// formatCountValue formats a numeric value, empty on zero.
func formatCountValue(count int) string {
	if count == 0 {
		return ""
	}

	return fmt.Sprintf("%d", count)
}

// formatBytesValue formats a readable byte size, empty on zero.
func formatBytesValue(bytes int64) string {
	if bytes == 0 {
		return ""
	}

	return audio.GetReadableByteSize(bytes)
}

// formatDurationValue formats a readable duration, empty on zero.
func formatDurationValue(seconds int) string {
	if seconds == 0 {
		return ""
	}

	return audio.GetReadableDuration(seconds)
}
