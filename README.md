# G2 Password Tools

A small GUI password/passphrase generator and string randomizer built with Go and Fyne.

<p align="center">
  <img src="https://imgur.com/JlDf4pp.jpg" alt="Password Tools screenshot" width="400">
</p>

## Features

- Password generator
- Passphrase generator using a wordlist
- String/phrase randomizer with case changes, leet, and separators
- Export passphrases and copy generated text to the clipboard

## Files Included

- `Main.go` - main application source.
- `icon.ico` - embedded icon used by the GUI.
- `wordlist.txt`, `leetmap.json` - default resources.
- `config.json` - saved application defaults.

## Build

Make sure Go is installed. The module is already initialized, so fetch dependencies and build from the project root:

```bash
go mod tidy
go build -trimpath -ldflags "-H=windowsgui -s -w" -o g2-password-tools.exe .
```

For development validation:

```bash
go test ./...
go vet ./...
```

## Release

GitHub Actions builds a UPX-compressed Windows executable and publishes it to GitHub Releases when a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

You can also run the `Build Release` workflow manually and provide `release_tag`, such as `v1.0.0`, to publish a release. Manual runs without `release_tag` upload the build as a workflow artifact instead.

Release assets:

- `g2-password-tools-windows-amd64.exe`
- `g2-password-tools-windows-amd64.exe.sha256`
