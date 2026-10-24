// Package store is the only code that talks to Postgres.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"dsaquest/internal/game"
)

//go:embed schema.sql
var schemaSQL string

// Errors the API turns into status codes.
var (
	ErrNotFound     = errors.New("not found")
	ErrTimerRunning = errors.New("a problem timer is already running")
	ErrFinished     = errors.New("that attempt is already finished")
	ErrNotEnough    = errors.New("not enough points")
	ErrLocked       = errors.New("not unlocked yet")
	ErrClaimed      = errors.New("already claimed")
	ErrEmptyMystery = errors.New("the mystery pool is empty")
	ErrBadInput     = errors.New("bad input")
	ErrNotForSale   = errors.New("that reward is mystery-box only")
)

// lockKey serializes every write, so a double tap on the phone can't
// double-spend or double-award. It is transaction-scoped, which is safe
// through Neon's transaction-mode pooler.
const lockKey = 7331

type Store struct {
	db     *sql.DB
	mu     sync.Mutex
	ready  bool
	secret []byte
}

// Open prepares a connection pool. It does not connect yet.
func Open(dsn string) (*Store, error) {
	clean, err := cleanDSN(dsn)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", clean)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(2 * time.Minute)
	return &Store{db: db}, nil
}

// cleanDSN drops URL parameters lib/pq would forward to the server as
// unknown settings, and gives a sleeping Neon compute time to wake up.
func cleanDSN(dsn string) (string, error) {
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return dsn, nil
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("DATABASE_URL is not a valid URL: %w", err)
	}
	q := u.Query()
	for _, k := range []string{"channel_binding", "pgbouncer", "schema", "connection_limit", "pool_timeout"} {
		q.Del(k)
	}
	if q.Get("connect_timeout") == "" {
		q.Set("connect_timeout", "15")
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Ensure creates the tables on first use and loads the session secret.
// Safe to call on every request; it only does work once per process.
func (s *Store) Ensure(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ready {
		return nil
	}
	err := s.write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, schemaSQL); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
		if err := seedProblems(ctx, tx); err != nil {
			return fmt.Errorf("seed problems: %w", err)
		}
		if err := backfillURLKeys(ctx, tx); err != nil {
			return fmt.Errorf("backfill url keys: %w", err)
		}
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value) VALUES ('session_secret', $1) ON CONFLICT (key) DO NOTHING`,
			hex.EncodeToString(buf)); err != nil {
			return err
		}
		var secretHex string
		if err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = 'session_secret'`).Scan(&secretHex); err != nil {
			return err
		}
		secret, err := hex.DecodeString(secretHex)
		if err != nil {
			return err
		}
		s.secret = secret
		return nil
	})
	if err != nil {
		return err
	}
	s.ready = true
	return nil
}

// Secret is the key that signs login cookies. Call Ensure first.
func (s *Store) Secret() []byte { return s.secret }

func (s *Store) write(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockKey); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ---------- types the API returns ----------

type Attempt struct {
	ID         int64      `json:"id"`
	Kind       string     `json:"kind"`
	URL        string     `json:"url"`
	Title      string     `json:"title"`
	Difficulty string     `json:"difficulty"`
	StartedAt  time.Time  `json:"startedAt"`
	EndedAt    *time.Time `json:"endedAt"`
	Outcome    string     `json:"outcome"`
	Minutes    *float64   `json:"minutes"`
	Points     int        `json:"points"`
	urlKey     string
}

type Result struct {
	Awards  []game.Award `json:"awards"`
	Earned  int          `json:"earned"`
	Balance int          `json:"balance"`
}

type Reward struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Cost          *int   `json:"cost"`
	MysteryWeight int    `json:"mysteryWeight"`
}

type Milestone struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Needed    int        `json:"needed"`
	ClaimedAt *time.Time `json:"claimedAt"`
	Unlocked  bool       `json:"unlocked"`
}

type DayCell struct {
	Date    string `json:"date"`
	Active  bool   `json:"active"`
	IsToday bool   `json:"isToday"`
	Future  bool   `json:"future"`
}

type Week struct {
	Start          string    `json:"start"`
	Days           []DayCell `json:"days"`
	ActiveCount    int       `json:"activeCount"`
	Needed         int       `json:"needed"`
	SkipPassesLeft int       `json:"skipPassesLeft"`
	GoodWeek       bool      `json:"goodWeek"`
}

type NextReward struct {
	Name string `json:"name"`
	Cost int    `json:"cost"`
	Need int    `json:"need"`
}

type Summary struct {
	Balance     int         `json:"balance"`
	TodayPoints int         `json:"todayPoints"`
	Week        Week        `json:"week"`
	GoodWeeks   int         `json:"goodWeeks"`
	Milestones  []Milestone `json:"milestones"`
	NextReward  *NextReward `json:"nextReward"`
	Running     *Attempt    `json:"running"`
	// RunningProblem is the list problem the running timer is for, if any,
	// so the timer screen can show its description and hints.
	RunningProblem *Problem   `json:"runningProblem"`
	NextUp         *Problem   `json:"nextUp"`
	RevisionsDue   []Revision `json:"revisionsDue"`
}

type Redemption struct {
	At     time.Time `json:"at"`
	Name   string    `json:"name"`
	Spent  int       `json:"spent"`
	Reason string    `json:"reason"`
}

type WeekStat struct {
	WeekStart     string   `json:"weekStart"`
	Solved        int      `json:"solved"`
	MedianMinutes *float64 `json:"medianMinutes"`
}

// ---------- reads ----------

func balance(ctx context.Context, q queryer) (int, error) {
	var b int
	err := q.QueryRowContext(ctx, `SELECT COALESCE(SUM(delta), 0) FROM ledger`).Scan(&b)
	return b, err
}

func activeTimes(ctx context.Context, q queryer, from, to time.Time) ([]time.Time, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT started_at FROM attempts WHERE ended_at IS NOT NULL AND started_at >= $1 AND started_at < $2`,
		from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) Summary(ctx context.Context, now time.Time) (*Summary, error) {
	sum := &Summary{}
	var err error
	if sum.Balance, err = balance(ctx, s.db); err != nil {
		return nil, err
	}

	today := game.Day(now)
	dayFrom, dayTo := game.DayBounds(today)
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(delta), 0) FROM ledger WHERE delta > 0 AND at >= $1 AND at < $2`,
		dayFrom, dayTo).Scan(&sum.TodayPoints); err != nil {
		return nil, err
	}

	ws := game.WeekStart(today)
	from, to := game.WeekBounds(ws)
	times, err := activeTimes(ctx, s.db, from, to)
	if err != nil {
		return nil, err
	}
	active := game.ActiveDays(times)
	sum.Week = Week{
		Start:          ws.Format("2006-01-02"),
		ActiveCount:    len(active),
		Needed:         game.GoodWeekDays,
		SkipPassesLeft: game.SkipPassesLeft(ws, today, active),
		GoodWeek:       game.IsGoodWeek(len(active)),
	}
	for i := 0; i < 7; i++ {
		d := ws.AddDate(0, 0, i)
		sum.Week.Days = append(sum.Week.Days, DayCell{
			Date:    d.Format("2006-01-02"),
			Active:  active[d],
			IsToday: d.Equal(today),
			Future:  d.After(today),
		})
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM ledger WHERE reason = $1`, game.ReasonGoodWeek).Scan(&sum.GoodWeeks); err != nil {
		return nil, err
	}
	if sum.Milestones, err = milestones(ctx, s.db, sum.GoodWeeks); err != nil {
		return nil, err
	}

	var name string
	var cost int
	err = s.db.QueryRowContext(ctx,
		`SELECT name, cost FROM rewards WHERE active AND cost IS NOT NULL AND cost > $1 ORDER BY cost, id LIMIT 1`,
		sum.Balance).Scan(&name, &cost)
	switch {
	case err == nil:
		sum.NextReward = &NextReward{Name: name, Cost: cost, Need: cost - sum.Balance}
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	if sum.Running, err = runningAttempt(ctx, s.db); err != nil {
		return nil, err
	}
	list, err := loadProblems(ctx, s.db)
	if err != nil {
		return nil, err
	}
	sum.NextUp = nextUp(list)
	if sum.Running != nil && sum.Running.urlKey != "" {
		for i := range list.Problems {
			if list.Problems[i].urlKey == sum.Running.urlKey {
				sum.RunningProblem = &list.Problems[i]
				break
			}
		}
	}
	if sum.RevisionsDue, err = dueRevisions(ctx, s.db, today); err != nil {
		return nil, err
	}
	return sum, nil
}

func milestones(ctx context.Context, q queryer, goodWeeks int) ([]Milestone, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, name, good_weeks_needed, claimed_at FROM milestones ORDER BY good_weeks_needed, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Milestone{}
	for rows.Next() {
		var m Milestone
		if err := rows.Scan(&m.ID, &m.Name, &m.Needed, &m.ClaimedAt); err != nil {
			return nil, err
		}
		m.Unlocked = goodWeeks >= m.Needed
		out = append(out, m)
	}
	return out, rows.Err()
}

const attemptCols = `a.id, a.kind, COALESCE(a.url, ''), COALESCE(a.title, ''), COALESCE(a.difficulty, ''),
	a.started_at, a.ended_at, COALESCE(a.outcome, ''),
	COALESCE((SELECT SUM(l.delta) FROM ledger l WHERE l.attempt_id = a.id), 0), COALESCE(a.url_key, '')`

func scanAttempt(sc interface{ Scan(...any) error }) (*Attempt, error) {
	var a Attempt
	if err := sc.Scan(&a.ID, &a.Kind, &a.URL, &a.Title, &a.Difficulty,
		&a.StartedAt, &a.EndedAt, &a.Outcome, &a.Points, &a.urlKey); err != nil {
		return nil, err
	}
	if a.EndedAt != nil && a.Kind == "problem" {
		m := a.EndedAt.Sub(a.StartedAt).Minutes()
		a.Minutes = &m
	}
	return &a, nil
}

func runningAttempt(ctx context.Context, q queryer) (*Attempt, error) {
	a, err := scanAttempt(q.QueryRowContext(ctx,
		`SELECT `+attemptCols+` FROM attempts a
		 WHERE a.kind = 'problem' AND a.ended_at IS NULL ORDER BY a.started_at DESC LIMIT 1`))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (s *Store) History(ctx context.Context) ([]*Attempt, []Redemption, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+attemptCols+` FROM attempts a WHERE a.ended_at IS NOT NULL ORDER BY a.started_at DESC LIMIT 60`)
	if err != nil {
		return nil, nil, err
	}
	attempts := []*Attempt{}
	for rows.Next() {
		a, err := scanAttempt(rows)
		if err != nil {
			rows.Close()
			return nil, nil, err
		}
		attempts = append(attempts, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	rrows, err := s.db.QueryContext(ctx,
		`SELECT l.at, COALESCE(r.name, ''), -l.delta, l.reason FROM ledger l
		 LEFT JOIN rewards r ON r.id = l.reward_id
		 WHERE l.reason IN ($1, $2) ORDER BY l.at DESC LIMIT 30`,
		game.ReasonBuy, game.ReasonMystery)
	if err != nil {
		return nil, nil, err
	}
	defer rrows.Close()
	redemptions := []Redemption{}
	for rrows.Next() {
		var r Redemption
		if err := rrows.Scan(&r.At, &r.Name, &r.Spent, &r.Reason); err != nil {
			return nil, nil, err
		}
		redemptions = append(redemptions, r)
	}
	return attempts, redemptions, rrows.Err()
}

// MediumStats returns the median time of solved mediums for each of the last 8 weeks.
func (s *Store) MediumStats(ctx context.Context, now time.Time) ([]WeekStat, error) {
	const weeks = 8
	first := game.WeekStart(game.Day(now)).AddDate(0, 0, -7*(weeks-1))
	from, _ := game.WeekBounds(first)
	rows, err := s.db.QueryContext(ctx,
		`SELECT started_at, ended_at FROM attempts
		 WHERE kind = 'problem' AND difficulty = 'medium' AND outcome = 'solved' AND started_at >= $1`, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byWeek := map[time.Time][]float64{}
	for rows.Next() {
		var st, en time.Time
		if err := rows.Scan(&st, &en); err != nil {
			return nil, err
		}
		ws := game.WeekStart(game.Day(st))
		byWeek[ws] = append(byWeek[ws], en.Sub(st).Minutes())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]WeekStat, 0, weeks)
	for i := 0; i < weeks; i++ {
		ws := first.AddDate(0, 0, 7*i)
		vals := byWeek[ws]
		st := WeekStat{WeekStart: ws.Format("2006-01-02"), Solved: len(vals)}
		if len(vals) > 0 {
			m := game.Median(vals)
			st.MedianMinutes = &m
		}
		out = append(out, st)
	}
	return out, nil
}

func (s *Store) Rewards(ctx context.Context) ([]Reward, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, cost, mystery_weight FROM rewards WHERE active ORDER BY cost NULLS LAST, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Reward{}
	for rows.Next() {
		var r Reward
		if err := rows.Scan(&r.ID, &r.Name, &r.Cost, &r.MysteryWeight); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------- writes ----------

func addAwards(ctx context.Context, tx *sql.Tx, attemptID int64, awards []game.Award, now time.Time) error {
	for _, a := range awards {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ledger (at, delta, reason, attempt_id) VALUES ($1, $2, $3, $4)`,
			now, a.Points, a.Reason, attemptID); err != nil {
			return err
		}
	}
	return nil
}

// checkGoodWeek awards the weekly bonus once the week reaches enough active
// days. The UNIQUE (reason, week_start) constraint makes it impossible to pay twice.
func checkGoodWeek(ctx context.Context, tx *sql.Tx, activityAt, now time.Time) (*game.Award, error) {
	ws := game.WeekStart(game.Day(activityAt))
	from, to := game.WeekBounds(ws)
	times, err := activeTimes(ctx, tx, from, to)
	if err != nil {
		return nil, err
	}
	if !game.IsGoodWeek(len(game.ActiveDays(times))) {
		return nil, nil
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO ledger (at, delta, reason, week_start) VALUES ($1, $2, $3, $4::date)
		 ON CONFLICT (reason, week_start) DO NOTHING`,
		now, game.GoodWeekBonus, game.ReasonGoodWeek, ws.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 1 {
		return &game.Award{Reason: game.ReasonGoodWeek, Points: game.GoodWeekBonus}, nil
	}
	return nil, nil
}

func finishResult(ctx context.Context, tx *sql.Tx, attemptID int64, awards []game.Award, activityAt, now time.Time) (*Result, error) {
	if err := addAwards(ctx, tx, attemptID, awards, now); err != nil {
		return nil, err
	}
	bonus, err := checkGoodWeek(ctx, tx, activityAt, now)
	if err != nil {
		return nil, err
	}
	if bonus != nil {
		awards = append(awards, *bonus)
	}
	r := &Result{Awards: awards}
	for _, a := range awards {
		r.Earned += a.Points
	}
	if r.Balance, err = balance(ctx, tx); err != nil {
		return nil, err
	}
	return r, nil
}

// StartAttempt starts the timer. With a problemID, the link, title and
// difficulty come from the list; otherwise from the arguments.
func (s *Store) StartAttempt(ctx context.Context, problemID, url, title, difficulty string, now time.Time) (*Attempt, error) {
	if problemID == "" && !game.ValidDifficulty(difficulty) {
		return nil, ErrBadInput
	}
	var out *Attempt
	err := s.write(ctx, func(tx *sql.Tx) error {
		ref, err := resolveRef(ctx, tx, problemID, url, title, difficulty)
		if err != nil {
			return err
		}
		running, err := runningAttempt(ctx, tx)
		if err != nil {
			return err
		}
		if running != nil {
			return ErrTimerRunning
		}
		var id int64
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO attempts (kind, url, title, difficulty, started_at, url_key)
			 VALUES ('problem', NULLIF($1, ''), NULLIF($2, ''), $3, $4, NULLIF($5, '')) RETURNING id`,
			ref.url, ref.title, ref.difficulty, now, ref.urlKey).Scan(&id); err != nil {
			return err
		}
		out, err = scanAttempt(tx.QueryRowContext(ctx, `SELECT `+attemptCols+` FROM attempts a WHERE a.id = $1`, id))
		return err
	})
	return out, err
}

func (s *Store) FinishAttempt(ctx context.Context, id int64, outcome string, now time.Time) (*Result, error) {
	if outcome != game.Solved && outcome != game.Stuck {
		return nil, ErrBadInput
	}
	var out *Result
	err := s.write(ctx, func(tx *sql.Tx) error {
		var difficulty, urlKey string
		var started time.Time
		var ended *time.Time
		err := tx.QueryRowContext(ctx,
			`SELECT difficulty, started_at, ended_at, COALESCE(url_key, '') FROM attempts WHERE id = $1 AND kind = 'problem'`, id).
			Scan(&difficulty, &started, &ended, &urlKey)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if ended != nil {
			return ErrFinished
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE attempts SET ended_at = $1, outcome = $2 WHERE id = $3`, now, outcome, id); err != nil {
			return err
		}
		if err := completeRevisions(ctx, tx, urlKey, now, now); err != nil {
			return err
		}
		awards := game.ProblemAwards(difficulty, outcome, now.Sub(started))
		out, err = finishResult(ctx, tx, id, awards, started, now)
		return err
	})
	return out, err
}

// LogAttempt records a problem solved without the timer: you say how long it took.
func (s *Store) LogAttempt(ctx context.Context, problemID, url, title, difficulty, outcome string, minutes, daysAgo int, now time.Time) (*Result, error) {
	if (problemID == "" && !game.ValidDifficulty(difficulty)) || (outcome != game.Solved && outcome != game.Stuck) ||
		minutes < 1 || minutes > 600 || daysAgo < 0 || daysAgo > 1 {
		return nil, ErrBadInput
	}
	ended := now.AddDate(0, 0, -daysAgo)
	started := ended.Add(-time.Duration(minutes) * time.Minute)
	var out *Result
	err := s.write(ctx, func(tx *sql.Tx) error {
		ref, err := resolveRef(ctx, tx, problemID, url, title, difficulty)
		if err != nil {
			return err
		}
		var id int64
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO attempts (kind, url, title, difficulty, started_at, ended_at, outcome, url_key)
			 VALUES ('problem', NULLIF($1, ''), NULLIF($2, ''), $3, $4, $5, $6, NULLIF($7, '')) RETURNING id`,
			ref.url, ref.title, ref.difficulty, started, ended, outcome, ref.urlKey).Scan(&id); err != nil {
			return err
		}
		if err := completeRevisions(ctx, tx, ref.urlKey, ended, now); err != nil {
			return err
		}
		awards := game.ProblemAwards(ref.difficulty, outcome, ended.Sub(started))
		out, err = finishResult(ctx, tx, id, awards, started, now)
		return err
	})
	return out, err
}

func (s *Store) DiscardAttempt(ctx context.Context, id int64) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`DELETE FROM attempts WHERE id = $1 AND kind = 'problem' AND ended_at IS NULL`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Store) LogChess(ctx context.Context, now time.Time) (*Result, error) {
	var out *Result
	err := s.write(ctx, func(tx *sql.Tx) error {
		var id int64
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO attempts (kind, started_at, ended_at) VALUES ('chess', $1, $1) RETURNING id`, now).
			Scan(&id); err != nil {
			return err
		}
		var err error
		out, err = finishResult(ctx, tx, id,
			[]game.Award{{Reason: game.ReasonChess, Points: game.ChessPoints}}, now, now)
		return err
	})
	return out, err
}

type SpendResult struct {
	Reward  Reward `json:"reward"`
	Spent   int    `json:"spent"`
	Balance int    `json:"balance"`
}

func (s *Store) Buy(ctx context.Context, rewardID int64, now time.Time) (*SpendResult, error) {
	var out *SpendResult
	err := s.write(ctx, func(tx *sql.Tx) error {
		var r Reward
		err := tx.QueryRowContext(ctx,
			`SELECT id, name, cost, mystery_weight FROM rewards WHERE id = $1 AND active`, rewardID).
			Scan(&r.ID, &r.Name, &r.Cost, &r.MysteryWeight)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if r.Cost == nil {
			return ErrNotForSale
		}
		bal, err := balance(ctx, tx)
		if err != nil {
			return err
		}
		if bal < *r.Cost {
			return ErrNotEnough
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ledger (at, delta, reason, reward_id) VALUES ($1, $2, $3, $4)`,
			now, -*r.Cost, game.ReasonBuy, r.ID); err != nil {
			return err
		}
		out = &SpendResult{Reward: r, Spent: *r.Cost, Balance: bal - *r.Cost}
		return nil
	})
	return out, err
}

func (s *Store) Mystery(ctx context.Context, now time.Time) (*SpendResult, error) {
	var out *SpendResult
	err := s.write(ctx, func(tx *sql.Tx) error {
		bal, err := balance(ctx, tx)
		if err != nil {
			return err
		}
		if bal < game.MysteryCost {
			return ErrNotEnough
		}
		rows, err := tx.QueryContext(ctx,
			`SELECT id, name, cost, mystery_weight FROM rewards WHERE active AND mystery_weight > 0`)
		if err != nil {
			return err
		}
		byID := map[int64]Reward{}
		var pool []game.Prize
		for rows.Next() {
			var r Reward
			if err := rows.Scan(&r.ID, &r.Name, &r.Cost, &r.MysteryWeight); err != nil {
				rows.Close()
				return err
			}
			byID[r.ID] = r
			pool = append(pool, game.Prize{ID: r.ID, Weight: r.MysteryWeight})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		id, ok := game.DrawMystery(pool, nil)
		if !ok {
			return ErrEmptyMystery
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ledger (at, delta, reason, reward_id) VALUES ($1, $2, $3, $4)`,
			now, -game.MysteryCost, game.ReasonMystery, id); err != nil {
			return err
		}
		out = &SpendResult{Reward: byID[id], Spent: game.MysteryCost, Balance: bal - game.MysteryCost}
		return nil
	})
	return out, err
}

func (s *Store) ClaimMilestone(ctx context.Context, id int64, now time.Time) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		var needed int
		var claimed *time.Time
		err := tx.QueryRowContext(ctx,
			`SELECT good_weeks_needed, claimed_at FROM milestones WHERE id = $1`, id).Scan(&needed, &claimed)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if claimed != nil {
			return ErrClaimed
		}
		var have int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM ledger WHERE reason = $1`, game.ReasonGoodWeek).Scan(&have); err != nil {
			return err
		}
		if have < needed {
			return ErrLocked
		}
		_, err = tx.ExecContext(ctx, `UPDATE milestones SET claimed_at = $1 WHERE id = $2`, now, id)
		return err
	})
}

// RewardInput is what the shop editor sends.
type RewardInput struct {
	Name          string `json:"name"`
	Cost          *int   `json:"cost"`
	MysteryWeight int    `json:"mysteryWeight"`
}

func (in RewardInput) valid() bool {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 80 || in.MysteryWeight < 0 || in.MysteryWeight > 1000 {
		return false
	}
	if in.Cost != nil && (*in.Cost < 1 || *in.Cost > 100000) {
		return false
	}
	return in.Cost != nil || in.MysteryWeight > 0 // must be buyable or drawable
}

func (s *Store) CreateReward(ctx context.Context, in RewardInput) error {
	if !in.valid() {
		return ErrBadInput
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO rewards (name, cost, mystery_weight) VALUES ($1, $2, $3)`,
			strings.TrimSpace(in.Name), in.Cost, in.MysteryWeight)
		return err
	})
}

func (s *Store) UpdateReward(ctx context.Context, id int64, in RewardInput) error {
	if !in.valid() {
		return ErrBadInput
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`UPDATE rewards SET name = $1, cost = $2, mystery_weight = $3 WHERE id = $4 AND active`,
			strings.TrimSpace(in.Name), in.Cost, in.MysteryWeight, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// RemoveReward hides a reward. Past purchases keep pointing at it.
func (s *Store) RemoveReward(ctx context.Context, id int64) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE rewards SET active = false WHERE id = $1 AND active`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// Ping checks the database is reachable.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
