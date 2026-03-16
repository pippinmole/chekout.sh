# Part 2 — Desktop App (Go)

Located in `app/`. Compiles to a single binary. Registers `chekout://` with the OS, intercepts links, and opens the correct IDE.

## URL scheme registration

### macOS
- Bundle as a `.app` with `Info.plist` containing `CFBundleURLSchemes: [chekout]`
- Use `lsregister` to register on first run

### Windows
- Write to registry on first run:
  ```
  HKEY_CLASSES_ROOT\chekout
  HKEY_CLASSES_ROOT\chekout\shell\open\command → "path\to\chekout.exe" "%1"
  ```

### Linux
- Write a `.desktop` file to `~/.local/share/applications/chekout.desktop`
- Run `xdg-mime default chekout.desktop x-scheme-handler/chekout`

## Config file

Path: `~/.config/chekout/config.json`

```json
{
  "default_ide": "vscode",
  "repos": {
    "github.com/org/repo": {
      "path": "/Users/jonny/code/repo",
      "ide": "vscode"
    }
  }
}
```

Supported IDE values: `vscode`, `cursor`, `intellij`, `webstorm`, `goland`, `phpstorm`

## IDE deep links

| IDE | Scheme |
|-----|--------|
| VS Code | `vscode://file/{path}` |
| Cursor | `cursor://file/{path}` |
| IntelliJ | `jetbrains://idea/open?file={path}` |
| WebStorm | `jetbrains://web-storm/open?file={path}` |
| GoLand | `jetbrains://go-land/open?file={path}` |

## Handler flow

1. Parse `chekout://open?repo=...&branch=...`
2. Look up repo in config — if missing, show native OS dialog to locate folder, then save to config
3. Optionally read `.chekout.json` from the repo root for a team-level IDE override
4. Run `git fetch origin {branch} && git checkout {branch}` via `os/exec`
5. Fire IDE deep link via `os/exec open {ide-scheme}`
6. Show OS notification on success or failure via `gen2brain/beeep`

## Auto-discovery on first run

Walk these directories looking for `.git/config` files to pre-populate the repo map:
- `~/code`, `~/projects`, `~/dev`, `~/workspace`, `~/src`

Parse `[remote "origin"]` from each `.git/config` to extract the GitHub repo name.

## Tray UI

- Icon in system tray (use a simple monochrome icon, 22x22px on macOS)
- Dropdown menu items:
  - Recent opens (last 5)
  - Open config file
  - Quit

## Dependencies

```
github.com/getlantern/systray   # system tray
github.com/gen2brain/beeep      # OS notifications
```

## Distribution

- `goreleaser` config at `app/.goreleaser.yml`
- Builds: `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `linux/amd64`
- Homebrew tap: `brew install chekout-sh/tap/chekout`
- Winget manifest in `app/winget/`
