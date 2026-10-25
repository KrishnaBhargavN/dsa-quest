// Package lessons serves the theory lessons: one Markdown file per topic,
// embedded in the binary. Each file starts with a front matter block:
//
//	---
//	title: Sliding window
//	group: Core patterns
//	order: 5
//	topics: Sliding window
//	minutes: 12
//	summary: One line on what it is for.
//	---
//
// topics is a comma-separated list of problem topics (as in problems.json)
// that this lesson teaches.
package lessons

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.md
var files embed.FS

type Meta struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Group   string   `json:"group"`
	Order   int      `json:"order"`
	Topics  []string `json:"topics"`
	Minutes int      `json:"minutes"`
	Summary string   `json:"summary"`
}

type Lesson struct {
	Meta
	Body string `json:"body"`
}

var (
	all     []Lesson
	byID    = map[string]*Lesson{}
	byTopic = map[string]string{}
	loadErr error
)

func init() { loadErr = load() }

func load() error {
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		raw, err := files.ReadFile(e.Name())
		if err != nil {
			return err
		}
		l, err := parse(strings.TrimSuffix(e.Name(), ".md"), strings.ReplaceAll(string(raw), "\r\n", "\n"))
		if err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		all = append(all, l)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Order < all[j].Order })
	for i := range all {
		byID[all[i].ID] = &all[i]
		for _, t := range all[i].Topics {
			if _, taken := byTopic[t]; !taken {
				byTopic[t] = all[i].ID
			}
		}
	}
	return nil
}

func parse(id, raw string) (Lesson, error) {
	l := Lesson{Meta: Meta{ID: id, Topics: []string{}}}
	if !strings.HasPrefix(raw, "---\n") {
		return l, fmt.Errorf("missing front matter")
	}
	head, body, ok := strings.Cut(raw[4:], "\n---\n")
	if !ok {
		return l, fmt.Errorf("unterminated front matter")
	}
	for _, line := range strings.Split(head, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		switch strings.TrimSpace(k) {
		case "title":
			l.Title = v
		case "group":
			l.Group = v
		case "summary":
			l.Summary = v
		case "order":
			l.Order, _ = strconv.Atoi(v)
		case "minutes":
			l.Minutes, _ = strconv.Atoi(v)
		case "topics":
			for _, t := range strings.Split(v, ",") {
				if t = strings.TrimSpace(t); t != "" {
					l.Topics = append(l.Topics, t)
				}
			}
		}
	}
	if l.Title == "" || l.Group == "" || l.Order == 0 {
		return l, fmt.Errorf("front matter needs title, group and order")
	}
	l.Body = strings.TrimSpace(body)
	return l, nil
}

// Err reports a problem loading the embedded lessons.
func Err() error { return loadErr }

// Index lists every lesson without its body, in reading order.
func Index() []Meta {
	out := make([]Meta, len(all))
	for i, l := range all {
		out[i] = l.Meta
	}
	return out
}

// Get returns one lesson.
func Get(id string) (*Lesson, bool) {
	l, ok := byID[id]
	return l, ok
}

// ForTopic returns the lesson that teaches a problem topic, or "".
func ForTopic(topic string) string { return byTopic[topic] }
