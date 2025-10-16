# G2 Password Tools
A small GUI password/passphrase generator &amp; string randomizer built with Go and Fyne.

Features
- Passphrase generator (wordlist-based)
- String randomizer (case changes, leet, separators)
- Export passphrases and copy to clipboard
- Embedded application icon (`icon.ico`) via go:embed. :contentReference[oaicite:3]{index=3}

## Files included
- `Main.go` — main application source. :contentReference[oaicite:4]{index=4}
- `icon.ico` — embedded icon used by the GUI.
- `wordlist.txt`, `leetmap.json` — default resources (created automatically if missing).
- `config.json` — optional saved configuration.

## Build
Make sure you have Go installed (1.20+ recommended) and `fyne` dependencies will be fetched automatically.

```bash
go mod init github.com/gegestrnad/g2-password-tools
go mod tidy
go build -ldflags "-H=windowsgui -s -w" -o g2-password-tools.exe Main.go
# Optional: compress with UPX
upx --best --lzma g2-password-tools.exe
