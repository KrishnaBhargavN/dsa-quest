package api

// End-to-end tests against a real Postgres. They run only when
// TEST_DATABASE_URL is set, and they wipe that database first.
//
//   TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/dsaquest_test?sslmode=disable go test ./...

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"dsaquest/internal/game"
	"dsaquest/internal/store"
)

type client struct {
	t    *testing.T
	base string
	http *http.Client
}

func (c *client) do(method, path string, body any, out any) int {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, rdr)
	req.Header.Set("X-DQ", "1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			c.t.Fatalf("%s %s: decode %q: %v", method, path, raw, err)
		}
	}
	return res.StatusCode
}

func setup(t *testing.T) (*client, *time.Time, *httptest.Server) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	raw, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`DROP TABLE IF EXISTS ledger, milestones, rewards, attempts, settings CASCADE`); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	st, err := store.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	static := fstest.MapFS{
		"index.html":           {Data: []byte("<!doctype html><title>DSA Quest</title>")},
		"assets/app-1.js":      {Data: []byte("console.log(1)")},
		"manifest.webmanifest": {Data: []byte("{}")},
	}
	srv := New(st, "hunter2", static)
	now := time.Date(2026, 10, 19, 20, 0, 0, 0, game.IST) // Monday 8 PM IST
	srv.SetClock(func() time.Time { return now })
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	jar, _ := cookiejar.New(nil)
	return &client{t: t, base: ts.URL, http: &http.Client{Jar: jar}}, &now, ts
}

func login(t *testing.T, c *client) {
	t.Helper()
	if code := c.do("POST", "/api/login", map[string]string{"password": "hunter2"}, nil); code != 200 {
		t.Fatalf("login = %d", code)
	}
}

func summary(t *testing.T, c *client) store.Summary {
	t.Helper()
	var s store.Summary
	if code := c.do("GET", "/api/summary", nil, &s); code != 200 {
		t.Fatalf("summary = %d", code)
	}
	return s
}

func TestAuthAndGuards(t *testing.T) {
	c, _, ts := setup(t)

	var h map[string]any
	if code := c.do("GET", "/api/health", nil, &h); code != 200 || h["database"] != "ok" {
		t.Fatalf("health = %d %v", code, h)
	}
	if code := c.do("GET", "/api/summary", nil, nil); code != 401 {
		t.Errorf("summary without login = %d, want 401", code)
	}
	start := time.Now()
	if code := c.do("POST", "/api/login", map[string]string{"password": "nope"}, nil); code != 401 {
		t.Errorf("wrong password = %d, want 401", code)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Error("wrong password should be slowed down")
	}
	login(t, c)
	if code := c.do("GET", "/api/summary", nil, nil); code != 200 {
		t.Errorf("summary after login = %d", code)
	}

	// A write without the X-DQ header (what a cross-site form would send) is refused.
	req, _ := http.NewRequest("POST", ts.URL+"/api/chess", nil)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 403 {
		t.Errorf("write without X-DQ = %d, want 403", res.StatusCode)
	}

	// A forged cookie doesn't work.
	req, _ = http.NewRequest("GET", ts.URL+"/api/summary", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "99999999999.deadbeef"})
	res, _ = http.DefaultClient.Do(req)
	res.Body.Close()
	if res.StatusCode != 401 {
		t.Errorf("forged cookie = %d, want 401", res.StatusCode)
	}

	c.do("POST", "/api/logout", nil, nil)
	if code := c.do("GET", "/api/summary", nil, nil); code != 401 {
		t.Errorf("summary after logout = %d, want 401", code)
	}
}

func TestStaticFiles(t *testing.T) {
	_, _, ts := setup(t)
	for path, wantCache := range map[string]string{
		"/":                "no-cache",
		"/shop":            "no-cache", // app route falls back to index.html
		"/assets/app-1.js": "public, max-age=31536000, immutable",
	} {
		res, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 200 || res.Header.Get("Cache-Control") != wantCache {
			t.Errorf("GET %s = %d, Cache-Control %q", path, res.StatusCode, res.Header.Get("Cache-Control"))
		}
	}
	res, _ := http.Get(ts.URL + "/api/nope")
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Errorf("unknown api path = %d, want 404", res.StatusCode)
	}
}

func TestTimerScoringAndShop(t *testing.T) {
	c, now, _ := setup(t)
	login(t, c)

	// Start a medium, finish it in 24 minutes: 15 + 10 bonus.
	var a store.Attempt
	if code := c.do("POST", "/api/attempts", map[string]string{
		"url": "https://leetcode.com/problems/two-sum/description/", "difficulty": "medium",
	}, &a); code != 201 {
		t.Fatalf("start = %d", code)
	}
	if a.Title != "Two Sum" {
		t.Errorf("title from URL = %q", a.Title)
	}
	if code := c.do("POST", "/api/attempts", map[string]string{"difficulty": "easy"}, nil); code != 409 {
		t.Errorf("second timer = %d, want 409", code)
	}
	if s := summary(t, c); s.Running == nil || s.Running.ID != a.ID {
		t.Error("summary should show the running timer")
	}
	*now = now.Add(24 * time.Minute)
	var r store.Result
	if code := c.do("POST", "/api/attempts/"+itoa(a.ID)+"/finish", map[string]string{"outcome": "solved"}, &r); code != 200 {
		t.Fatalf("finish = %d", code)
	}
	if r.Earned != 25 || r.Balance != 25 {
		t.Errorf("fast medium earned %d (balance %d), want 25", r.Earned, r.Balance)
	}
	if code := c.do("POST", "/api/attempts/"+itoa(a.ID)+"/finish", map[string]string{"outcome": "solved"}, nil); code != 409 {
		t.Errorf("finishing twice = %d, want 409", code)
	}

	// Stuck on a hard: 5. Discarding a timer logs nothing.
	c.do("POST", "/api/attempts", map[string]string{"difficulty": "hard"}, &a)
	*now = now.Add(70 * time.Minute)
	c.do("POST", "/api/attempts/"+itoa(a.ID)+"/finish", map[string]string{"outcome": "stuck"}, &r)
	if r.Earned != 5 || r.Balance != 30 {
		t.Errorf("stuck earned %d (balance %d), want 5 (30)", r.Earned, r.Balance)
	}
	c.do("POST", "/api/attempts", map[string]string{"difficulty": "easy"}, &a)
	if code := c.do("DELETE", "/api/attempts/"+itoa(a.ID), nil, nil); code != 200 {
		t.Errorf("discard = %d", code)
	}
	if s := summary(t, c); s.Running != nil || s.Balance != 30 || s.TodayPoints != 30 {
		t.Errorf("after discard: running=%v balance=%d today=%d", s.Running, s.Balance, s.TodayPoints)
	}

	// Shop: next reward is the cheapest one above the balance.
	s := summary(t, c)
	if s.NextReward == nil || s.NextReward.Cost != 50 || s.NextReward.Need != 20 {
		t.Errorf("next reward = %+v", s.NextReward)
	}
	var rewards []store.Reward
	c.do("GET", "/api/rewards", nil, &rewards)
	byName := map[string]store.Reward{}
	for _, rw := range rewards {
		byName[rw.Name] = rw
	}
	if code := c.do("POST", "/api/rewards/"+itoa(byName["1 hour of gaming"].ID)+"/buy", nil, nil); code != 409 {
		t.Errorf("buying with too few points = %d, want 409", code)
	}
	if code := c.do("POST", "/api/rewards/"+itoa(byName["Dessert"].ID)+"/buy", nil, nil); code != 409 {
		t.Errorf("buying a mystery-only prize = %d, want 409", code)
	}
	var sp store.SpendResult
	if code := c.do("POST", "/api/rewards/"+itoa(byName["30 min of gaming"].ID)+"/buy", nil, &sp); code != 200 || sp.Balance != 0 {
		t.Errorf("buy = %d, balance %d", code, sp.Balance)
	}
	if code := c.do("POST", "/api/mystery", nil, nil); code != 409 {
		t.Errorf("mystery with 0 points = %d, want 409", code)
	}

	// Earn 45 more with logged problems, then open a mystery box.
	c.do("POST", "/api/attempts/log", map[string]any{"difficulty": "hard", "outcome": "solved", "minutes": 50, "daysAgo": 0}, &r)
	c.do("POST", "/api/attempts/log", map[string]any{"difficulty": "medium", "outcome": "solved", "minutes": 30, "daysAgo": 0}, &r)
	if r.Balance != 45 {
		t.Fatalf("balance = %d, want 45", r.Balance)
	}
	if code := c.do("POST", "/api/mystery", nil, &sp); code != 200 || sp.Balance != 5 || sp.Reward.MysteryWeight == 0 {
		t.Errorf("mystery = %d %+v", code, sp)
	}
	if code := c.do("POST", "/api/attempts/log", map[string]any{"difficulty": "easy", "outcome": "solved", "minutes": 0}, nil); code != 400 {
		t.Errorf("log with 0 minutes = %d, want 400", code)
	}

	// Edit the shop.
	cost := 60
	if code := c.do("PUT", "/api/rewards/"+itoa(byName["1 hour of gaming"].ID), store.RewardInput{Name: "1 hour of gaming", Cost: &cost, MysteryWeight: 15}, nil); code != 200 {
		t.Errorf("update reward = %d", code)
	}
	if code := c.do("POST", "/api/rewards", store.RewardInput{Name: "Movie night", Cost: &cost}, nil); code != 201 {
		t.Errorf("create reward = %d", code)
	}
	if code := c.do("POST", "/api/rewards", store.RewardInput{Name: "Nothing"}, nil); code != 400 {
		t.Errorf("reward with no price and no weight = %d, want 400", code)
	}
	if code := c.do("DELETE", "/api/rewards/"+itoa(byName["Big treat"].ID), nil, nil); code != 200 {
		t.Errorf("remove reward = %d", code)
	}
	c.do("GET", "/api/rewards", nil, &rewards)
	for _, rw := range rewards {
		if rw.Name == "Big treat" {
			t.Error("removed reward still listed")
		}
		if rw.Name == "1 hour of gaming" && (rw.Cost == nil || *rw.Cost != 60) {
			t.Error("price update didn't stick")
		}
	}

	var hist struct {
		Attempts    []store.Attempt    `json:"attempts"`
		Redemptions []store.Redemption `json:"redemptions"`
	}
	c.do("GET", "/api/history", nil, &hist)
	if len(hist.Attempts) != 4 || len(hist.Redemptions) != 2 {
		t.Errorf("history: %d attempts, %d redemptions; want 4, 2", len(hist.Attempts), len(hist.Redemptions))
	}
	var stats []store.WeekStat
	c.do("GET", "/api/stats/mediums", nil, &stats)
	last := stats[len(stats)-1]
	if len(stats) != 8 || last.Solved != 2 || last.MedianMinutes == nil || *last.MedianMinutes != 27 {
		t.Errorf("medium stats last week = %+v", last)
	}
}

func TestGoodWeeksAndMilestones(t *testing.T) {
	c, now, _ := setup(t)
	login(t, c)
	monday := *now // Monday Oct 19, 8 PM IST

	logOn := func(day int, hour int) store.Result {
		t.Helper()
		*now = monday.AddDate(0, 0, day).Add(time.Duration(hour-20) * time.Hour)
		var r store.Result
		if code := c.do("POST", "/api/chess", nil, &r); code != 201 {
			t.Fatalf("chess = %d", code)
		}
		return r
	}

	// Four active days: no bonus yet. Tuesday is skipped.
	logOn(0, 20)
	logOn(2, 20)
	logOn(3, 20)
	// Friday 1 AM IST counts for Thursday, which already has activity.
	if r := logOn(4, 1); r.Earned != 10 {
		t.Errorf("Thursday late night earned %d, want 10 (no bonus)", r.Earned)
	}
	s := summary(t, c)
	if s.Week.ActiveCount != 3 || s.Week.SkipPassesLeft != 1 {
		t.Errorf("week = %d active, %d passes; want 3, 1", s.Week.ActiveCount, s.Week.SkipPassesLeft)
	}
	logOn(4, 21)
	r := logOn(5, 20) // fifth active day: Saturday
	if r.Earned != 60 {
		t.Errorf("fifth day earned %d, want 60 (10 + 50 bonus)", r.Earned)
	}
	if r := logOn(6, 20); r.Earned != 10 {
		t.Errorf("sixth day earned %d, want 10 (bonus only once)", r.Earned)
	}
	s = summary(t, c)
	if !s.Week.GoodWeek || s.GoodWeeks != 1 {
		t.Errorf("good week = %v, total %d", s.Week.GoodWeek, s.GoodWeeks)
	}

	steam := s.Milestones[0]
	if steam.Name != "Steam game" || steam.Unlocked {
		t.Fatalf("first milestone = %+v", steam)
	}
	if code := c.do("POST", "/api/milestones/"+itoa(steam.ID)+"/claim", nil, nil); code != 409 {
		t.Errorf("claim while locked = %d, want 409", code)
	}

	// Two more good weeks, with a bad week in between that erases nothing.
	for _, week := range []int{1, 3} {
		for d := 0; d < 5; d++ {
			logOn(7*week+d, 20)
		}
	}
	logOn(14, 20) // week 2: just one day
	s = summary(t, c)
	if s.GoodWeeks != 3 || !s.Milestones[0].Unlocked || s.Milestones[1].Unlocked {
		t.Fatalf("after 3 good weeks: total %d, milestones %+v", s.GoodWeeks, s.Milestones)
	}
	if code := c.do("POST", "/api/milestones/"+itoa(steam.ID)+"/claim", nil, nil); code != 200 {
		t.Errorf("claim = %d", code)
	}
	if code := c.do("POST", "/api/milestones/"+itoa(steam.ID)+"/claim", nil, nil); code != 409 {
		t.Errorf("claim twice = %d, want 409", code)
	}
	// 18 chess sessions x 10 + 3 bonuses x 50 = 330.
	if s = summary(t, c); s.Balance != 330 {
		t.Errorf("balance = %d, want 330", s.Balance)
	}
}

func TestDoubleTapCantDoubleSpend(t *testing.T) {
	c, _, _ := setup(t)
	login(t, c)
	for i := 0; i < 3; i++ { // 3 x 10 = 30: enough for exactly one 30-point reward
		c.do("POST", "/api/chess", nil, nil)
	}
	var rewards []store.Reward
	c.do("GET", "/api/rewards", nil, &rewards)
	id := itoa(rewards[0].ID) // cheapest: 30 min of gaming, 30 points

	var wg sync.WaitGroup
	codes := make([]int, 6)
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i] = c.do("POST", "/api/rewards/"+id+"/buy", nil, nil)
		}(i)
	}
	wg.Wait()
	ok := 0
	for _, code := range codes {
		if code == 200 {
			ok++
		}
	}
	if ok != 1 {
		t.Errorf("%d of 6 simultaneous buys succeeded, want exactly 1 (codes %v)", ok, codes)
	}
	if s := summary(t, c); s.Balance != 0 {
		t.Errorf("balance = %d, want 0", s.Balance)
	}
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
