# Running on macOS

Release builds of ytmemchat are **not code-signed**. macOS Gatekeeper blocks unsigned apps downloaded from the internet. That is expected. You do not need a developer certificate to run the app on your own Mac.

The macOS archive is `ytmemchat-vMAJOR.MINOR.PATCH-darwin-universal.zip` (Intel and Apple Silicon). Inside it is `ytmemchat.app`.

## 1. Unzip, then open

1. Download the zip from [GitHub Releases](https://github.com/HardDie/ytmemchat_wails/releases).
2. Double-click the zip so Finder extracts `ytmemchat.app`. Do not launch the app from inside the zip window.
3. Optional: drag `ytmemchat.app` to **Applications**.

## 2. Control-click → Open

The first time, do **not** double-click.

1. Control-click (or right-click) `ytmemchat.app`.
2. Choose **Open**.
3. In the dialog, click **Open** again.

macOS remembers that choice for this copy of the app. After that, double-click works. Apple’s steps: [Open a Mac app from an unidentified developer](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac).

## 3. If macOS still blocks it

You may see *“Apple cannot check it for malicious software”* or *“the developer cannot be verified.”*

1. Open **System Settings → Privacy & Security**.
2. Scroll to **Security**.
3. Next to the message about ytmemchat, click **Open Anyway**.
4. Confirm with your password or Touch ID.

Apple’s steps: [If you want to open an app that hasn’t been notarized](https://support.apple.com/102445).

## Terminal (optional)

If the dialogs never appear (for example you copied the app from Downloads), strip the quarantine flag and open it:

```bash
xattr -dr com.apple.quarantine /path/to/ytmemchat.app
open /path/to/ytmemchat.app
```

Use the real path (Downloads or Applications). Do not disable Gatekeeper system-wide.

## Built on this Mac

`make build` writes `build/bin/ytmemchat.app`. A binary you compiled locally is usually **not** quarantined, so double-click is enough. If you zip it, send it, then download it again, treat it like a release and use Control-click → Open.

Then continue with [Getting Started](Getting-Started).
