-- Runs on every cold start inside one transaction behind an advisory lock.
-- Every statement must be safe to run again.

CREATE TABLE IF NOT EXISTS settings (
  key    TEXT PRIMARY KEY,
  value  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS attempts (
  id          BIGSERIAL PRIMARY KEY,
  kind        TEXT NOT NULL CHECK (kind IN ('problem', 'chess')),
  url         TEXT,
  title       TEXT,
  difficulty  TEXT CHECK (difficulty IN ('easy', 'medium', 'hard')),
  started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at    TIMESTAMPTZ,
  outcome     TEXT CHECK (outcome IN ('solved', 'stuck')),
  CHECK (kind = 'chess' OR difficulty IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS attempts_started_idx ON attempts (started_at);

CREATE TABLE IF NOT EXISTS rewards (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  cost            INT CHECK (cost IS NULL OR cost > 0),  -- NULL = mystery pool only
  mystery_weight  INT NOT NULL DEFAULT 0 CHECK (mystery_weight >= 0),
  active          BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS ledger (
  id          BIGSERIAL PRIMARY KEY,
  at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  delta       INT NOT NULL,
  reason      TEXT NOT NULL,
  attempt_id  BIGINT REFERENCES attempts(id) ON DELETE CASCADE,
  reward_id   BIGINT REFERENCES rewards(id),
  week_start  DATE,
  UNIQUE (reason, week_start)
);
CREATE INDEX IF NOT EXISTS ledger_attempt_idx ON ledger (attempt_id);

CREATE TABLE IF NOT EXISTS milestones (
  id                 BIGSERIAL PRIMARY KEY,
  name               TEXT NOT NULL,
  good_weeks_needed  INT NOT NULL CHECK (good_weeks_needed > 0),
  claimed_at         TIMESTAMPTZ
);

INSERT INTO rewards (name, cost, mystery_weight)
SELECT * FROM (VALUES
  ('30 min of gaming',      30,            40),
  ('1 hour of gaming',      50,            15),
  ('Favourite food order',  120,           0),
  ('Big treat',             400,           0),
  ('Dessert',               NULL::int,     40),
  ('2-hour gaming jackpot', NULL::int,     5)
) AS seed (name, cost, mystery_weight)
WHERE NOT EXISTS (SELECT 1 FROM rewards);

INSERT INTO milestones (name, good_weeks_needed)
SELECT * FROM (VALUES ('Steam game', 3), ('Gaming mouse', 6)) AS seed (name, good_weeks_needed)
WHERE NOT EXISTS (SELECT 1 FROM milestones);

-- Problem list. Seeded rows are upserted from problems.json on every cold
-- start (see seedProblems). Progress lives in attempts (matched by url_key)
-- and done_before_at, which the seed never touches.
CREATE TABLE IF NOT EXISTS problems (
  id              TEXT PRIMARY KEY,
  ord             DOUBLE PRECISION NOT NULL,
  phase           INT,
  topic           TEXT NOT NULL DEFAULT '',
  title           TEXT NOT NULL,
  url             TEXT NOT NULL DEFAULT '',
  url_key         TEXT,
  difficulty      TEXT NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
  note            TEXT NOT NULL DEFAULT '',
  description     TEXT NOT NULL DEFAULT '',
  hints           TEXT[] NOT NULL DEFAULT '{}',
  blind           BOOLEAN NOT NULL DEFAULT false,
  custom          BOOLEAN NOT NULL DEFAULT false,
  done_before_at  TIMESTAMPTZ,
  active          BOOLEAN NOT NULL DEFAULT true
);
CREATE UNIQUE INDEX IF NOT EXISTS problems_url_key_idx ON problems (url_key) WHERE url_key IS NOT NULL AND active;

ALTER TABLE attempts ADD COLUMN IF NOT EXISTS url_key TEXT;
CREATE INDEX IF NOT EXISTS attempts_url_key_idx ON attempts (url_key);

-- Revision calendar: "solve this again on <due>".
CREATE TABLE IF NOT EXISTS revisions (
  id          BIGSERIAL PRIMARY KEY,
  url         TEXT NOT NULL DEFAULT '',
  url_key     TEXT,
  title       TEXT NOT NULL,
  difficulty  TEXT NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
  due         DATE NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  done_at     TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS revisions_due_idx ON revisions (due) WHERE done_at IS NULL;
