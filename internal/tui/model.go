package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"krillion-solver/internal/clipboard"
	"krillion-solver/internal/game"
)

type loaded struct {
	result game.Result
	err    error
}
type tick time.Time
type copied struct{ err error }

type Model struct {
	load           func() (game.Result, error)
	result         game.Result
	err            error
	loading, demo  bool
	width, height  int
	selected, pick int
	scroll, frame  int
	notice         string
	noticeUntil    time.Time
	copy           func(string) error
}

func New(load func() (game.Result, error), demo bool) Model {
	return Model{load: load, copy: clipboard.Copy, loading: true, demo: demo, width: 100, height: 32}
}

func (m Model) Failed() bool { return m.err != nil }

func (m Model) fetch() tea.Msg {
	r, err := m.load()
	return loaded{r, err}
}

func (m Model) pulse() tea.Cmd {
	interval := time.Second
	if m.loading {
		interval = 150 * time.Millisecond
	}
	return tea.Tick(interval, func(at time.Time) tea.Msg { return tick(at) })
}

func (m Model) Init() tea.Cmd { return tea.Batch(m.fetch, m.pulse()) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.scroll = 0
	case loaded:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.result = msg.result
			m.selected, m.pick, m.scroll = 0, 0, 0
		}
	case copied:
		if msg.err != nil {
			m.notice = "Copy failed: " + msg.err.Error()
		} else {
			m.notice = "Copied to clipboard"
		}
		m.noticeUntil = time.Now().Add(5 * time.Second)
	case tick:
		m.frame++
		if time.Now().After(m.noticeUntil) {
			m.notice = ""
		}
		return m, m.pulse()
	case tea.KeyPressMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "r":
			if !m.loading {
				m.loading, m.err, m.notice = true, nil, ""
				return m, m.fetch
			}
		}
		if m.loading || m.err != nil || len(m.result.Prompts) == 0 {
			return m, nil
		}
		previous := m.selected
		switch key {
		case "down", "j", "tab":
			m.selected = (m.selected + 1) % len(m.result.Prompts)
		case "up", "k", "shift+tab":
			m.selected = (m.selected + len(m.result.Prompts) - 1) % len(m.result.Prompts)
		case "1", "2", "3", "4", "5", "6", "7":
			m.selected = int(key[0] - '1')
		case "left", "h", "right", "l":
			count := len(m.result.Prompts[m.selected].Winners)
			delta := 1
			if key == "left" || key == "h" {
				delta = -1
			}
			m.pick = (m.pick + count + delta) % count
			m.scroll = 0
		case "pgdown", "ctrl+d":
			m.scroll += 5
		case "pgup", "ctrl+u":
			m.scroll = max(0, m.scroll-5)
		case "c", "a":
			text := m.result.Prompts[m.selected].Winners[m.pick].Answer
			if key == "a" {
				text = m.result.Plain()
			}
			text = game.SafeText(text)
			return m, m.copyText(text)
		}
		if m.selected != previous {
			m.pick, m.scroll = 0, 0
		}
	}
	return m, nil
}

func (m Model) copyText(text string) tea.Cmd {
	return func() tea.Msg { return copied{err: m.copy(text)} }
}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.BackgroundColor = ocean
	v.ForegroundColor = pearl
	v.WindowTitle = "Krillion Solver"
	return v
}

func (m Model) render() string {
	if m.width < 42 || m.height < 20 {
		return ansi.Truncate("KRILLION · resize to 42 × 20 · q quit", max(1, m.width), "")
	}
	w := min(112, m.width-4)
	header := m.header(w)
	footer := m.footer(w)
	bodyHeight := max(5, m.height-lipgloss.Height(header)-lipgloss.Height(footer)-4)
	var body string
	switch {
	case m.loading:
		frames := []string{"·     ", "° ·   ", "◦ ° · ", "○ ◦ ° ", "◦ ○ ◦ ", "° ◦ ○ "}
		bubble := ink(cyan).Render(frames[m.frame%len(frames)])
		content := bubble + "\n\n" + ink(pearl).Bold(true).Render("Descending into today's dive…") + "\n\n" + ink(muted).Render("Following the current. Finding the gold.")
		body = lipgloss.Place(w, bodyHeight, lipgloss.Center, lipgloss.Center, content)
	case m.err != nil:
		content := ink(coral).Bold(true).Render("The current got away.") + "\n\n" + ink(pearl).Width(w-8).Render(game.SafeText(m.err.Error())) + "\n\n" + ink(muted).Render("r retry   ·   q surface")
		body = lipgloss.Place(w, bodyHeight, lipgloss.Center, lipgloss.Center, content)
	default:
		if w >= 88 && bodyHeight >= 17 {
			leftWidth := 39
			list := panel(m.roundList(leftWidth-4, bodyHeight-4 < 22), leftWidth, bodyHeight, line)
			rightWidth := w - leftWidth - 2
			detail := panel(m.detail(rightWidth-4, bodyHeight-4), rightWidth, bodyHeight, gold)
			body = lipgloss.JoinHorizontal(lipgloss.Top, list, "  ", detail)
		} else {
			// Compact terminals keep the selected answer visible and every round reachable.
			strip := m.roundTabs(w)
			box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gold).
				Padding(0, 1).Width(w).Height(bodyHeight - 2).MaxHeight(bodyHeight - 2)
			body = strip + "\n" + box.Render(m.detail(w-4, bodyHeight-4))
		}
	}
	content := header + "\n\n" + body + "\n\n" + footer
	return lipgloss.NewStyle().Padding(0, 2).Render(content)
}

func (m Model) header(w int) string {
	title := ink(coral).Bold(true).Render("KRILLION SOLVER")
	if m.height >= 32 && w >= 88 {
		title = ink(coral).Bold(true).Render(asciiLogo)
	}
	meta := "7 prompts. Every perfect answer."
	if !m.loading && m.err == nil && len(m.result.Prompts) > 0 {
		meta = fmt.Sprintf("DIVE %03d  /  %s  /  %d top answers", m.result.DayNumber, m.result.Date, m.result.WinnerCount())
		if n := m.result.FallbackCount(); n > 0 {
			meta += fmt.Sprintf("  /  %d without Krillion", n)
		}
		if m.demo {
			meta = "ILLUSTRATIVE ANSWERS / not today's puzzle"
		}
	}
	return title + "\n" + ink(line).Render(strings.Repeat("─", w)) + "\n" + ink(muted).Render(ansi.Truncate(meta, w, "…"))
}

func (m Model) roundList(w int, compact bool) string {
	var rows []string
	rows = append(rows, ink(muted).Render("THE CATCH"), "")
	for i, round := range m.result.Prompts {
		color := muted
		marker := " "
		if i == m.selected {
			color, marker = gold, "›"
		}
		name := game.SafeText(round.Winners[0].Answer)
		if len(round.Winners) > 1 {
			name += fmt.Sprintf(" +%d", len(round.Winners)-1)
		}
		row := ink(color).Bold(i == m.selected).Render(fmt.Sprintf("%s %02d  %s", marker, i+1, ansi.Truncate(name, w-7, "…")))
		rows = append(rows, row)
		if !compact {
			rows = append(rows, "     "+ink(muted).Render(ansi.Truncate(game.SafeText(round.Text), w-5, "…")))
		}
		if !compact && i < len(m.result.Prompts)-1 {
			rows = append(rows, "")
		}
	}
	return strings.Join(rows, "\n")
}

func (m Model) roundTabs(w int) string {
	var tabs []string
	for i := range m.result.Prompts {
		style := ink(muted)
		if i == m.selected {
			style = ink(gold).Bold(true)
		}
		tabs = append(tabs, style.Render(fmt.Sprintf(" %d ", i+1)))
	}
	score := m.result.Prompts[m.selected].Winners[m.pick].Score
	return between(strings.Join(tabs, " "), ink(gold).Render(fmt.Sprintf("+%d", score)), w)
}

func (m Model) detail(w, h int) string {
	p := m.result.Prompts[m.selected]
	a := p.Winners[m.pick]
	var parts []string
	parts = append(parts, between(ink(cyan).Render(fmt.Sprintf("ROUND %02d / 07", m.selected+1)), ink(gold).Bold(true).Render(fmt.Sprintf("+%d", a.Score)), w), "")
	if h < 12 {
		parts = []string{parts[0], ink(gold).Bold(true).Width(w).Render("◆ " + game.SafeText(a.Answer)), "", ink(pearl).Width(w).Render(game.SafeText(p.Text)), ""}
		if p.Fallback {
			parts = []string{parts[0], ink(coral).Width(w).Render("No One in a Krillion answer"), ink(gold).Bold(true).Width(w).Render("◆ " + game.SafeText(a.Answer)), ink(pearl).Width(w).Render(game.SafeText(p.Text)), ""}
		}
	} else {
		parts = append(parts, ink(pearl).Width(w).Render(game.SafeText(p.Text)), "", ink(line).Render(strings.Repeat("─", w)), "")
		label := "◆ ONE IN A KRILLION"
		if p.Fallback {
			label = "◆ BEST AVAILABLE · NO ONE IN A KRILLION"
		}
		parts = append(parts, ink(gold).Render(label), ink(gold).Bold(true).Width(w).Render(game.SafeText(a.Answer)), "")
	}
	if a.Quip != "" {
		parts = append(parts, ink(muted).Width(w).Render(game.SafeText(a.Quip)), "")
	}
	if len(p.Winners) > 1 {
		parts = append(parts, ink(cyan).Width(w).Render(fmt.Sprintf("%d equally scoring picks · ← / → to switch", len(p.Winners))))
		for i, alt := range p.Winners {
			if i != m.pick {
				parts = append(parts, ink(pearl).Width(w).Render("Also: "+game.SafeText(alt.Answer)))
			}
		}
		parts = append(parts, "")
	}
	parts = append(parts, ink(muted).Width(w).Render(fmt.Sprintf("Selected from %s accepted answers.", comma(p.Accepted))))
	if m.demo {
		parts = append(parts, "", ink(coral).Width(w).Render("Demo only. Scores and answer sets here are illustrative."))
	}
	lines := strings.Split(strings.Join(parts, "\n"), "\n")
	h = max(1, h)
	if len(lines) <= h {
		return strings.Join(lines, "\n")
	}
	offset := min(m.scroll, max(0, len(lines)-max(1, h-1)))
	visible := lines[offset:min(len(lines), offset+max(1, h-1))]
	visible = append(visible, ink(muted).Render(ansi.Truncate("pgup / pgdn scroll details", w, "")))
	return strings.Join(visible, "\n")
}

func (m Model) footer(w int) string {
	stat := "Loading today's answers"
	if !m.loading && m.err == nil && len(m.result.Prompts) > 0 {
		stat = fmt.Sprintf("%d / 700 possible points · 7 / 7 rounds", m.result.Score())
		if n := m.result.FallbackCount(); n > 0 {
			rounds := "rounds"
			if n == 1 {
				rounds = "round"
			}
			stat = fmt.Sprintf("%d points · %d %s without Krillion", m.result.Score(), n, rounds)
		}
	}
	reset := ""
	if m.result.ResetsAt != nil && !m.demo {
		remaining := time.Until(*m.result.ResetsAt)
		if remaining > 0 {
			h := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			reset = fmt.Sprintf("resets in %02dh %02dm", h, minutes)
		} else {
			reset = "new dive available · r refresh"
		}
	}
	keys := "↑↓/jk round · c copy · r refresh · q quit"
	if w >= 88 {
		keys = "↑↓ / jk / 1–7 round   ←→ alternate   c copy   a copy all   r refresh   q quit"
	}
	if m.notice != "" {
		keys = m.notice
	}
	return between(ink(cyan).Render(stat), ink(muted).Render(reset), w) + "\n\n" + ink(muted).Render(ansi.Truncate(keys, w, "…"))
}
