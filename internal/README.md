# Terminal app

Download a [macOS, Linux, or Windows build](https://github.com/owenHochwald/krillion-solver/releases/latest), or build from the repository root with Go 1.25 or later:

```sh
make
./krillion
```

On Windows without Make:

```powershell
go build -o krillion.exe .
.\krillion.exe
```

Use `↑`/`↓` or `j`/`k` to choose a round, `←`/`→` to switch equal answers, `c` to copy one answer, `a` to copy all, and `r` to refresh. Press `q` to quit.

For scripts, run `./krillion --plain` or `./krillion --json`. Use `./krillion --demo` for an offline preview. `make release` builds all supported binaries in `dist/`.
