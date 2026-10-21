package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"

	"krillion-solver/internal/game"
	"krillion-solver/internal/tui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("krillion", flag.ContinueOnError)
	flags.SetOutput(errOut)
	demo := flags.Bool("demo", false, "preview the ocean interface with illustrative data (offline)")
	plain := flags.Bool("plain", false, "print all winning answers without the interactive interface")
	jsonOutput := flags.Bool("json", false, "print winning answers as JSON")
	timeout := flags.Duration("timeout", 25*time.Second, "overall time limit for fetching a daily puzzle")
	flags.Usage = func() {
		fmt.Fprint(errOut, "KRILLION · the daily deep cut\n\nUsage: krillion [--demo] [--plain | --json]\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 || (*plain && *jsonOutput) || *timeout <= 0 {
		fmt.Fprintln(errOut, "use no positional arguments, choose either --plain or --json, and set a positive timeout")
		return 2
	}
	load := func() (game.Result, error) {
		if *demo {
			return game.Demo(), nil
		}
		requestCtx, cancel := context.WithTimeout(ctx, *timeout)
		defer cancel()
		return game.NewClient().Solve(requestCtx)
	}
	interactive := !*plain && !*jsonOutput && term.IsTerminal(os.Stdout.Fd()) && term.IsTerminal(os.Stdin.Fd())
	if interactive {
		final, err := tea.NewProgram(tui.New(load, *demo), tea.WithContext(ctx)).Run()
		if err != nil {
			fmt.Fprintln(errOut, "krillion:", err)
			return 1
		}
		if m, ok := final.(tui.Model); ok && m.Failed() {
			return 1
		}
		return 0
	}
	result, err := load()
	if err != nil {
		fmt.Fprintln(errOut, "krillion:", err)
		return 1
	}
	if *jsonOutput {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(result)
	} else {
		_, err = io.WriteString(out, result.Plain())
	}
	if err != nil {
		fmt.Fprintln(errOut, "krillion:", err)
		return 1
	}
	return 0
}
