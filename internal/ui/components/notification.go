package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	lipgloss "github.com/charmbracelet/lipgloss"
	config "wired/internal/config"
)

type NotificationType int

const (
	NotificationInfo NotificationType = iota
	NotificationError
	NotificationSuccess
)

type Notification struct {
	Message       string
	Type          NotificationType
	Sequence      int
	ExpiresAt     time.Time
	TotalDuration time.Duration
}

type Notifications struct {
	queue    []Notification
	sequence int
}

func (notifications *Notifications) Enqueue(message string, notificationType NotificationType, duration time.Duration) {
	if len(notifications.queue) == 0 {
		notifications.sequence = 0
	}

	notifications.sequence++
	notifications.queue = append(
		notifications.queue,
		NewNotification(message, notificationType, notifications.sequence, duration),
	)
}

func (notifications *Notifications) Prune() {
	if len(notifications.queue) == 0 {
		return
	}

	notifications.queue = pruneExpired(notifications.queue)
}

func (notifications *Notifications) Visible(shownMax int) []Notification {
	return visibleNotifications(notifications.queue, shownMax)
}

func (notification Notification) Expired() bool {
	return time.Now().After(notification.ExpiresAt)
}

func NewNotification(
	message string,
	notificationType NotificationType,
	sequence int,
	duration time.Duration,
) Notification {
	return Notification{
		Message:       message,
		Type:          notificationType,
		Sequence:      sequence,
		ExpiresAt:     time.Now().Add(duration),
		TotalDuration: duration,
	}
}

func RenderNotification(notification Notification, cfg *config.Config) string {
	var colorHex string

	switch notification.Type {
	case NotificationError:
		colorHex = cfg.Colors.NotificationError
	case NotificationSuccess:
		colorHex = cfg.Colors.NotificationSuccess
	default:
		colorHex = cfg.Colors.NotificationInfo
	}

	// Fade out to black
	// TODO: do we really want to fade out to black? is there a better way?
	remainingTime := time.Until(notification.ExpiresAt)

	var ratio float64
	if notification.TotalDuration <= 0 {
		ratio = 1
	} else {
		ratio = float64(remainingTime) / float64(notification.TotalDuration)
		ratio = math.Max(0.05, math.Min(1, ratio))
	}

	color := fadeColor(colorHex, ratio)

	style := lipgloss.NewStyle().
		Foreground(color).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)

	prefixedMsg := fmt.Sprintf("[%d] %s", notification.Sequence, notification.Message)

	// Padding + border = 4 (2+2=4)
	contentWidth := cfg.Notification.NotificationMaxWidth - 4
	lines := wrapText(prefixedMsg, contentWidth, cfg.Notification.NotificationMaxHeight)
	content := strings.Join(lines, "\n")

	return style.Render(content)
}

// TODO: for now we are acting like a stack, it's better to rewrite as a priority queue
func visibleNotifications(notifications []Notification, shownMax int) []Notification {
	if len(notifications) == 0 {
		return nil
	}

	start := 0
	if len(notifications) > shownMax {
		start = len(notifications) - shownMax
	}

	return notifications[start:]
}

func pruneExpired(notifications []Notification) []Notification {
	result := make([]Notification, 0, len(notifications))
	for _, n := range notifications {
		if !n.Expired() {
			result = append(result, n)
		}
	}

	return result
}

func wrapText(text string, width int, maxHeight int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string

	processWord := func(w string) string {
		for len(w) > width {
			lines = append(lines, w[:width])
			w = w[width:]
		}

		return w
	}

	current := processWord(words[0])

	for _, word := range words[1:] {
		if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = processWord(word)
		}
	}

	lines = append(lines, current)

	if len(lines) > maxHeight {
		if maxHeight <= 0 {
			return lines
		}

		lines = lines[:maxHeight]
		lastIdx := maxHeight - 1

		if width <= 3 {
			if width <= 0 {
				lines[lastIdx] = ""
			} else {
				ellipsis := "..."
				lines[lastIdx] = ellipsis[:min(len(ellipsis), width)]
			}
		} else {
			lastLine := lines[lastIdx]
			truncWidth := min(len(lastLine), width-3)
			lines[lastIdx] = lastLine[:truncWidth] + "..."
		}
	}

	return lines
}

// TODO: maybe this is too naive, or there's a better way to fade out I'm not seeing
func fadeColor(hex string, ratio float64) lipgloss.Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return lipgloss.Color("#" + hex)
	}

	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	r = uint64(float64(r) * ratio)
	g = uint64(float64(g) * ratio)
	b = uint64(float64(b) * ratio)

	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}
