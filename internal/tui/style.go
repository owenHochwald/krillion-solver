package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	ocean = lipgloss.Color("#061623")
	pearl = lipgloss.Color("#E4EDF1")
	muted = lipgloss.Color("#7895A8")
	cyan  = lipgloss.Color("#56D8ED")
	gold  = lipgloss.Color("#F4C95D")
	coral = lipgloss.Color("#F06479")
	line  = lipgloss.Color("#244354")
)

func ink(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

func panel(content string, width, height int, border color.Color) string {
	// Lip Gloss v2 dimensions include borders and padding.
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(border).
		Padding(1).Width(max(1, width)).Height(max(1, height)).
		MaxWidth(width).MaxHeight(height).Render(content)
}

func between(left, right string, width int) string {
	if lipgloss.Width(left)+lipgloss.Width(right)+1 > width {
		return ansi.Truncate(left, width, "…")
	}
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right))) + right
}

func comma(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
