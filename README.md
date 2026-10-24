# DSA Quest

Points for DSA practice, spent on real rewards. Log a problem with a timer, earn points, buy gaming time or food in the shop, and unlock big rewards (a Steam game, a gaming mouse) by stacking good weeks.

One Go binary serves the API under `/api` and the React app for everything else. The app is prebuilt into `web/dist` and embedded, so deploying needs only Go. Tables are created automatically on first request.

## Deploy (about 10 minutes, no code)

1. **Push to GitHub.** In this folder:
   ```sh
   gh repo create dsa-quest --private --source=. --push
   ```
   No `gh`? Create an empty private repo named `dsa-quest` on github.com, then:
   ```sh
   git remote add origin https://github.com/<you>/dsa-quest.git
   git push -u origin main
   ```
2. **Vercel → Add New → Project →** import `dsa-quest`. Before clicking Deploy, open **Environment Variables** and add `APP_PASSWORD` with the password you'll log in with (not one you use anywhere else). Deploy.
3. **Project → Storage → Create Database → Neon.** Pick region **Singapore**, the **Free** plan, create it, then **Connect** it to this project for **Production only**, so preview builds never write real points. This sets `DATABASE_URL` (pooled) and `DATABASE_URL_UNPOOLED` (direct) for you.
4. **Deployments →** the latest one **→ ⋯ → Redeploy**, so the new database variable is picked up.
5. **Settings → Domains →** add `dsa.bhargavnkrishna.site`. In Cloudflare DNS add the CNAME Vercel shows (name `dsa`), proxy status **DNS only** (grey cloud). Wait for the certificate.

Check it: `https://dsa.bhargavnkrishna.site/api/health` should say `"database":"ok"` and `"passwordSet":true`. Then open the site on your phone, log in, and use **Add to Home Screen**.

Functions run in Singapore (`sin1`, set in `vercel.json`) next to the database. If Vercel's Go runtime (beta) ever misbehaves, the `Dockerfile` runs the same app on any container host.

## Rules

| Logged | Points |
| --- | --- |
| Easy / Medium / Hard solved | 5 / 15 / 30 |
| Medium solved in under 25 min | +10 |
| Honest attempt (stuck) | 5 |
| Chess project session | 10 |
| Good week (5 active days, Mon–Sun) | +50 |

A day runs 4 AM to 4 AM IST. Big rewards unlock at 3 and 6 good weeks, counted in total, not in a row. All numbers live in `internal/game/game.go`; shop prices and the mystery-box pool are edited in the app (Shop → Edit rewards and prices).

## Environment

| Variable | Needed | What |
| --- | --- | --- |
| `APP_PASSWORD` | yes | Your login password |
| `DATABASE_URL` | yes | Postgres URL; set by the Neon integration (pooled) |
| `PORT` | no | Set by Vercel; defaults to 8080 |

## Local development

```sh
# terminal 1: API (any Postgres works, e.g. a Neon dev branch)
DATABASE_URL=postgres://... APP_PASSWORD=dev go run .
# terminal 2: frontend with hot reload, proxies /api to :8080
cd web && npm install && npm run dev
```

After changing the frontend, rebuild it so the Go binary embeds the new version, then commit `web/dist`:

```sh
cd web && npm run build
```

Tests (the API tests need a throwaway database and wipe it):

```sh
go test ./internal/game/
TEST_DATABASE_URL=postgres://...test-db... go test ./...
```

## Backups

Neon's free plan keeps 6 hours of history. Now and then, run `pg_dump` against the direct connection string (`DATABASE_URL_UNPOOLED` in Vercel's env vars).
