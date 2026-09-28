# Installing on Linux

Linux releases are `ytmemchat-vMAJOR.MINOR.PATCH-linux-amd64.tar.gz` and `linux-arm64.tar.gz`.

Download one from [GitHub Releases](https://github.com/HardDie/ytmemchat_wails/releases). Unpack it.

The folder contains `ytmemchat`, `ytmemchat.png`, `ytmemchat.desktop`, and `install.sh`.

The app needs GTK 3 and WebKitGTK 4.1. On Ubuntu 24.04:

```bash
sudo apt install libgtk-3-0t64 libwebkit2gtk-4.1-0
```

Pick one option below.

## 1. Run the binary

Start `ytmemchat` from the unpacked folder.

It stays a file. It does not show in the app list.

```bash
chmod +x ytmemchat
./ytmemchat
```

## 2. Install for this user

Run `install.sh` from that same folder.

The app is installed only for your user. Other accounts on the machine do not get it.

No administrator password. Ubuntu, GNOME, and KDE show it in the app list.

```bash
chmod +x install.sh
./install.sh
```

The script copies:

- `~/.local/bin/ytmemchat`
- `~/.local/share/applications/ytmemchat.desktop`
- `~/.local/share/icons/hicolor/512x512/apps/ytmemchat.png`

Open ytmemchat from the app list after that.

In-app update replaces `~/.local/bin/ytmemchat`. The launcher and icon stay.

Then continue with [Getting Started](Getting-Started).
