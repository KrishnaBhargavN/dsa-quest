package store

import (
	"testing"

	"dsaquest/internal/game"
)

func TestCatalogue(t *testing.T) {
	rows, err := catalogue()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 151 {
		t.Fatalf("got %d problems, want 151", len(rows))
	}
	ids, keys := map[string]bool{}, map[string]bool{}
	for i, r := range rows {
		k := game.URLKey(r.URL)
		switch {
		case r.Order != i+1:
			t.Errorf("%s: order %d at position %d", r.ID, r.Order, i+1)
		case ids[r.ID]:
			t.Errorf("duplicate id %s", r.ID)
		case k == "" || keys[k]:
			t.Errorf("%s: missing or duplicate link key %q", r.ID, k)
		case r.Description == "" || len(r.Hints) != 2 || r.Hints[0] == "" || r.Hints[1] == "":
			t.Errorf("%s: needs a description and two hints", r.ID)
		case !game.ValidDifficulty(r.Difficulty):
			t.Errorf("%s: bad difficulty %q", r.ID, r.Difficulty)
		case r.Topic == "" || r.Phase < 1 || r.Phase > 3:
			t.Errorf("%s: bad topic or phase", r.ID)
		}
		ids[r.ID], keys[k] = true, true
	}
}
