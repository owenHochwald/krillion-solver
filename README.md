# Krillion · the daily deep cut

An ocean themed Go terminal app built with Bubble Tea v2 and Lip Gloss v2. Fetches the daily puzzle and prints its curated **One in a Krillion** answers, including every alternative when a round has multiple winners.

## Run

Requires Go 1.25 or later and internet access to download dependencies on the first build.

```sh
go build -o krillion .
./krillion
```

Preview the full interface without network access:

```sh
./krillion --demo
```

Demo mode is explicitly labeled and contains illustrative prompts, answers, scores, and counts. It does **not** solve today's puzzle.

```sh
./krillion --plain           # Every prompt and all of its +100 answers
./krillion --json            # Structured result for scripts or a future frontend
./krillion --timeout 15s     # Overall fetch deadline
```

Piping output automatically selects plain text. Use `--json` for machine readable output. Live results are never replaced with demo data on failure.

## Inside the dive

Navy ocean background, cyan metadata, coral lettering, gold winning answers, and a trench depth bar. Wide terminals show all seven rounds beside the selected answer. Narrow terminals use numbered round tabs. The minimum size is 42 columns by 20 rows.

| Key | Action |
| --- | --- |
| `↑` / `↓`, `j` / `k`, `Tab` | Change round |
| `1`–`7` | Jump to a round |
| `←` / `→`, `h` / `l` | Switch between equally scoring answers |
| `PgUp` / `PgDn`, `Ctrl+U` / `Ctrl+D` | Scroll long details |
| `c` | Copy the selected answer |
| `a` | Copy every prompt and winning answer |
| `r` | Fetch the current daily again |
| `q`, `Esc`, `Ctrl+C` | Quit |

Copy uses the terminal's OSC 52 clipboard protocol; support and permissions depend on the terminal. The app reports that it sent a clipboard request, rather than claiming the clipboard was updated.

## How solving works

1. `GET https://krillion.io/api/today` supplies the date, daily number, and seven prompt IDs.
2. `GET https://krillion.io/api/reveal?date=<server date>` supplies all answer sets.
3. Match rounds by ID and retain every answer whose tier is `krillion` and score is `100`.

The date always comes from the server. There is no hardcoded reset time, answer website, language model, or word dictionary. Each load normally makes two anonymous GET requests, with no login, submissions, completion request, or progress changes. A date mismatch gets one fresh attempt; other failures do not retry automatically. There is no persistent cache or background polling. If provided as an RFC 3339 string, `resetsAt` powers the displayed countdown; after reset, press `r` to load the new daily.

A second winning answer is an alternative, not another 100 points. Seven successful rounds have a possible score of 700. Missing winners, unexpected scores, invalid data, denied access, and rate limits produce an error instead of an incomplete success.

This uses an undocumented endpoint. Its accessibility can change. The anonymous reveal behavior was established in the supplied inspection of dive #98 on 21 October 2026. The implementation was verified against in-memory HTTP fixtures; this development session could not reach Krillion directly.

## Development

```sh
go test ./...
go vet ./...
```

`main.go` handles flags and terminal detection. `internal/game` contains the HTTP client, result types, and offline demo. `internal/tui` owns the Bubble Tea model and theme. Tests cover anonymous requests, multiple winners, shuffled prompt order, malformed data, bounded retries, cancellation, CLI output, navigation, and terminal layouts.
