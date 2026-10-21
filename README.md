# Krillion Solver

An unofficial solver for [Krillion](https://krillion.io/). See today's seven prompts and their top answers, read each description, switch between equal picks, and copy an answer or the full list. If a round has no One in a Krillion answer, the app labels its highest-scoring available pick.

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

## macOS menu bar app

The TUI also has a companion menu bar app for macOS 14 or later. Click its shrimp icon to see all seven answers in a compact panel, expand a round for the prompt and alternate winners, or copy an answer. It works alongside any browser.

With Xcode installed, run:

```sh
make macos-app
open dist/xcode/Build/Products/Release/KrillionBar.app
```

The app uses the same daily Krillion endpoints and validation rules as the TUI. The generated Xcode project is also at `macos/KrillionBar.xcodeproj`; edit its source files under `macos/KrillionBar/`. The local `make macos-app` build is unsigned. A distributed build needs normal macOS signing and notarization.
