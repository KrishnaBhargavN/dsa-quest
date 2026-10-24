package game

import (
	"math/rand/v2"
	"testing"
	"time"
)

func ist(y int, m time.Month, d, hh, mm int) time.Time {
	return time.Date(y, m, d, hh, mm, 0, 0, IST)
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func total(a []Award) int {
	n := 0
	for _, x := range a {
		n += x.Points
	}
	return n
}

func TestProblemAwards(t *testing.T) {
	cases := []struct {
		name       string
		difficulty string
		outcome    string
		took       time.Duration
		want       int
	}{
		{"easy solved", Easy, Solved, 10 * time.Minute, 5},
		{"medium solved slow", Medium, Solved, 40 * time.Minute, 15},
		{"medium solved fast", Medium, Solved, 24 * time.Minute, 25},
		{"medium exactly 25 min gets no bonus", Medium, Solved, 25 * time.Minute, 15},
		{"hard solved", Hard, Solved, 90 * time.Minute, 30},
		{"hard fast gets no medium bonus", Hard, Solved, 10 * time.Minute, 30},
		{"stuck on a hard", Hard, Stuck, 60 * time.Minute, 5},
		{"stuck on an easy", Easy, Stuck, 5 * time.Minute, 5},
	}
	for _, c := range cases {
		got := total(ProblemAwards(c.difficulty, c.outcome, c.took))
		if got != c.want {
			t.Errorf("%s: got %d points, want %d", c.name, got, c.want)
		}
	}
}

func TestDayRollsOverAt4AMIST(t *testing.T) {
	cases := []struct {
		at   time.Time
		want time.Time
	}{
		{ist(2026, 10, 20, 23, 0), date(2026, 10, 20)},
		{ist(2026, 10, 21, 1, 30), date(2026, 10, 20)}, // late night counts for the evening before
		{ist(2026, 10, 21, 3, 59), date(2026, 10, 20)},
		{ist(2026, 10, 21, 4, 0), date(2026, 10, 21)},
		// 22:00 UTC on the 20th is 03:30 IST on the 21st: before 4 AM, so still the 20th.
		{time.Date(2026, 10, 20, 22, 0, 0, 0, time.UTC), date(2026, 10, 20)},
	}
	for _, c := range cases {
		if got := Day(c.at); !got.Equal(c.want) {
			t.Errorf("Day(%s) = %s, want %s", c.at, got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func TestWeekStartIsMonday(t *testing.T) {
	// 2026-10-19 is a Monday.
	for d := 19; d <= 25; d++ {
		if got := WeekStart(date(2026, 10, d)); !got.Equal(date(2026, 10, 19)) {
			t.Errorf("WeekStart(Oct %d) = %s, want 2026-10-19", d, got.Format("2006-01-02"))
		}
	}
	if got := WeekStart(date(2026, 10, 26)); !got.Equal(date(2026, 10, 26)) {
		t.Errorf("next Monday should start a new week, got %s", got.Format("2006-01-02"))
	}
}

func TestWeekBounds(t *testing.T) {
	from, to := WeekBounds(date(2026, 10, 19))
	if !from.Equal(ist(2026, 10, 19, 4, 0)) || !to.Equal(ist(2026, 10, 26, 4, 0)) {
		t.Errorf("bounds = %s .. %s", from, to)
	}
	// Sunday 2 AM IST after the week "ends" on the calendar still belongs to it.
	late := ist(2026, 10, 26, 2, 0)
	if !(late.Before(to) && !late.Before(from)) {
		t.Errorf("Monday 2 AM should still be inside the previous week")
	}
}

func TestGoodWeekAndSkipPasses(t *testing.T) {
	week := date(2026, 10, 19)
	active := ActiveDays([]time.Time{
		ist(2026, 10, 19, 21, 0),
		ist(2026, 10, 19, 22, 0), // same day twice counts once
		ist(2026, 10, 21, 1, 0),  // counts for the 20th
		ist(2026, 10, 22, 20, 0),
	})
	if len(active) != 3 {
		t.Fatalf("active days = %d, want 3", len(active))
	}
	if IsGoodWeek(len(active)) {
		t.Error("3 days should not be a good week")
	}
	// Today is Thursday the 22nd: Mon, Tue, Wed before it. Wed (21st) was missed.
	if got := SkipPassesLeft(week, date(2026, 10, 22), active); got != 1 {
		t.Errorf("skip passes left = %d, want 1", got)
	}
	if !IsGoodWeek(5) || IsGoodWeek(4) {
		t.Error("good week threshold should be 5 days")
	}
	// Missing three days uses more passes than you have; it floors at 0.
	if got := SkipPassesLeft(week, date(2026, 10, 25), map[time.Time]bool{}); got != 0 {
		t.Errorf("skip passes left = %d, want 0", got)
	}
}

func TestDrawMystery(t *testing.T) {
	pool := []Prize{{1, 40}, {2, 15}, {3, 0}, {4, 5}}
	r := rand.New(rand.NewPCG(1, 2))
	counts := map[int64]int{}
	for i := 0; i < 6000; i++ {
		id, ok := DrawMystery(pool, r)
		if !ok {
			t.Fatal("draw failed")
		}
		counts[id]++
	}
	if counts[3] != 0 {
		t.Error("a zero-weight prize was drawn")
	}
	if !(counts[1] > counts[2] && counts[2] > counts[4] && counts[4] > 0) {
		t.Errorf("draws don't follow the weights: %v", counts)
	}
	if _, ok := DrawMystery([]Prize{{1, 0}}, r); ok {
		t.Error("an empty pool should not draw")
	}
}

func TestMedian(t *testing.T) {
	if Median(nil) != 0 || Median([]float64{3, 1, 2}) != 2 || Median([]float64{4, 1, 3, 2}) != 2.5 {
		t.Error("median is wrong")
	}
}

func TestURLKey(t *testing.T) {
	cases := map[string]string{
		"https://leetcode.com/problems/Two-Sum/description/?envType=x": "leetcode:two-sum",
		"leetcode.com/problems/two-sum":                                "leetcode:two-sum",
		"https://www.leetcode.com/problems/two-sum/":                   "leetcode:two-sum",
		"https://cses.fi/problemset/task/1158/":                        "cses:1158",
		"https://cses.fi/problemset/result/1158":                       "https://cses.fi/problemset/result/1158",
		"  HTTPS://Example.com/x/  ":                                   "https://example.com/x",
		"":                                                             "",
		"https://leetcode.com/problems/":                                "https://leetcode.com/problems",
	}
	for in, want := range cases {
		if got := URLKey(in); got != want {
			t.Errorf("URLKey(%q) = %q, want %q", in, got, want)
		}
	}
}
