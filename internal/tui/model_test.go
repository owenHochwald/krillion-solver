package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"krillion-solver/internal/game"
)

func ready() Model {
	m := New(func() (game.Result, error) { return game.Demo(), nil }, true)
	m.loading, m.result = false, game.Demo()
	return m
}

func TestLayoutsFitAndShowWinner(t *testing.T) {
	for _, size := range [][2]int{{112, 36}, {100, 32}, {80, 24}, {44, 20}, {160, 48}, {30, 10}} {
		m := ready()
		m.width, m.height = size[0], size[1]
		view := m.View().Content
		if lipgloss.Width(view) > m.width || lipgloss.Height(view) > m.height {
			t.Errorf("layout %dx%d overflow: %dx%d\n%s", m.width, m.height, lipgloss.Width(view), lipgloss.Height(view), ansi.Strip(view))
		}
		if m.width >= 42 && m.height >= 20 && !strings.Contains(view, "Dekalog") {
			t.Errorf("winner missing at %dx%d", m.width, m.height)
		}
	}
}

func TestRoundAndAlternateNavigation(t *testing.T) {
	m := ready()
	next, _ := m.Update(tea.KeyPressMsg{Code: '4', Text: "4"})
	m = next.(Model)
	if m.selected != 3 {
		t.Fatal("number key did not select round four")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = next.(Model)
	if m.pick != 1 || !strings.Contains(m.View().Content, "Hydrofoil") {
		t.Fatal("alternate answer was not selected")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(Model)
	if m.selected != 4 || m.pick != 0 {
		t.Fatal("changing rounds must reset alternate selection")
	}
}
