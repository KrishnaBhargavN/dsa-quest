package api

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"dsaquest/internal/store"
)

func problemList(t *testing.T, c *client) store.ProblemList {
	t.Helper()
	var l store.ProblemList
	if code := c.do("GET", "/api/problems", nil, &l); code != 200 {
		t.Fatalf("problems = %d", code)
	}
	return l
}

func find(l store.ProblemList, id string) *store.Problem {
	for i := range l.Problems {
		if l.Problems[i].ID == id {
			return &l.Problems[i]
		}
	}
	return nil
}

// solve starts a list problem with the timer and finishes it after the given minutes.
func solve(t *testing.T, c *client, now *time.Time, id, outcome string, minutes int) store.Result {
	t.Helper()
	var a store.Attempt
	if code := c.do("POST", "/api/attempts", map[string]string{"problemId": id}, &a); code != 201 {
		t.Fatalf("start %s = %d", id, code)
	}
	*now = now.Add(time.Duration(minutes) * time.Minute)
	var r store.Result
	if code := c.do("POST", "/api/attempts/"+itoa(a.ID)+"/finish", map[string]string{"outcome": outcome}, &r); code != 200 {
		t.Fatalf("finish %s = %d", id, code)
	}
	return r
}

func TestNextUpStatusAndRedo(t *testing.T) {
	c, now, _ := setup(t)
	login(t, c)

	l := problemList(t, c)
	if l.Total != 151 || l.Done != 0 || l.NextID != "lc-1462" {
		t.Fatalf("fresh list: total %d done %d next %s", l.Total, l.Done, l.NextID)
	}
	if s := summary(t, c); s.NextUp == nil || s.NextUp.ID != "lc-1462" || len(s.NextUp.Hints) != 2 || s.NextUp.Description == "" {
		t.Fatalf("summary next up = %+v", s.NextUp)
	}

	// Starting from the list fills in title and difficulty; the timer screen gets the problem.
	var a store.Attempt
	c.do("POST", "/api/attempts", map[string]string{"problemId": "lc-1462"}, &a)
	if a.Title != "Course Schedule IV" || a.Difficulty != "medium" {
		t.Errorf("attempt from list = %q %q", a.Title, a.Difficulty)
	}
	if s := summary(t, c); s.RunningProblem == nil || s.RunningProblem.ID != "lc-1462" {
		t.Errorf("running problem = %+v", s.RunningProblem)
	}
	*now = now.Add(10 * time.Minute)
	var r store.Result
	c.do("POST", "/api/attempts/"+itoa(a.ID)+"/finish", map[string]string{"outcome": "solved"}, &r)
	if r.Earned != 25 {
		t.Errorf("solved fast medium earned %d, want 25", r.Earned)
	}

	// Stuck moves you on and puts the problem on the redo list.
	solve(t, c, now, "lc-827", "stuck", 50)
	l = problemList(t, c)
	if p := find(l, "lc-827"); p.Status != store.StatusStuck || !p.Redo {
		t.Errorf("stuck problem = %s redo %v", p.Status, p.Redo)
	}
	if l.NextID != "lc-1584" {
		t.Errorf("next after stuck = %s, want lc-1584", l.NextID)
	}

	// Done before: no points, drops out of next up, and can be undone.
	bal := summary(t, c).Balance
	if code := c.do("POST", "/api/problems/lc-1584/before", nil, nil); code != 200 {
		t.Fatalf("done before = %d", code)
	}
	if l = problemList(t, c); l.NextID != "lc-300" || find(l, "lc-1584").Status != store.StatusBefore {
		t.Errorf("after done before: next %s status %s", l.NextID, find(l, "lc-1584").Status)
	}
	if s := summary(t, c); s.Balance != bal {
		t.Errorf("done before changed balance %d -> %d", bal, s.Balance)
	}
	c.do("DELETE", "/api/problems/lc-1584/before", nil, nil)
	if l = problemList(t, c); l.NextID != "lc-1584" {
		t.Errorf("after undo: next %s", l.NextID)
	}
	c.do("POST", "/api/problems/lc-1584/before", nil, nil)

	// A link pasted by hand, however it was copied, ticks off the list item.
	// A medium over 40 minutes goes on the redo list; a fast re-solve clears it.
	if code := c.do("POST", "/api/attempts/log", map[string]any{
		"url":        "https://leetcode.com/problems/Longest-Increasing-Subsequence/description/?envType=study-plan",
		"difficulty": "medium", "outcome": "solved", "minutes": 45, "daysAgo": 0,
	}, nil); code != 201 {
		t.Fatalf("manual log = %d", code)
	}
	l = problemList(t, c)
	if p := find(l, "lc-300"); p.Status != store.StatusSolved || !p.Redo {
		t.Errorf("pasted link: status %s redo %v", p.Status, p.Redo)
	}
	if l.NextID != "lc-213" {
		t.Errorf("next after pasted solve = %s", l.NextID)
	}
	solve(t, c, now, "lc-300", "solved", 20)
	if p := find(problemList(t, c), "lc-300"); p.Redo {
		t.Error("fast re-solve should clear redo")
	}
	if l = problemList(t, c); l.Done != 4 {
		t.Errorf("done = %d, want 4", l.Done)
	}
}

func TestBlindTopicHiddenUntilFinished(t *testing.T) {
	c, _, _ := setup(t)
	login(t, c)
	l := problemList(t, c)
	blind := 0
	for _, p := range l.Problems {
		if p.Blind {
			blind++
			if p.Topic != "" {
				t.Errorf("%s is blind but shows topic %q", p.ID, p.Topic)
			}
		} else if p.Topic == "" {
			t.Errorf("%s has no topic", p.ID)
		}
	}
	if blind != 17 {
		t.Errorf("blind problems = %d, want 17", blind)
	}
	c.do("POST", "/api/problems/lc-994/before", nil, nil)
	if p := find(problemList(t, c), "lc-994"); p.Topic != "Graphs" {
		t.Errorf("finished blind topic = %q, want Graphs", p.Topic)
	}
}

func TestSeedKeepsProgressAndRemovals(t *testing.T) {
	c, now, _ := setup(t)
	login(t, c)
	solve(t, c, now, "lc-1462", "solved", 10)
	if code := c.do("DELETE", "/api/problems/lc-213", nil, nil); code != 200 {
		t.Fatalf("remove = %d", code)
	}
	c.do("POST", "/api/problems", map[string]any{"title": "Mine", "difficulty": "easy", "placement": "end"}, nil)

	// A fresh process runs the schema and seed again.
	st, err := store.Open(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	srv := New(st, "hunter2", fstest.MapFS{"index.html": {Data: []byte("x")}})
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	jar, _ := cookiejar.New(nil)
	c2 := &client{t: t, base: ts.URL, http: &http.Client{Jar: jar}}
	login(t, c2)

	l := problemList(t, c2)
	if l.Total != 151 {
		t.Errorf("total after reseed = %d, want 151 (150 seeded + 1 own)", l.Total)
	}
	if find(l, "lc-213") != nil {
		t.Error("removed problem came back")
	}
	if p := find(l, "lc-1462"); p == nil || p.Status != store.StatusSolved {
		t.Error("progress lost on reseed")
	}
	seen := map[string]bool{}
	for _, p := range l.Problems {
		if seen[p.ID] {
			t.Errorf("duplicate %s", p.ID)
		}
		seen[p.ID] = true
	}
}

func TestOwnProblems(t *testing.T) {
	c, now, _ := setup(t)
	login(t, c)

	var created map[string]string
	if code := c.do("POST", "/api/problems", map[string]any{
		"title": "Two Sum again", "url": "https://leetcode.com/problems/two-sum/", "difficulty": "easy",
		"topic": "Warm-up", "description": "Quick one.", "hints": []string{"hash map", " "}, "placement": "next",
	}, &created); code != 201 {
		t.Fatalf("create = %d", code)
	}
	id := created["id"]
	l := problemList(t, c)
	p := find(l, id)
	if l.NextID != id || p == nil || !p.Custom || len(p.Hints) != 1 {
		t.Fatalf("own problem: next %s, %+v", l.NextID, p)
	}
	if code := c.do("POST", "/api/problems", map[string]any{
		"title": "Dup", "url": "https://leetcode.com/problems/Two-Sum/description/", "difficulty": "easy",
	}, nil); code != 409 {
		t.Errorf("duplicate link = %d, want 409", code)
	}
	if code := c.do("POST", "/api/problems", map[string]any{
		"title": "Too many hints", "difficulty": "easy", "hints": []string{"a", "b", "c"},
	}, nil); code != 400 {
		t.Errorf("3 hints = %d, want 400", code)
	}
	if code := c.do("PUT", "/api/problems/"+id, map[string]any{
		"title": "Two Sum, renamed", "url": "https://leetcode.com/problems/two-sum/", "difficulty": "medium",
	}, nil); code != 200 {
		t.Errorf("edit own = %d", code)
	}
	if code := c.do("PUT", "/api/problems/lc-1462", map[string]any{"title": "x", "difficulty": "easy"}, nil); code != 404 {
		t.Errorf("edit seeded = %d, want 404", code)
	}

	// A problem without a link still tracks its own progress.
	*now = now.Add(time.Second)
	c.do("POST", "/api/problems", map[string]any{"title": "Whiteboard drill", "difficulty": "hard", "placement": "end"}, &created)
	noLink := created["id"]
	l = problemList(t, c)
	if l.Problems[len(l.Problems)-1].ID != noLink {
		t.Errorf("end placement: last is %s", l.Problems[len(l.Problems)-1].ID)
	}
	solve(t, c, now, noLink, "solved", 30)
	if p := find(problemList(t, c), noLink); p.Status != store.StatusSolved {
		t.Errorf("linkless problem status = %s", p.Status)
	}
}

func TestRevisions(t *testing.T) {
	c, now, _ := setup(t) // Monday 19 Oct 2026, 8 PM IST
	login(t, c)
	solve(t, c, now, "lc-1462", "solved", 10)

	var rev store.Revision
	if code := c.do("POST", "/api/revisions", map[string]any{"problemId": "lc-1462", "inDays": 7}, &rev); code != 201 {
		t.Fatalf("schedule = %d", code)
	}
	if rev.Due != "2026-10-26" || rev.ProblemID != "lc-1462" {
		t.Errorf("revision = %+v", rev)
	}
	var moved store.Revision
	c.do("POST", "/api/revisions", map[string]any{"problemId": "lc-1462", "inDays": 3}, &moved)
	if moved.ID != rev.ID || moved.Due != "2026-10-22" {
		t.Errorf("rescheduling should move the open revision: %+v", moved)
	}
	var other store.Revision
	c.do("POST", "/api/revisions", map[string]any{
		"url": "https://example.com/p/1", "title": "Off-list", "difficulty": "easy", "inDays": 1,
	}, &other)
	if code := c.do("POST", "/api/revisions", map[string]any{"problemId": "lc-1462", "due": "2026-10-01"}, nil); code != 400 {
		t.Errorf("past due date = %d, want 400", code)
	}

	var cal struct {
		Today     string           `json:"today"`
		Revisions []store.Revision `json:"revisions"`
	}
	c.do("GET", "/api/revisions?from=2026-10-01&to=2026-10-31", nil, &cal)
	if cal.Today != "2026-10-19" || len(cal.Revisions) != 2 {
		t.Errorf("calendar = %+v", cal)
	}
	if s := summary(t, c); len(s.RevisionsDue) != 0 {
		t.Errorf("nothing is due yet, got %d", len(s.RevisionsDue))
	}

	*now = now.Add(3 * 24 * time.Hour) // Thursday
	if s := summary(t, c); len(s.RevisionsDue) != 2 {
		t.Fatalf("due on Thursday = %d, want 2", len(s.RevisionsDue))
	}
	r := solve(t, c, now, "lc-1462", "solved", 15)
	if r.Earned != 25 {
		t.Errorf("revising earns normal points, got %d", r.Earned)
	}
	s := summary(t, c)
	if len(s.RevisionsDue) != 1 || s.RevisionsDue[0].ID != other.ID {
		t.Errorf("after solving, due = %+v", s.RevisionsDue)
	}
	if code := c.do("PUT", "/api/revisions/"+itoa(other.ID), map[string]string{"due": "2026-10-30"}, nil); code != 200 {
		t.Errorf("move = %d", code)
	}
	if s := summary(t, c); len(s.RevisionsDue) != 0 {
		t.Error("moved revision should not be due")
	}
	if code := c.do("POST", "/api/revisions/"+itoa(other.ID)+"/done", nil, nil); code != 200 {
		t.Errorf("done = %d", code)
	}
	if code := c.do("DELETE", "/api/revisions/"+itoa(other.ID), nil, nil); code != 200 {
		t.Errorf("delete = %d", code)
	}
}
