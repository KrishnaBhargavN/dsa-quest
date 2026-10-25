package api

import (
	"net/http"
	"time"

	"dsaquest/internal/game"
	"dsaquest/internal/lessons"
	"dsaquest/internal/store"
)

func (s *Server) problems(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.Problems(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) createProblem(w http.ResponseWriter, r *http.Request) {
	var in store.ProblemInput
	if err := readJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	id, err := s.store.CreateProblem(r.Context(), in, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) updateProblem(w http.ResponseWriter, r *http.Request) {
	var in store.ProblemInput
	if err := readJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := s.store.UpdateProblem(r.Context(), r.PathValue("id"), in); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) removeProblem(w http.ResponseWriter, r *http.Request) {
	if err := s.store.RemoveProblem(r.Context(), r.PathValue("id")); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) doneBefore(on bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.store.SetDoneBefore(r.Context(), r.PathValue("id"), on, s.now()); err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// revisions lists the revisions due in [from, to]; it defaults to the
// six weeks around today, enough for any month grid.
func (s *Server) revisions(w http.ResponseWriter, r *http.Request) {
	today := game.Day(s.now())
	from, to := today.AddDate(0, 0, -21), today.AddDate(0, 0, 21)
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad from date")
			return
		}
		from = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil || t.Before(from) || t.Sub(from) > 62*24*time.Hour {
			writeErr(w, http.StatusBadRequest, "bad to date")
			return
		}
		to = t
	}
	revs, err := s.store.Revisions(r.Context(), from, to)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"today": today.Format("2006-01-02"), "revisions": revs})
}

func (s *Server) scheduleRevision(w http.ResponseWriter, r *http.Request) {
	var in store.RevisionInput
	if err := readJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	rev, err := s.store.ScheduleRevision(r.Context(), in, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rev)
}

func (s *Server) moveRevision(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	var b struct {
		Due string `json:"due"`
	}
	if err := readJSON(r, &b); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := s.store.MoveRevision(r.Context(), id, b.Due, s.now()); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) completeRevision(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	if err := s.store.CompleteRevision(r.Context(), id, s.now()); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) deleteRevision(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	if err := s.store.DeleteRevision(r.Context(), id); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) lessonIndex(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, lessons.Index())
}

func (s *Server) lesson(w http.ResponseWriter, r *http.Request) {
	l, ok := lessons.Get(r.PathValue("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "no such lesson")
		return
	}
	writeJSON(w, http.StatusOK, l)
}
