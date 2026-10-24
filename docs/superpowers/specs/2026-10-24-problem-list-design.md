# Problem list: design

Date: 24 Oct 2026. Status: waiting for Krishna's approval.

## Why

Krishna loses time and momentum searching for what to solve next, and being stuck is what makes him quit. DSA Quest should hand him the next problem on a plan built for his level (Claude Cowork's handoff, 151 problems, Oct 26 – Dec 20), help him get unstuck cheaply, and let him add his own problems.

Success: he opens the app, sees one problem, taps Start, and never has to search LeetCode for what to do.

## Goals

- **One "Next up" problem** by default: the first unfinished problem by `order`.
- **A way out when stuck:** two hints revealed one at a time (available during the timer too), and a "Show other problems" list.
- **Every problem has** a link (LeetCode or CSES), a 1–2 line description in our own words, and 2 hints that go from a nudge to the key idea without giving the full solution.
- **The full list** grouped by phase and topic, with progress marked.
- **Add, edit and remove his own problems** in the app, like the shop editor.
- **Blind problems** (the last 17) hide their topic until he finishes them.
- **Done before:** one tap marks a problem as already solved, with no points, so it drops out of Next up.
- **Redo list:** problems whose latest attempt ended stuck, or mediums that took over 40 minutes. This feeds Phase 4 (Dec 21–31).

## Non-goals and anti-goals

- **No new scoring rules.** Points stay exactly as they are; a "done before" tap earns nothing. More rules would put him off.
- **Being stuck must never feel punished.** An honest attempt still gives 5 points, and the problem moves to the redo list rather than blocking Next up.
- **No copying or fetching** of LeetCode or CSES problem statements. Descriptions and hints are written by us.
- **No phase dates enforced in the UI.** Dates are guidance; the list just carries forward in order.

## Constraints (from the existing code)

- All writes go through `Store.write` (advisory lock). Day/week math stays in Go. No session state (Neon pooler).
- Schema changes go in `schema.sql` and must be idempotent (`ADD COLUMN IF NOT EXISTS`, etc.).
- `web/dist` is committed; rebuild after frontend changes.
- Seeding runs on every cold start, so it must be one cheap statement.

## Design

### Data

New table `problems`:

| Column | Notes |
| --- | --- |
| `id TEXT PRIMARY KEY` | `lc-1462`, `cses-1648`, or `my-<n>` for his own |
| `ord DOUBLE PRECISION` | sort order; seeded from `order`. A float so a new problem can go between two others. |
| `phase INT NULL` | 1–3; NULL for his own problems |
| `topic`, `title`, `url`, `difficulty`, `note`, `description` | `TEXT`; `difficulty` checked like `attempts` |
| `hints TEXT[]` | 0–2 hints |
| `blind BOOLEAN` | |
| `url_key TEXT UNIQUE` | normalised link used for matching (see below) |
| `custom BOOLEAN` | true for problems he adds |
| `done_before_at TIMESTAMPTZ NULL` | set by the "done before" tap |
| `active BOOLEAN DEFAULT true` | soft delete, like rewards |

`attempts` gets `problem_id TEXT NULL REFERENCES problems(id)`.

**Matching links.** `url_key` is `leetcode:<slug>` for any leetcode.com/problems/<slug>/… link (ignoring `/description/`, query strings, trailing slashes and case) and `cses:<number>` for cses.fi/problemset/task/<n>; anything else is the trimmed lower-cased URL. When an attempt is started or logged, by tapping Start on a card or by pasting a link by hand, the server looks up the problem by `url_key` and sets `problem_id`. So a problem solved through the old "paste a link" flow still counts.

**Seeding.** The 151 problems, with descriptions and hints, live in `internal/store/problems.json` (embedded). On every cold start, one statement upserts them from `jsonb_to_recordset`: insert missing rows by `id`; for existing non-custom rows, update the content columns only (title, url, url_key, topic, phase, ord, difficulty, note, description, hints, blind). Progress lives in `attempts` and `done_before_at`, which the seed never touches, so fixing a typo in a hint later doesn't lose anything. A seeded problem he removed stays removed (`active` isn't touched).

### Status (computed, not stored)

For each active problem:
- **solved:** has a finished attempt with outcome `solved`.
- **done before:** `done_before_at` is set (and not solved).
- **stuck:** has finished attempts, none solved.
- **todo:** none of the above.

**Next up** is the first `todo` problem by `ord`. Stuck problems are not todo, so getting stuck moves him on to a fresh problem.

**Redo list:** problems whose most recent finished attempt was `stuck`, or a solved medium that took over 40 minutes (timer attempts only; manual logs carry minutes too and count). Solving it again cleanly takes it off.

**Blind:** the API leaves out `topic` for a blind problem until it is no longer `todo`. The UI shows "Mixed · topic hidden".

### API

- `GET /api/problems`: the full list with computed status, plus `nextId` and the redo list. About 151 rows; small enough to send in one go.
- `POST /api/problems`: add his own (title, url, difficulty, topic, description, up to 2 hints, `placement`: `next` or `end`). `next` puts it just before the current Next up.
- `PUT /api/problems/{id}`, `DELETE /api/problems/{id}`: edit or soft-delete. Only his own problems can be edited; any problem can be removed from the list.
- `POST /api/problems/{id}/before` and `DELETE` of the same: mark or unmark "done before".
- `POST /api/attempts` accepts an optional `problemId`; it fills in url, title and difficulty from the problem.
- `GET /api/summary` gains `nextUp` (the next problem, compact) so Home and Solve can show it without a second request.

### Screens (phone first, same look as today)

- **Solve tab:** a "Next up" card at the top: title, difficulty, topic (or "topic hidden"), Cowork's note if any, description, and a link to open the problem. Buttons: **Start timer** (uses the problem's link and difficulty), **Hint** (reveals hint 1, then hint 2), **Done before**, **Show other problems**. Below it, the existing manual form folds into "Solving something else? Paste a link".
- **Show other problems:** an inline list of the next 5 todo problems and the redo list. Tapping one swaps it into the card for this session; Next up itself doesn't change.
- **While the timer runs:** the running card keeps the title, link, description and hints, so hints are one tap away when stuck.
- **Home:** a one-line "Next up: <title>" that opens the Solve tab.
- **New "List" tab:** progress count (e.g. 12 / 151), the redo list, then problems grouped by phase and topic, each with a status mark (solved, stuck, done before, todo). Tap a row to expand its description and get Start / Done before. At the bottom, an editor to add his own problems, and edit or remove them.

### Content

I write a description (1–2 lines) and 2 hints for each of the 151 problems. Hint 1 points at how to think about it (the state, what to track, what's being searched); hint 2 names the key idea. For blind problems hint 1 avoids naming the pattern, since diagnosing the pattern is the point.

### Revision calendar (added 24 Oct)

Krishna's ask: when he solves a problem and feels he should solve it again in a week, he should see it on a calendar, for spaced revision.

- **Scheduling.** After finishing any problem attempt (solved or stuck, timer or manual log), the result step offers "Revise in: 3 days · 1 week · 2 weeks · pick a date", skippable. A problem row in the List tab also has "Schedule revision". Works for problems not on the list too (keyed by the attempt's url/title).
- **Data.** New table `revisions(id, problem_id NULL, url, title, difficulty, due DATE, created_at, done_at NULL)`. `due` is an app day (4 AM IST boundary), computed in Go.
- **Due revisions surface first.** On the Solve tab, revisions due today or overdue appear above Next up as "Revise" cards with Start (prefilled link and difficulty). Home shows "N revisions due today".
- **Completing one.** Finishing any attempt on the same `url_key` after the revision was scheduled marks its open revision done, even if that's earlier than the due day (doing it early still counts as revising). Then the result step offers to schedule the next one, so 3 days → 1 week → 2 weeks chains naturally. Revising earns the normal points for that attempt; no new rules.
- **Calendar view.** In the List tab, above the problems: a month grid (Mon–Sun, matching the good-week board) with a count on each day that has revisions, today highlighted, overdue marked. Tap a day to list its revisions; each can be started, moved to another date, or removed.

## Calls I made (Claude's, not Krishna's)

- The revision calendar lives in the List tab rather than a sixth tab.
- Solving a scheduled problem early completes its revision.
- A revision needs a `url_key` match to complete; a revision with no link is completed by tapping Done on it in the calendar.
- Revision offers come after every finished attempt, including stuck ones (re-trying a stuck problem in a few days is the best use of it).

- Being stuck moves the problem to the redo list instead of keeping it as Next up.
- The seed updates content on existing rows (the handoff suggested insert-only); progress is safe because it's stored elsewhere.
- Picking another problem from "Show other problems" doesn't reorder the list.
- Only his own problems are editable; seeded ones can be marked done before or removed.
- The topic for blind problems is hidden by the server, not just the UI.
- A fifth "List" tab rather than putting the full list on Home.

## Testing

- Rule tests for `url_key` normalisation.
- API tests: seed is idempotent and keeps progress; Next up skips solved, stuck and done-before; a pasted LeetCode link with `/description/` matches; blind topic hidden until finished; redo list rules; adding a problem at `next` makes it Next up; done before earns no points.
- Check every screen at phone width in light and dark mode, as v1 was.
