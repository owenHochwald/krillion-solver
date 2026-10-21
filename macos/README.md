# macOS menu bar app

With macOS 14 or later and Xcode installed, run this from the repository root:

```sh
make menu-bar
```

Click the shrimp in the menu bar to browse all seven rounds, see alternate top answers, and copy an answer while playing. The app has no Dock window.

`make menu-bar` builds and opens `dist/KrillionBar.app`. Use `make menu-bar-build` to build without opening it, or `open dist/KrillionBar.app` to reopen it.

This local build is unsigned. Distribution requires Developer ID signing and notarization.
