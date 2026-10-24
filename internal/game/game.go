// Package game holds every rule of DSA Quest: how many points a problem is
// worth, when a day starts, and what counts as a good week. It has no database
// or HTTP code, so the rules can be read, tested and retuned in one place.
package game

import (
	"math/rand/v2"
	"sort"
	"strings"
	"time"
)

// Difficulties.
const (
	Easy   = "easy"
	Medium = "medium"
	Hard   = "hard"
)

// Outcomes.
const (
	Solved = "solved"
	Stuck  = "stuck"
)

// Ledger reasons.
const (
	ReasonSolve     = "solve"
	ReasonFastBonus = "fast_bonus"
	ReasonAttempt   = "attempt"
	ReasonChess     = "chess"
	ReasonGoodWeek  = "good_week"
	ReasonBuy       = "buy"
	ReasonMystery   = "mystery"
)

// Tunable numbers. Change these and redeploy to retune the game.
const (
	EasyPoints       = 5
	MediumPoints     = 15
	HardPoints       = 30
	FastMediumBonus  = 10
	FastMediumLimit  = 25 * time.Minute
	AttemptPoints    = 5 // honest attempt that didn't land
	ChessPoints      = 10
	GoodWeekBonus    = 50
	GoodWeekDays     = 5 // 7-day target with 2 skip passes
	SkipPasses       = 2
	MysteryCost      = 40
	DayRolloverHours = 4 // a day runs 4 AM to 4 AM IST
)

// IST is India Standard Time. India has no daylight saving, so a fixed zone is exact.
var IST = time.FixedZone("IST", 5*60*60+30*60)

// Award is one line of points earned.
type Award struct {
	Reason string `json:"reason"`
	Points int    `json:"points"`
}

// ValidDifficulty reports whether d is easy, medium or hard.
func ValidDifficulty(d string) bool { return d == Easy || d == Medium || d == Hard }

// ProblemAwards returns the points for finishing a problem.
func ProblemAwards(difficulty, outcome string, took time.Duration) []Award {
	if outcome != Solved {
		return []Award{{ReasonAttempt, AttemptPoints}}
	}
	var awards []Award
	switch difficulty {
	case Easy:
		awards = append(awards, Award{ReasonSolve, EasyPoints})
	case Medium:
		awards = append(awards, Award{ReasonSolve, MediumPoints})
		if took > 0 && took < FastMediumLimit {
			awards = append(awards, Award{ReasonFastBonus, FastMediumBonus})
		}
	case Hard:
		awards = append(awards, Award{ReasonSolve, HardPoints})
	}
	return awards
}

// Day returns the game day an instant belongs to, as midnight UTC of that
// calendar date. 1:30 AM IST on Tuesday still counts as Monday.
func Day(t time.Time) time.Time {
	local := t.In(IST).Add(-DayRolloverHours * time.Hour)
	y, m, d := local.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// WeekStart returns the Monday of the game week that day belongs to.
func WeekStart(day time.Time) time.Time {
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	return day.AddDate(0, 0, -offset)
}

// WeekBounds returns the real instants a game week covers:
// Monday 4 AM IST up to the next Monday 4 AM IST.
func WeekBounds(weekStart time.Time) (from, to time.Time) {
	y, m, d := weekStart.Date()
	from = time.Date(y, m, d, DayRolloverHours, 0, 0, 0, IST)
	return from, from.AddDate(0, 0, 7)
}

// DayBounds returns the real instants a game day covers: 4 AM IST to 4 AM IST.
func DayBounds(day time.Time) (from, to time.Time) {
	y, m, d := day.Date()
	from = time.Date(y, m, d, DayRolloverHours, 0, 0, 0, IST)
	return from, from.AddDate(0, 0, 1)
}

// ActiveDays maps each instant to its game day and returns the distinct days.
func ActiveDays(times []time.Time) map[time.Time]bool {
	days := make(map[time.Time]bool, len(times))
	for _, t := range times {
		days[Day(t)] = true
	}
	return days
}

// IsGoodWeek reports whether this many active days makes a good week.
func IsGoodWeek(activeDays int) bool { return activeDays >= GoodWeekDays }

// SkipPassesLeft returns how many skip passes remain this week, given the
// active days so far and today's game day. Days before today that had no
// activity use up a pass; today doesn't count until it's over.
func SkipPassesLeft(weekStart, today time.Time, active map[time.Time]bool) int {
	used := 0
	for d := weekStart; d.Before(today); d = d.AddDate(0, 0, 1) {
		if !active[d] {
			used++
		}
	}
	left := SkipPasses - used
	if left < 0 {
		return 0
	}
	return left
}

// Prize is something the mystery box can draw.
type Prize struct {
	ID     int64
	Weight int
}

// DrawMystery picks a prize with probability proportional to its weight.
// It returns false when nothing in the pool has a positive weight.
func DrawMystery(pool []Prize, r *rand.Rand) (int64, bool) {
	total := 0
	for _, p := range pool {
		if p.Weight > 0 {
			total += p.Weight
		}
	}
	if total == 0 {
		return 0, false
	}
	var pick int
	if r == nil {
		pick = rand.IntN(total)
	} else {
		pick = r.IntN(total)
	}
	for _, p := range pool {
		if p.Weight <= 0 {
			continue
		}
		if pick < p.Weight {
			return p.ID, true
		}
		pick -= p.Weight
	}
	return 0, false
}

// Median returns the median of the values, or 0 for none.
func Median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	s := append([]float64(nil), vals...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// URLKey turns a problem link into a stable key, so the same problem matches
// however the link was copied: leetcode:<slug>, cses:<task number>, or the
// trimmed lower-cased link for anything else.
func URLKey(raw string) string {
	u := strings.ToLower(strings.TrimSpace(raw))
	if u == "" {
		return ""
	}
	if i := strings.Index(u, "leetcode.com/problems/"); i >= 0 {
		parts := strings.FieldsFunc(u[i+len("leetcode.com/problems/"):], func(r rune) bool { return r == '/' || r == '?' || r == '#' })
		if len(parts) > 0 {
			return "leetcode:" + parts[0]
		}
	}
	if i := strings.Index(u, "cses.fi/problemset/task/"); i >= 0 {
		n := strings.FieldsFunc(u[i+len("cses.fi/problemset/task/"):], func(r rune) bool { return r < '0' || r > '9' })
		if len(n) > 0 {
			return "cses:" + n[0]
		}
	}
	return strings.TrimRight(u, "/")
}

// SlowMediumLimit is how long a solved medium can take before it goes on the
// redo list.
const SlowMediumLimit = 40 * time.Minute

// NeedsRedo reports whether a problem whose latest finished attempt had this
// outcome, difficulty and duration belongs on the redo list.
func NeedsRedo(outcome, difficulty string, took time.Duration) bool {
	return outcome == Stuck || (outcome == Solved && difficulty == Medium && took > SlowMediumLimit)
}

// RevisionDue returns the game day a revision scheduled now falls on, inDays later.
func RevisionDue(now time.Time, inDays int) time.Time {
	return Day(now).AddDate(0, 0, inDays)
}
