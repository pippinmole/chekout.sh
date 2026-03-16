# Part 1 — GitHub Action

Triggered on `pull_request: [opened, synchronize]`. Posts (or updates) a bot comment on the PR with a chekout.sh link.

## File location

`action/chekout.yml` — published to the GitHub Marketplace so any repo can use it in one line.

## Steps

1. Extract `context.payload.pull_request.head.ref` (branch) and `context.payload.repository.full_name` (repo)
2. Construct URL: `https://chekout.sh/open?repo={full_name}&branch={branch}`
3. Check if the bot has already commented on this PR — search for the fingerprint `<!-- chekout-bot -->` in existing comments
4. If found: edit the existing comment via `github.rest.issues.updateComment`
5. If not found: create a new comment via `github.rest.issues.createComment`

## Comment format

```markdown
<!-- chekout-bot -->
**Open this PR locally** — [chekout.sh](https://chekout.sh/open?repo=org/repo&branch=feat/branch)

> Install the [chekout.sh desktop app](https://chekout.sh) if you haven't already.
```

## Permissions

The workflow needs `pull-requests: write` permission to post comments.

## Marketplace publish

- `action.yml` at repo root with `name`, `description`, `branding`, and `inputs`
- Tag a release to trigger marketplace listing
