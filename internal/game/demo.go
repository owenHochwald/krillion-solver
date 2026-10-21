package game

import "fmt"

// Demo contains illustrative data only; it never represents a live daily puzzle.
func Demo() Result {
	samples := []struct{ prompt, answer, quip string }{
		{"Name a TV show that's rated 8.5+ on IMDb (1,000+ votes)", "Dekalog", "Ten hour-long films, one for each of the Ten Commandments."},
		{"Name a creature that lives in the deep ocean", "Dumbo octopus", "An umbrella of arms. A very small ghost in a very big ocean."},
		{"Name an animal that produces its own light", "Atolla jellyfish", "When trouble comes, this jellyfish puts on a light show."},
		{"Name a vehicle that doesn't travel on wheels", "Bathyscaphe", "A tiny cabin, a giant ocean, and a very long way down."},
		{"Name a word for a shade of blue", "Smalt", "Blue glass, ground into a pigment. A little ocean in a bottle."},
		{"Name a country that borders Greece or Austria", "Liechtenstein", "A small place with an excellent view."},
		{"Name a place deeper than 6,000 metres", "Challenger Deep", "The end of the descent. There is nowhere deeper to go."},
	}
	r := Result{Date: "2026-10-21", DayNumber: 98, Demo: true}
	for i, s := range samples {
		r.Prompts = append(r.Prompts, Round{
			ID: fmt.Sprintf("p%d", i+1), Text: s.prompt, Accepted: 100 + i*71,
			Winners: []Answer{{Answer: s.answer, Tier: "krillion", Score: 100, Quip: s.quip}},
		})
	}
	// Exercise the real API's multiple-winner case in the preview, too.
	r.Prompts[3].Winners = append(r.Prompts[3].Winners, Answer{
		Answer: "Hydrofoil", Tier: "krillion", Score: 100,
		Quip: "A second illustrative winner. Either answer earns the round's points.",
	})
	return r
}
