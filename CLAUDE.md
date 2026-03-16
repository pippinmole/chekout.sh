# chekout.sh

A tool that generates a per-PR deep link in GitHub comments. When clicked, the link opens the repo in the user's local IDE at the correct branch — no manual checkout required.

## Architecture

Three independent parts, built in this order:

1. **GitHub Action** — posts a `chekout://` deep link as a bot comment on every PR
2. **Desktop app (Go)** — registers the `chekout://` URL scheme, looks up local clone path, fires the IDE
3. **Website (Next.js)** — redirect layer at `chekout.sh/open` that bridges the GitHub link to the custom scheme

## URL format

```
https://chekout.sh/open?repo=github.com/org/repo&branch=feat/my-branch
```

This hits the Next.js site, which immediately fires:

```
chekout://open?repo=github.com/org/repo&branch=feat/my-branch
```

Which the desktop app intercepts, looks up the local clone, checks out the branch, and opens the IDE.

## Repo structure

```
/
├── action/        # GitHub Action (YAML + JS)
├── app/           # Go desktop application
└── web/           # Next.js website
```

## Key decisions

- **Go** for the desktop app — single binary, low memory, easy URL scheme registration on all platforms
- **No Electron** — use `getlantern/systray` for tray, `gen2brain/beeep` for notifications
- **Config** stored at `~/.config/chekout/config.json` as a map of `github.com/org/repo → /local/path`
- **IDE preference** stored per-repo in config, falling back to a global default; can also be set via `.chekout.json` in the repo root
- **Comment upsert** — the Action edits its existing comment on re-push rather than posting a new one (fingerprinted with a hidden HTML comment)
- **Fallback detection** — the website uses a 2500ms timeout to detect if the app is not installed and shows an install prompt

## Tech stack

| Part | Stack |
|------|-------|
| GitHub Action | YAML, `actions/github-script` |
| Desktop app | Go, `getlantern/systray`, `gen2brain/beeep`, `goreleaser` |
| Website | Next.js, TypeScript, Tailwind, Vercel |

## Detailed plans

See `.claude/rules/` for part-by-part implementation detail:

- `.claude/rules/github-action.md`
- `.claude/rules/desktop-app.md`
- `.claude/rules/website.md`
