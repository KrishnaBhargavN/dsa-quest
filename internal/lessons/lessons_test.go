package lessons

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLessonsLoad(t *testing.T) {
	if err := Err(); err != nil {
		t.Fatal(err)
	}
	orders := map[int]string{}
	for _, m := range Index() {
		if prev, dup := orders[m.Order]; dup {
			t.Errorf("%s and %s share order %d", prev, m.ID, m.Order)
		}
		orders[m.Order] = m.ID
		l, _ := Get(m.ID)
		if m.Summary == "" || m.Minutes == 0 || len(l.Body) < 500 {
			t.Errorf("%s: needs a summary, minutes and a real body", m.ID)
		}
		if strings.Count(l.Body, "```")%2 != 0 {
			t.Errorf("%s: unbalanced code fence", m.ID)
		}
	}
}

// Every topic on the problem list must have a lesson.
func TestEveryProblemTopicHasALesson(t *testing.T) {
	raw, err := os.ReadFile("../store/problems.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Topic string }
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	missing := map[string]bool{}
	for _, r := range rows {
		if ForTopic(r.Topic) == "" {
			missing[r.Topic] = true
		}
	}
	for topic := range missing {
		t.Errorf("no lesson teaches %q", topic)
	}
}
