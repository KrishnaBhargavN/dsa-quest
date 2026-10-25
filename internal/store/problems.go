package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"dsaquest/internal/game"
	"dsaquest/internal/lessons"
)

//go:embed problems.json
var problemsJSON []byte

// ErrDuplicate means a problem with the same link is already on the list.
var ErrDuplicate = errors.New("that problem is already on the list")

// Problem statuses.
const (
	StatusTodo   = "todo"
	StatusSolved = "solved"
	StatusStuck  = "stuck"
	StatusBefore = "before"
)

type Problem struct {
	ID          string   `json:"id"`
	Ord         float64  `json:"ord"`
	Phase       *int     `json:"phase"`
	Topic       string   `json:"topic,omitempty"` // empty while a blind problem is unfinished
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Difficulty  string   `json:"difficulty"`
	Note        string   `json:"note"`
	Description string   `json:"description"`
	Hints       []string `json:"hints"`
	Blind       bool     `json:"blind"`
	Custom      bool     `json:"custom"`
	Status      string   `json:"status"`
	Redo        bool     `json:"redo"`
	RevisionDue *string  `json:"revisionDue"`
	// LessonID is the theory lesson for this problem's topic; empty while a
	// blind problem is unfinished, so it can't give the topic away.
	LessonID string `json:"lessonId"`
	// FirstOfTopic is true when no problem with this topic is finished yet.
	FirstOfTopic bool `json:"firstOfTopic"`
	urlKey       string
}

type ProblemList struct {
	Problems []Problem `json:"problems"`
	NextID   string    `json:"nextId"`
	Done     int       `json:"done"`
	Total    int       `json:"total"`
}

type Revision struct {
	ID         int64      `json:"id"`
	ProblemID  string     `json:"problemId,omitempty"`
	URL        string     `json:"url"`
	Title      string     `json:"title"`
	Difficulty string     `json:"difficulty"`
	Due        string     `json:"due"`
	DoneAt     *time.Time `json:"doneAt"`
}

// catalogueRow is one entry of problems.json.
type catalogueRow struct {
	Order       int      `json:"order"`
	ID          string   `json:"id"`
	Phase       int      `json:"phase"`
	Topic       string   `json:"topic"`
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"`
	URL         string   `json:"url"`
	Blind       bool     `json:"blind"`
	Note        string   `json:"note"`
	Description string   `json:"description"`
	Hints       []string `json:"hints"`
}

func catalogue() ([]catalogueRow, error) {
	var rows []catalogueRow
	if err := json.Unmarshal(problemsJSON, &rows); err != nil {
		return nil, fmt.Errorf("problems.json: %w", err)
	}
	return rows, nil
}

// seedProblems upserts the built-in list in one statement. Existing rows get
// their content refreshed; progress (attempts, done_before_at) and removals
// (active) are never touched. A row whose link is already used by another
// active problem (say, one he added himself) is skipped.
func seedProblems(ctx context.Context, tx *sql.Tx) error {
	rows, err := catalogue()
	if err != nil {
		return err
	}
	type seedRow struct {
		catalogueRow
		URLKey string `json:"url_key"`
	}
	seed := make([]seedRow, len(rows))
	for i, r := range rows {
		seed[i] = seedRow{r, game.URLKey(r.URL)}
	}
	payload, err := json.Marshal(seed)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO problems (id, ord, phase, topic, title, url, url_key, difficulty, note, description, hints, blind)
		SELECT r.id, r."order", r.phase, r.topic, r.title, r.url, r.url_key, r.difficulty, r.note, r.description,
		       ARRAY(SELECT jsonb_array_elements_text(r.hints)), r.blind
		FROM jsonb_to_recordset($1::jsonb) AS r(id text, "order" float8, phase int, topic text, title text, url text,
		     url_key text, difficulty text, note text, description text, hints jsonb, blind boolean)
		WHERE NOT EXISTS (SELECT 1 FROM problems q WHERE q.url_key = r.url_key AND q.id <> r.id AND q.active)
		ON CONFLICT (id) DO UPDATE SET
		  ord = EXCLUDED.ord, phase = EXCLUDED.phase, topic = EXCLUDED.topic, title = EXCLUDED.title,
		  url = EXCLUDED.url, url_key = EXCLUDED.url_key, difficulty = EXCLUDED.difficulty, note = EXCLUDED.note,
		  description = EXCLUDED.description, hints = EXCLUDED.hints, blind = EXCLUDED.blind
		WHERE NOT problems.custom`, string(payload))
	return err
}

// backfillURLKeys fills attempts.url_key for attempts logged before the column existed.
func backfillURLKeys(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT id, url FROM attempts WHERE url_key IS NULL AND url IS NOT NULL`)
	if err != nil {
		return err
	}
	keys := map[int64]string{}
	for rows.Next() {
		var id int64
		var u string
		if err := rows.Scan(&id, &u); err != nil {
			rows.Close()
			return err
		}
		keys[id] = game.URLKey(u)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for id, k := range keys {
		if _, err := tx.ExecContext(ctx, `UPDATE attempts SET url_key = NULLIF($1, '') WHERE id = $2`, k, id); err != nil {
			return err
		}
	}
	return nil
}

// ---------- reads ----------

type attemptStats struct {
	solved       bool
	lastOutcome  string
	lastDiff     string
	lastDuration time.Duration
}

func loadProblems(ctx context.Context, q queryer) (*ProblemList, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, ord, phase, topic, title, url, COALESCE(url_key, ''), difficulty, note, description, hints,
		       blind, custom, done_before_at IS NOT NULL
		FROM problems WHERE active ORDER BY ord, id`)
	if err != nil {
		return nil, err
	}
	list := &ProblemList{Problems: []Problem{}}
	before := map[string]bool{}
	for rows.Next() {
		var p Problem
		var hints pq.StringArray
		var isBefore bool
		if err := rows.Scan(&p.ID, &p.Ord, &p.Phase, &p.Topic, &p.Title, &p.URL, &p.urlKey, &p.Difficulty,
			&p.Note, &p.Description, &hints, &p.Blind, &p.Custom, &isBefore); err != nil {
			rows.Close()
			return nil, err
		}
		p.Hints = []string(hints)
		if p.Hints == nil {
			p.Hints = []string{}
		}
		before[p.ID] = isBefore
		list.Problems = append(list.Problems, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats := map[string]*attemptStats{}
	arows, err := q.QueryContext(ctx, `
		SELECT url_key, outcome, difficulty, EXTRACT(EPOCH FROM ended_at - started_at)
		FROM attempts WHERE kind = 'problem' AND ended_at IS NOT NULL AND url_key IS NOT NULL
		ORDER BY ended_at`)
	if err != nil {
		return nil, err
	}
	for arows.Next() {
		var key, outcome, diff string
		var secs float64
		if err := arows.Scan(&key, &outcome, &diff, &secs); err != nil {
			arows.Close()
			return nil, err
		}
		st := stats[key]
		if st == nil {
			st = &attemptStats{}
			stats[key] = st
		}
		st.solved = st.solved || outcome == game.Solved
		st.lastOutcome, st.lastDiff, st.lastDuration = outcome, diff, time.Duration(secs*float64(time.Second))
	}
	arows.Close()
	if err := arows.Err(); err != nil {
		return nil, err
	}

	due := map[string]string{}
	rrows, err := q.QueryContext(ctx,
		`SELECT url_key, MIN(due) FROM revisions WHERE done_at IS NULL AND url_key IS NOT NULL GROUP BY url_key`)
	if err != nil {
		return nil, err
	}
	for rrows.Next() {
		var key string
		var d time.Time
		if err := rrows.Scan(&key, &d); err != nil {
			rrows.Close()
			return nil, err
		}
		due[key] = d.Format("2006-01-02")
	}
	rrows.Close()
	if err := rrows.Err(); err != nil {
		return nil, err
	}

	topicStarted := map[string]bool{}
	for i := range list.Problems {
		p := &list.Problems[i]
		st := stats[p.urlKey]
		switch {
		case st != nil && st.solved:
			p.Status = StatusSolved
		case before[p.ID]:
			p.Status = StatusBefore
		case st != nil:
			p.Status = StatusStuck
		default:
			p.Status = StatusTodo
		}
		if st != nil {
			p.Redo = game.NeedsRedo(st.lastOutcome, st.lastDiff, st.lastDuration)
		}
		if d, ok := due[p.urlKey]; ok {
			p.RevisionDue = &d
		}
		if p.Status != StatusTodo && p.Topic != "" {
			topicStarted[p.Topic] = true
		}
		if p.Blind && p.Status == StatusTodo {
			p.Topic = ""
		}
		p.LessonID = lessons.ForTopic(p.Topic)
		if p.Status == StatusTodo {
			if list.NextID == "" {
				list.NextID = p.ID
			}
		} else {
			list.Done++
		}
	}
	for i := range list.Problems {
		p := &list.Problems[i]
		p.FirstOfTopic = p.Topic != "" && !topicStarted[p.Topic]
	}
	list.Total = len(list.Problems)
	return list, nil
}

func (s *Store) Problems(ctx context.Context) (*ProblemList, error) { return loadProblems(ctx, s.db) }

// nextUp returns the first unfinished problem, or nil when the list is done.
func nextUp(list *ProblemList) *Problem {
	for i := range list.Problems {
		if list.Problems[i].ID == list.NextID {
			return &list.Problems[i]
		}
	}
	return nil
}

const revisionCols = `r.id, COALESCE(p.id, ''), r.url, r.title, r.difficulty, r.due, r.done_at`

func scanRevisions(rows *sql.Rows) ([]Revision, error) {
	defer rows.Close()
	out := []Revision{}
	for rows.Next() {
		var r Revision
		var d time.Time
		if err := rows.Scan(&r.ID, &r.ProblemID, &r.URL, &r.Title, &r.Difficulty, &d, &r.DoneAt); err != nil {
			return nil, err
		}
		r.Due = d.Format("2006-01-02")
		out = append(out, r)
	}
	return out, rows.Err()
}

// Revisions returns every revision due between from and to (inclusive), done or not.
func (s *Store) Revisions(ctx context.Context, from, to time.Time) ([]Revision, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+revisionCols+` FROM revisions r
		LEFT JOIN problems p ON p.url_key = r.url_key AND p.active
		WHERE r.due BETWEEN $1::date AND $2::date ORDER BY r.due, r.id`,
		from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	return scanRevisions(rows)
}

func dueRevisions(ctx context.Context, q queryer, today time.Time) ([]Revision, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+revisionCols+` FROM revisions r
		LEFT JOIN problems p ON p.url_key = r.url_key AND p.active
		WHERE r.done_at IS NULL AND r.due <= $1::date ORDER BY r.due, r.id`, today.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	return scanRevisions(rows)
}

// ---------- writes ----------

// problemRef is what an attempt or revision needs to know about a problem.
type problemRef struct {
	url, title, difficulty, urlKey string
}

func lookupProblem(ctx context.Context, q queryer, id string) (*problemRef, error) {
	var r problemRef
	err := q.QueryRowContext(ctx,
		`SELECT url, title, difficulty, COALESCE(url_key, '') FROM problems WHERE id = $1 AND active`, id).
		Scan(&r.url, &r.title, &r.difficulty, &r.urlKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &r, err
}

// completeRevisions marks open revisions for this problem done when an attempt
// on it finishes after they were scheduled.
func completeRevisions(ctx context.Context, tx *sql.Tx, urlKey string, endedAt, now time.Time) error {
	if urlKey == "" {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE revisions SET done_at = $1 WHERE url_key = $2 AND done_at IS NULL AND created_at < $3`,
		now, urlKey, endedAt)
	return err
}

func (s *Store) SetDoneBefore(ctx context.Context, id string, on bool, now time.Time) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		var at *time.Time
		if on {
			at = &now
		}
		res, err := tx.ExecContext(ctx, `UPDATE problems SET done_before_at = $1 WHERE id = $2 AND active`, at, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ProblemInput is what the problem editor sends.
type ProblemInput struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Difficulty  string   `json:"difficulty"`
	Topic       string   `json:"topic"`
	Description string   `json:"description"`
	Hints       []string `json:"hints"`
	Placement   string   `json:"placement"` // "next" or "end"; create only
}

func (in *ProblemInput) clean() bool {
	in.Title = strings.TrimSpace(in.Title)
	in.URL = strings.TrimSpace(in.URL)
	in.Topic = strings.TrimSpace(in.Topic)
	in.Description = strings.TrimSpace(in.Description)
	hints := []string{}
	for _, h := range in.Hints {
		if h = strings.TrimSpace(h); h != "" {
			hints = append(hints, h)
		}
	}
	in.Hints = hints
	return in.Title != "" && len(in.Title) <= 200 && len(in.URL) <= 500 && len(in.Topic) <= 80 &&
		len(in.Description) <= 2000 && len(in.Hints) <= 2 && game.ValidDifficulty(in.Difficulty)
}

func keyFor(id, url string) string {
	if k := game.URLKey(url); k != "" {
		return k
	}
	return "id:" + id
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (s *Store) CreateProblem(ctx context.Context, in ProblemInput, now time.Time) (string, error) {
	if !in.clean() {
		return "", ErrBadInput
	}
	id := fmt.Sprintf("my-%d", now.UnixMilli())
	err := s.write(ctx, func(tx *sql.Tx) error {
		list, err := loadProblems(ctx, tx)
		if err != nil {
			return err
		}
		ord := 1.0
		if n := len(list.Problems); n > 0 {
			ord = list.Problems[n-1].Ord + 1
		}
		if next := nextUp(list); next != nil && in.Placement == "next" {
			ord = next.Ord - 0.5
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO problems (id, ord, topic, title, url, url_key, difficulty, description, hints, custom)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)`,
			id, ord, in.Topic, in.Title, in.URL, keyFor(id, in.URL), in.Difficulty, in.Description, pq.StringArray(in.Hints))
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		return err
	})
	return id, err
}

func (s *Store) UpdateProblem(ctx context.Context, id string, in ProblemInput) error {
	if !in.clean() {
		return ErrBadInput
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE problems SET topic = $1, title = $2, url = $3, url_key = $4, difficulty = $5, description = $6, hints = $7
			WHERE id = $8 AND custom AND active`,
			in.Topic, in.Title, in.URL, keyFor(id, in.URL), in.Difficulty, in.Description, pq.StringArray(in.Hints), id)
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// RemoveProblem hides a problem from the list. Its attempts stay in history.
func (s *Store) RemoveProblem(ctx context.Context, id string) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE problems SET active = false WHERE id = $1 AND active`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// RevisionInput schedules a revision, either for a list problem or for any link.
type RevisionInput struct {
	ProblemID  string `json:"problemId"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	InDays     int    `json:"inDays"`
	Due        string `json:"due"` // YYYY-MM-DD, used when InDays is 0
}

func parseDue(due string, now time.Time) (time.Time, error) {
	d, err := time.Parse("2006-01-02", due)
	if err != nil || d.Before(game.Day(now)) || d.After(game.Day(now).AddDate(1, 0, 0)) {
		return time.Time{}, ErrBadInput
	}
	return d, nil
}

// ScheduleRevision adds a revision, or moves the open one if the problem already has one.
func (s *Store) ScheduleRevision(ctx context.Context, in RevisionInput, now time.Time) (*Revision, error) {
	var due time.Time
	switch {
	case in.InDays > 0 && in.InDays <= 365:
		due = game.RevisionDue(now, in.InDays)
	case in.InDays == 0:
		d, err := parseDue(in.Due, now)
		if err != nil {
			return nil, err
		}
		due = d
	default:
		return nil, ErrBadInput
	}
	var out *Revision
	err := s.write(ctx, func(tx *sql.Tx) error {
		ref := &problemRef{url: strings.TrimSpace(in.URL), title: strings.TrimSpace(in.Title), difficulty: in.Difficulty}
		if in.ProblemID != "" {
			var err error
			if ref, err = lookupProblem(ctx, tx, in.ProblemID); err != nil {
				return err
			}
		} else {
			ref.urlKey = game.URLKey(ref.url)
			if ref.title == "" {
				ref.title = ref.url
			}
		}
		if ref.title == "" || !game.ValidDifficulty(ref.difficulty) {
			return ErrBadInput
		}
		var id int64
		err := tx.QueryRowContext(ctx,
			`UPDATE revisions SET due = $1::date WHERE url_key = $2 AND done_at IS NULL RETURNING id`,
			due.Format("2006-01-02"), ref.urlKey).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO revisions (url, url_key, title, difficulty, due, created_at)
				VALUES ($1, NULLIF($2, ''), $3, $4, $5::date, $6) RETURNING id`,
				ref.url, ref.urlKey, ref.title, ref.difficulty, due.Format("2006-01-02"), now).Scan(&id)
		}
		if err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT `+revisionCols+` FROM revisions r
			LEFT JOIN problems p ON p.url_key = r.url_key AND p.active WHERE r.id = $1`, id)
		if err != nil {
			return err
		}
		revs, err := scanRevisions(rows)
		if err != nil {
			return err
		}
		out = &revs[0]
		return nil
	})
	return out, err
}

func (s *Store) MoveRevision(ctx context.Context, id int64, due string, now time.Time) error {
	d, err := parseDue(due, now)
	if err != nil {
		return err
	}
	return s.updateRevision(ctx, `UPDATE revisions SET due = $1::date WHERE id = $2 AND done_at IS NULL`, d.Format("2006-01-02"), id)
}

func (s *Store) CompleteRevision(ctx context.Context, id int64, now time.Time) error {
	return s.updateRevision(ctx, `UPDATE revisions SET done_at = $1 WHERE id = $2 AND done_at IS NULL`, now, id)
}

func (s *Store) DeleteRevision(ctx context.Context, id int64) error {
	return s.updateRevision(ctx, `DELETE FROM revisions WHERE id = $1`, id)
}

func (s *Store) updateRevision(ctx context.Context, query string, args ...any) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// resolveRef returns what an attempt records: the list problem's details when
// problemID is set, otherwise the given link, title and difficulty.
func resolveRef(ctx context.Context, q queryer, problemID, url, title, difficulty string) (*problemRef, error) {
	if problemID != "" {
		return lookupProblem(ctx, q, problemID)
	}
	return &problemRef{url: url, title: title, difficulty: difficulty, urlKey: game.URLKey(url)}, nil
}
