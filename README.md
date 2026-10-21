# Krillion Solver

An unofficial solver for [Krillion](https://krillion.io/). See every **One in a Krillion** answer for today's seven prompts, read its description, switch between equally scoring picks, and copy an answer or the full list.

![Krillion Solver showing today's answers and alternate picks](assets/krillion-solver.png)

## Run

[Download a macOS, Linux, or Windows build](https://github.com/owenHochwald/krillion-solver/releases/latest), or build from source with Go 1.25 or later:

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
