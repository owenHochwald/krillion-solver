# macOS answer overlay plan

## Goal

When a user opens the daily game at `krillion.io` in Safari on macOS, show today's seven top-scoring answers in a small, readable overlay on the page. The overlay opens automatically, can be collapsed or moved, and never enters answers into the game.

## Recommended shape

Build a **Safari Web Extension** with a content script scoped to `https://krillion.io/*`. This attaches the UI to the webpage instead of placing a free-floating macOS window over the browser. The same web assets can later be packaged as a Chrome or Edge extension. A small containing macOS app is needed for Safari installation and distribution, but it does not need to run the solver continuously.

The first version should show a narrow panel at the upper right of the game. Its header shows the dive date, a collapse button, and a refresh button. Seven numbered rows show one winner per prompt; opening a row reveals the prompt, alternate 100-point winners, and the quip. The panel should be draggable between left and right edges, remember its collapsed state, and remain usable on smaller windows. Use a dark, low-contrast surface so it is visible without obscuring the game. Do not auto-fill game controls.

## Data flow

```text
krillion.io tab
  -> content script mounts isolated overlay
  -> extension background fetches /api/today and /api/reveal?date=...
  -> shared validator matches rounds by ID and keeps tier=krillion, score=100
  -> content script renders seven rows and handles collapse/refresh
```

Keep the endpoint paths, date-match retry, seven-prompt check, duplicate-ID check, and winner validation consistent with `internal/game/client.go`. Port that small validation layer to JavaScript for the extension and use the existing Go client as the reference implementation. Calling the Go binary from Safari would require native messaging and a separate installed helper; that adds setup and failure points without helping the first version. The extension background should request only `krillion.io`, return structured data to the content script, and cache by puzzle date until reset or manual refresh. It should avoid polling and avoid reading the user's account, guesses, or cookies.

## Build sequence

1. **Site check:** Inspect the live daily game in Safari. Confirm the actual page URL, usable panel placement, whether navigation replaces the document or uses client-side routing, and whether the two anonymous endpoints still return the expected schema. Record any site elements the overlay must avoid.
2. **Extension core:** Create `extension/manifest.json`, a background fetch/validation module, and a content script. Scope page injection and network access to `https://krillion.io/*`. Fetch once on entry; handle loading, stale date, rate limits, missing winners, and network errors with an explicit retry action.
3. **Overlay UI:** Mount into a Shadow DOM root to limit CSS collisions. Render remote strings as text, not HTML. Add the seven answer rows, alternate answers, collapse, refresh, and a left/right position control. Keep focus and keyboard navigation accessible.
4. **Safari package:** Create an Xcode macOS containing app with the Safari Web Extension target. Its only required UI is setup guidance and extension status. Validate installation and website permission in Safari.
5. **Verification:** Compare extension results with `krillion --json` for the same date, including a round with multiple winners. Check first load, refresh, reset-day rollover, narrow windows, page navigation, site controls, and clean removal of the injected UI. Add fixtures for malformed or changed API responses.
6. **Optional later work:** Repackage the web extension for Chromium browsers. Consider a menu bar shortcut only if users want access outside the game tab.

## Release criteria

- Opening the daily game with the extension enabled shows all seven current winners without a click.
- Answers and alternates match the existing Go solver for the same puzzle date.
- The overlay can be hidden and does not block the game's primary controls at common window sizes.
- A failed or changed API response shows an error and retry, never stale answers as if they were current.
- The extension requests access only to `krillion.io` and makes no guess submissions.

## Open decisions before implementation

- Confirm Safari is the first browser to support. If Chrome is primary, build the same extension files there first and package for Safari afterward.
- Decide whether the collapsed state should persist across visits; the proposed default is **open on the first visit, then remember the user's choice**.
- Confirm whether the feature is strictly for the daily game; the current solver returns today's seven prompts only, so archive and unlimited modes are outside this first version.
