# Update

The **Update** pane asks GitHub if a newer tagged build exists. It does not check on startup.

1. **Check** reads [GitHub Releases](https://github.com/HardDie/ytmemchat_wails/releases) (`releases/latest`).
2. **Download** fetches this OS’s archive and checks SHA-256 against `SHA256SUMS.txt`.
3. **Quit and install** quits ytmemchat, replaces the binary or macOS `.app`, and starts the new build.

A `dev` or commit-stamped sidebar version is not treated as “behind” a tag. You can still download the latest archive.

If the install folder is not writable, use **Open GitHub** and replace the app yourself.

macOS builds are unsigned. A new `.app` may need [Running on macOS](Running-on-macOS) again.

Chat and overlay stop when the window quits.
