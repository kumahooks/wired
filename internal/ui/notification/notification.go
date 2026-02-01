// Package notification implements a timed notification stack with fade-out rendering
package notification

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	lipgloss "github.com/charmbracelet/lipgloss"

	config "wired/internal/config"
)

type Type int

const (
	Info Type = iota
	Error
	Success
)

type Notification struct {
	message       string
	nType         Type
	sequence      int
	expiresAt     time.Time
	totalDuration time.Duration
}

func (n Notification) Expired() bool {
	return time.Now().After(n.expiresAt)
}

type Style struct {
	InfoColor    string
	ErrorColor   string
	SuccessColor string
	MaxWidth     int
	MaxHeight    int
}

func defaultStyle() Style {
	return Style{
		InfoColor:    "#539686",
		ErrorColor:   "#a52a2a",
		SuccessColor: "#639653",
		MaxWidth:     44,
		MaxHeight:    10,
	}
}

type StackNode struct {
	notification Notification
	prev         *StackNode
	next         *StackNode
}

type NotificationStack struct {
	head     *StackNode
	tail     *StackNode
	capacity int
	count    int
	sequence int
	style    Style
}

func New() NotificationStack {
	return NotificationStack{
		capacity: 32,
		style:    defaultStyle(),
	}
}

func (stack *NotificationStack) Push(message string, notificationType Type, duration time.Duration) {
	if stack.capacity <= 0 {
		return
	}

	stack.sequence++

	n := Notification{
		message:       message,
		nType:         notificationType,
		sequence:      stack.sequence,
		expiresAt:     time.Now().Add(duration),
		totalDuration: duration,
	}

	newNode := &StackNode{notification: n}

	if stack.head == nil {
		stack.head = newNode
		stack.tail = newNode
	} else {
		newNode.next = stack.head
		stack.head.prev = newNode
		stack.head = newNode
	}

	stack.count++

	if stack.count > stack.capacity {
		stack.removeTail()
	}
}

func (stack *NotificationStack) Prune() {
	for stack.tail != nil && stack.tail.notification.Expired() {
		stack.removeTail()
	}
}

func (stack *NotificationStack) Visible(maxShown int) []Notification {
	if stack.count == 0 || maxShown <= 0 {
		return nil
	}

	toShow := min(maxShown, stack.count)

	result := make([]Notification, toShow)
	current := stack.head

	for i := toShow - 1; i >= 0 && current != nil; i-- {
		result[i] = current.notification
		current = current.next
	}

	return result
}

func (stack *NotificationStack) ApplyConfig(cfg *config.Config) {
	stack.capacity = cfg.Notification.NotificationStackMax

	stack.style = Style{
		InfoColor:    cfg.Colors.NotificationInfo,
		ErrorColor:   cfg.Colors.NotificationError,
		SuccessColor: cfg.Colors.NotificationSuccess,
		MaxWidth:     cfg.Notification.NotificationMaxWidth,
		MaxHeight:    cfg.Notification.NotificationMaxHeight,
	}
}

func (stack NotificationStack) Render(n Notification) string {
	var colorHex string

	switch n.nType {
	case Error:
		colorHex = stack.style.ErrorColor
	case Success:
		colorHex = stack.style.SuccessColor
	default:
		colorHex = stack.style.InfoColor
	}

	remainingTime := time.Until(n.expiresAt)

	var ratio float64
	if n.totalDuration <= 0 {
		ratio = 1
	} else {
		ratio = float64(remainingTime) / float64(n.totalDuration)
		ratio = max(0.05, min(1, ratio))
	}

	color := fadeColor(colorHex, ratio)

	style := lipgloss.NewStyle().
		Foreground(color).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)

	prefixedMsg := fmt.Sprintf("[%d] %s", n.sequence, n.message)

	contentWidth := stack.style.MaxWidth - 4
	lines := wrapText(prefixedMsg, contentWidth, stack.style.MaxHeight)
	content := strings.Join(lines, "\n")

	return style.Render(content)
}

func (stack *NotificationStack) removeTail() {
	if stack.tail == nil {
		return
	}

	if stack.tail == stack.head {
		stack.head = nil
		stack.tail = nil
		stack.sequence = 0
	} else {
		stack.tail = stack.tail.prev
		stack.tail.next = nil
	}

	stack.count--
}

// wrapText wraps text to fit within width, counting single bytes per character
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

// TODO: do I really want to do it this way...
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
