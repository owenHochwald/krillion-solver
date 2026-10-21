# macOS menu bar app

With macOS 14 or later and Xcode installed, run this from the repository root:

```sh
make menu-bar
```

Click the shrimp in the menu bar to browse all seven rounds, see alternate top answers, and copy an answer while playing. The app has no Dock window.

`make menu-bar` builds `dist/KrillionBar.app`, closes any old running copy, and opens the new one. Use `make menu-bar-build` to build without opening it, or `open dist/KrillionBar.app` to reopen it. The menu bar icon is a small shrimp near the top-right of the screen.

This local build is unsigned and intended for building from source on your own Mac. Do not distribute it or ask users to disable Gatekeeper. A downloadable release needs Developer ID signing and Apple notarization.
