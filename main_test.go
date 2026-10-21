package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"krillion-solver/internal/game"
)

func TestDemoOutput(t *testing.T) {
	for _, mode := range []string{"--plain", "--json"} {
		var out, errs bytes.Buffer
		if code := run(context.Background(), []string{"--demo", mode}, &out, &errs); code != 0 {
			t.Fatalf("exit %d: %s", code, &errs)
		}
		if mode == "--json" {
			var r game.Result
			if err := json.Unmarshal(out.Bytes(), &r); err != nil || !r.Demo || r.WinnerCount() != 8 {
				t.Fatalf("invalid JSON output: %s", &out)
			}
		} else if !strings.Contains(out.String(), "DEMO") || !strings.Contains(out.String(), "Hydrofoil") {
			t.Fatalf("incomplete plain output: %s", &out)
		}
	}
}

func TestInvalidFlags(t *testing.T) {
	for _, args := range [][]string{{"--plain", "--json"}, {"--web", "--json"}, {"unexpected"}, {"--timeout=0"}} {
		var out, errs bytes.Buffer
		if code := run(context.Background(), args, &out, &errs); code != 2 || out.Len() != 0 {
			t.Fatalf("bad arguments should fail without output: %v", args)
		}
	}
}
