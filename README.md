# Krillion Solver

See every **One in a Krillion** answer for today's seven prompts. Browse the prompts, read each answer's description, switch between equally scoring picks, and copy an answer or the full list.

## Run

Requires Go 1.25 or later.

```sh
make
./krillion
```

On Windows without Make:

```powershell
go build -o krillion.exe .
.\krillion.exe
```

Use `↑`/`↓` or `j`/`k` to move between prompts, `←`/`→` to switch alternate answers, `c` to copy the selected answer, and `a` to copy every answer.

`make release` builds macOS, Linux, and Windows binaries in `dist/`.
