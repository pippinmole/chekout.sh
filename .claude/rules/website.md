# Part 3 — Website (Next.js)

Located in `web/`. Deployed to Vercel at `chekout.sh`.

## Routes

### `/open` — redirect page

The page that GitHub Action links point to. Immediately fires the `chekout://` deep link.

**Behaviour:**
1. On mount, read `repo` and `branch` from query params
2. Immediately set `window.location.href = chekout://open?repo={repo}&branch={branch}`
3. Start a `setTimeout` of 2500ms
4. If the page is still visible after 2500ms, the app is not installed — show install CTA
5. The install CTA should preserve the original URL params so after install the link works immediately

**Implementation notes:**
- Use `useEffect` with `window.location.href` — do not use Next.js `router.push` for custom schemes
- Add `<meta name="robots" content="noindex">` — this page should not be indexed
- Keep this page extremely fast — no heavy JS, no layout shift

### `/` — landing page

Marketing page. Should clearly explain:
- What the problem is (manual checkout friction)
- How chekout.sh solves it
- Two install steps: add the GitHub Action + install the desktop app
- Supported IDEs
- Link to GitHub repo

### `/docs` — setup guide

Step-by-step:
1. Add the Action to your repo (copy-paste workflow YAML)
2. Install the desktop app (Homebrew / direct download)
3. Configure your IDE preference

## Tech

- Next.js App Router
- TypeScript
- Tailwind CSS
- Deployed on Vercel (zero config)

## Environment

No environment variables needed — the site is fully static / client-side. No backend, no database.

## Domain

Point `chekout.sh` DNS A/CNAME records to Vercel. Configure in Vercel dashboard under project settings > domains.
