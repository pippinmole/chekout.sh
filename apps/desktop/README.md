# chekout — desktop app

The Go binary that handles `chekout://` deep links. Runs in the system tray, intercepts links from the browser, checks out the branch, and opens your IDE.

## Development

### Prerequisites

- Go 1.21+
- macOS: Xcode Command Line Tools (`xcode-select --install`) — required for CGO (Apple Events)

### Run

```bash
cd app
go run .
```

On first run the app will:
1. Walk `~/code`, `~/projects`, `~/dev`, `~/workspace`, and `~/src` to auto-populate known repos
2. Attempt to register the `chekout://` URL scheme

### URL scheme registration warning

When running as a plain binary (i.e. `go run .` or a bare `./chekout`), you'll see:

```
warning: failed to register URL scheme: could not locate .app bundle — skipping lsregister
```

This is expected. macOS URL scheme registration requires a `.app` bundle. To actually register the scheme during development, wrap the binary first:

```bash
go build -o chekout .

mkdir -p chekout.app/Contents/MacOS
cp chekout chekout.app/Contents/MacOS/
cp Info.plist chekout.app/Contents/

/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister -f $(pwd)/chekout.app

open chekout.app
```

After registration, `chekout://` links will open this app even when launched from the browser.

## Config

Config lives at `~/.config/chekout/config.json`:

```json
{
  "default_ide": "vscode",
  "repos": {
    "github.com/org/repo": {
      "path": "/Users/you/code/repo",
      "ide": "cursor"
    }
  }
}
```

Supported `ide` values: `vscode`, `cursor`, `intellij`, `webstorm`, `goland`, `phpstorm`

A per-repo override can also be committed at `.chekout.json` in the repo root:

```json
{ "ide": "goland" }
```

## Building for release

Releases use [goreleaser](https://goreleaser.com):

```bash
goreleaser release --clean
```

Builds produced: `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `linux/amd64`. Darwin builds require CGO (`CC=clang`) and must run on a Mac host.

The `.app` bundle for macOS is assembled from the binary + `Info.plist` as a post-build step in the release workflow.
