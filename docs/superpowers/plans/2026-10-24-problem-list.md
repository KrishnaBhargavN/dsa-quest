# Problem list and revision calendar: implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A "Next up" problem with description and hints, a full problem list he can add to, and a revision calendar.

**Architecture:** Pure link matching in `internal/game`; problems, status and revisions in `internal/store` (seeded from an embedded `problems.json`); thin handlers in `internal/api`; screens in `web/src/App.tsx`.

**Tech Stack:** Go 1.24+, lib/pq, Postgres 16 / Neon, React 19 + Vite.

**Spec:** `docs/superpowers/specs/2026-10-24-problem-list-design.md`

## Global Constraints

- No new scoring rules; "done before" earns 0 points.
- Every write goes through `Store.write`. Schema statements idempotent. No session state.
- Day math (revision `due`) uses `game.Day` (4 AM IST).
- Descriptions and hints are original text, never copied problem statements.
- Rebuild and commit `web/dist` after frontend changes.

## Review Focus

1. A LeetCode link pasted as `.../two-sum/description/?envType=...` or with capitals must match the list item.
2. Re-running the seed (every cold start) must not reset progress, revive removed problems, or duplicate rows.
3. A blind problem's topic must not appear anywhere in `/api/problems` or `/api/summary` until finished.
4. Getting stuck must advance Next up and put the problem on the redo list.
5. Finishing a problem with an open revision must complete it; finishing one without must not error.

---

### Task 1: Link matching

**Files:** Modify `internal/game/game.go`, `internal/game/game_test.go`

**Produces:** `func URLKey(raw string) string`

- [ ] Test `TestURLKey`: `https://leetcode.com/problems/Two-Sum/description/?envType=x` → `leetcode:two-sum`; `leetcode.com/problems/two-sum` → `leetcode:two-sum`; `https://cses.fi/problemset/task/1158/` → `cses:1158`; `https://cses.fi/problemset/result/1158` → lower-cased trimmed URL; `""` → `""`.
- [ ] Implement, run `go test ./internal/game/`, commit.

### Task 2: Problem content

**Files:** Create `internal/store/problems.json` (151 rows: the handoff fields plus `description` and `hints` [2]).

- [ ] Write descriptions and hints for all 151.
- [ ] Test `TestCatalogue` in `internal/store/catalogue_test.go`: parses; 151 rows; ids unique; `URLKey(url)` unique and non-empty; every row has a description and exactly 2 hints; difficulties valid; orders 1..151.

### Task 3: Store and API

**Files:** Modify `schema.sql`, `store.go`; create `internal/store/problems.go`, `internal/api/problems.go`, `internal/api/problems_test.go`; route in `api.go`.

**Produces (JSON):**
- `Problem {id, ord, phase, topic (omitted if hidden), title, url, difficulty, note, description, hints[], blind, custom, status: todo|solved|stuck|before, revisionDue?}`
- `GET /api/problems` → `{problems, nextId, redo: [id], done, total}`
- `POST /api/problems` `{title,url,difficulty,topic,description,hints,placement:"next"|"end"}`; `PUT/DELETE /api/problems/{id}`; `POST|DELETE /api/problems/{id}/before`
- `POST /api/attempts` accepts `problemId`
- `Summary.nextUp *Problem`, `Summary.revisionsDue []Revision`
- `Revision {id, problemId, url, title, difficulty, due "YYYY-MM-DD", doneAt}`; `GET /api/revisions?from=&to=`; `POST /api/revisions {problemId|url,title,difficulty, inDays|due}`; `PUT /api/revisions/{id} {due}`; `POST /api/revisions/{id}/done`; `DELETE /api/revisions/{id}`.

- [ ] Schema: `problems`, `attempts.problem_id`, `revisions`; seed upsert via `jsonb_to_recordset` in `Ensure` (content columns only on conflict, `WHERE NOT problems.custom`).
- [ ] Attempt start/log resolve `problem_id` by `url_key`; finishing completes open revisions with the same key scheduled before the attempt ended.
- [ ] Tests: seed idempotent and keeps progress/removal; next skips solved/stuck/before; pasted link matches; blind topic hidden until finished; redo rules (stuck, medium > 40 min, cleared by a fast re-solve); `next` placement; before = 0 points; revision scheduled, listed by range, due in summary, completed by attempt.
- [ ] `go test -count=1 ./...` with `TEST_DATABASE_URL`, commit.

### Task 4: Screens

**Files:** `web/src/api.ts`, `web/src/App.tsx`, `web/src/styles.css`, `web/dist`.

- [ ] Solve: revisions-due cards, Next up card (hints, done before, other problems), card kept while timer runs, manual form folded.
- [ ] After finishing: "Revise in 3 days · 1 week · 2 weeks · pick date · skip".
- [ ] List tab: month calendar with day detail (start, move, done, remove), progress, redo list, grouped list with expandable rows, own-problem editor.
- [ ] Home: Next up line and revisions-due line.
- [ ] `npm run build`, run locally, check phone width light/dark, commit.
