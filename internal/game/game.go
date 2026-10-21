package game

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Answer retains the canonical spelling the game accepts.
type Answer struct {
	Answer string `json:"answer"`
	Tier   string `json:"tier"`
	Score  int    `json:"score"`
	Quip   string `json:"quip,omitempty"`
}

type Round struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// Winners contains every highest-scoring answer, including lower-tier picks when needed.
	Winners  []Answer `json:"winners"`
	Accepted int      `json:"accepted"`
	Fallback bool     `json:"fallback,omitempty"`
}

type Result struct {
	Date      string     `json:"date"`
	DayNumber int        `json:"dayNumber"`
	ResetsAt  *time.Time `json:"resetsAt,omitempty"`
	Prompts   []Round    `json:"prompts"`
	Demo      bool       `json:"demo"`
}

func (r Result) Score() int {
	total := 0
	for _, round := range r.Prompts {
		if len(round.Winners) > 0 {
			total += round.Winners[0].Score
		}
	}
	return total
}

func (r Result) WinnerCount() int {
	total := 0
	for _, round := range r.Prompts {
		total += len(round.Winners)
	}
	return total
}

func (r Result) FallbackCount() int {
	total := 0
	for _, round := range r.Prompts {
		if round.Fallback {
			total++
		}
	}
	return total
}

// SafeText removes terminal controls from remote text without changing stored answers.
func SafeText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			if unicode.IsSpace(r) {
				return ' '
			}
			return -1
		}
		return r
	}, s)
}

func (r Result) Plain() string {
	var out strings.Builder
	if r.Demo {
		out.WriteString("DEMO — illustrative answers, not today's puzzle\n")
	} else {
		fmt.Fprintf(&out, "KRILLION · Dive #%d · %s\n", r.DayNumber, r.Date)
	}
	fmt.Fprintf(&out, "%d rounds · %d top answers · %d best available points\n", len(r.Prompts), r.WinnerCount(), r.Score())
	if n := r.FallbackCount(); n > 0 {
		rounds := "rounds"
		if n == 1 {
			rounds = "round"
		}
		fmt.Fprintf(&out, "Note: %d %s had no One in a Krillion answer; showing the highest-scoring available picks.\n", n, rounds)
	}
	out.WriteByte('\n')
	for i, p := range r.Prompts {
		fmt.Fprintf(&out, "%d. %s\n", i+1, SafeText(p.Text))
		if p.Fallback {
			out.WriteString("   No One in a Krillion answer — best available:\n")
		}
		for _, a := range p.Winners {
			fmt.Fprintf(&out, "   %s  (+%d)\n", SafeText(a.Answer), a.Score)
		}
		out.WriteByte('\n')
	}
	return out.String()
}
