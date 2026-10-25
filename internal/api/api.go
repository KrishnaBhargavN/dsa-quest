// Package api turns HTTP requests into store calls and serves the frontend.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"dsaquest/internal/store"
)

type Server struct {
	store    *store.Store // nil when DATABASE_URL is missing
	password string
	static   fs.FS
	now      func() time.Time
}

func New(st *store.Store, password string, static fs.FS) *Server {
	return &Server{store: st, password: password, static: static, now: time.Now}
}

// SetClock lets tests control time.
func (s *Server) SetClock(now func() time.Time) { s.now = now }

func (s *Server) Handler() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/health", s.health)
	api.HandleFunc("POST /api/login", s.login)
	api.HandleFunc("POST /api/logout", s.logout)
	api.HandleFunc("GET /api/me", s.authed(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}))
	api.HandleFunc("GET /api/summary", s.authed(s.summary))
	api.HandleFunc("POST /api/attempts", s.authed(s.startAttempt))
	api.HandleFunc("POST /api/attempts/log", s.authed(s.logAttempt))
	api.HandleFunc("POST /api/attempts/{id}/finish", s.authed(s.finishAttempt))
	api.HandleFunc("DELETE /api/attempts/{id}", s.authed(s.discardAttempt))
	api.HandleFunc("POST /api/chess", s.authed(s.logChess))
	api.HandleFunc("GET /api/history", s.authed(s.history))
	api.HandleFunc("GET /api/stats/mediums", s.authed(s.mediumStats))
	api.HandleFunc("GET /api/rewards", s.authed(s.rewards))
	api.HandleFunc("POST /api/rewards", s.authed(s.createReward))
	api.HandleFunc("PUT /api/rewards/{id}", s.authed(s.updateReward))
	api.HandleFunc("DELETE /api/rewards/{id}", s.authed(s.removeReward))
	api.HandleFunc("POST /api/rewards/{id}/buy", s.authed(s.buy))
	api.HandleFunc("POST /api/mystery", s.authed(s.mystery))
	api.HandleFunc("POST /api/milestones/{id}/claim", s.authed(s.claimMilestone))
	api.HandleFunc("GET /api/problems", s.authed(s.problems))
	api.HandleFunc("POST /api/problems", s.authed(s.createProblem))
	api.HandleFunc("PUT /api/problems/{id}", s.authed(s.updateProblem))
	api.HandleFunc("DELETE /api/problems/{id}", s.authed(s.removeProblem))
	api.HandleFunc("POST /api/problems/{id}/before", s.authed(s.doneBefore(true)))
	api.HandleFunc("DELETE /api/problems/{id}/before", s.authed(s.doneBefore(false)))
	api.HandleFunc("GET /api/revisions", s.authed(s.revisions))
	api.HandleFunc("POST /api/revisions", s.authed(s.scheduleRevision))
	api.HandleFunc("PUT /api/revisions/{id}", s.authed(s.moveRevision))
	api.HandleFunc("POST /api/revisions/{id}/done", s.authed(s.completeRevision))
	api.HandleFunc("DELETE /api/revisions/{id}", s.authed(s.deleteRevision))
	api.HandleFunc("GET /api/lessons", s.authed(s.lessonIndex))
	api.HandleFunc("GET /api/lessons/{id}", s.authed(s.lesson))
	api.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, http.StatusNotFound, "no such endpoint")
	})

	root := http.NewServeMux()
	root.Handle("/api/", s.guard(api))
	root.HandleFunc("/", s.serveStatic)
	return root
}

// guard makes the database ready and blocks cross-site writes. Browsers can't
// add a custom header to a cross-site form post without a CORS preflight,
// which this server never approves.
func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet && r.Header.Get("X-DQ") != "1" {
			writeErr(w, http.StatusForbidden, "missing X-DQ header")
			return
		}
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		if s.store == nil {
			writeErr(w, http.StatusServiceUnavailable, "DATABASE_URL is not set on the server")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		if err := s.store.Ensure(ctx); err != nil {
			log.Printf("database setup failed: %v", err)
			writeErr(w, http.StatusServiceUnavailable, "database is not reachable yet, try again in a few seconds")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) authed(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err != nil || !validSession(s.store.Secret(), c.Value, s.now()) {
			writeErr(w, http.StatusUnauthorized, "log in first")
			return
		}
		h(w, r)
	}
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 16<<10))
	return dec.Decode(v)
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrBadInput):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrTimerRunning), errors.Is(err, store.ErrFinished),
		errors.Is(err, store.ErrNotEnough), errors.Is(err, store.ErrLocked),
		errors.Is(err, store.ErrClaimed), errors.Is(err, store.ErrEmptyMystery),
		errors.Is(err, store.ErrNotForSale), errors.Is(err, store.ErrDuplicate):
		writeErr(w, http.StatusConflict, err.Error())
	default:
		log.Printf("error: %v", err)
		writeErr(w, http.StatusInternalServerError, "something went wrong on the server")
	}
}

// ---------- handlers ----------

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{"passwordSet": s.password != "", "databaseURLSet": s.store != nil, "database": "unknown"}
	if s.store != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		if err := s.store.Ensure(ctx); err != nil {
			log.Printf("health: %v", err)
			out["database"] = "unreachable"
		} else {
			out["database"] = "ok"
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if s.password == "" {
		writeErr(w, http.StatusServiceUnavailable, "APP_PASSWORD is not set on the server")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	if !passwordMatches(body.Password, s.password) {
		time.Sleep(time.Second) // slows down guessing
		writeErr(w, http.StatusUnauthorized, "wrong password")
		return
	}
	setSessionCookie(w, r, newSession(s.store.Secret(), s.now()), int(sessionTTL.Seconds()))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	setSessionCookie(w, r, "", -1)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	sum, err := s.store.Summary(r.Context(), s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

type problemBody struct {
	ProblemID  string `json:"problemId"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Outcome    string `json:"outcome"`
	Minutes    int    `json:"minutes"`
	DaysAgo    int    `json:"daysAgo"`
}

func (b *problemBody) clean() {
	b.URL = strings.TrimSpace(b.URL)
	if len(b.URL) > 500 {
		b.URL = b.URL[:500]
	}
	b.Title = strings.TrimSpace(b.Title)
	if b.Title == "" {
		b.Title = titleFromURL(b.URL)
	}
	if len(b.Title) > 200 {
		b.Title = b.Title[:200]
	}
}

// titleFromURL turns https://leetcode.com/problems/two-sum/description/ into "Two Sum".
func titleFromURL(u string) string {
	i := strings.Index(u, "/problems/")
	if i < 0 {
		return ""
	}
	slug := strings.SplitN(u[i+len("/problems/"):], "/", 2)[0]
	slug = strings.SplitN(slug, "?", 2)[0]
	small := map[string]bool{"a": true, "an": true, "and": true, "at": true, "by": true, "for": true,
		"from": true, "in": true, "of": true, "on": true, "or": true, "the": true, "to": true, "with": true}
	words := strings.Split(slug, "-")
	for j, w := range words {
		if w != "" && (j == 0 || !small[w]) {
			words[j] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.TrimSpace(strings.Join(words, " "))
}

func (s *Server) startAttempt(w http.ResponseWriter, r *http.Request) {
	var b problemBody
	if err := readJSON(r, &b); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	b.clean()
	a, err := s.store.StartAttempt(r.Context(), b.ProblemID, b.URL, b.Title, b.Difficulty, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) logAttempt(w http.ResponseWriter, r *http.Request) {
	var b problemBody
	if err := readJSON(r, &b); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	b.clean()
	res, err := s.store.LogAttempt(r.Context(), b.ProblemID, b.URL, b.Title, b.Difficulty, b.Outcome, b.Minutes, b.DaysAgo, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (s *Server) finishAttempt(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	var b struct {
		Outcome string `json:"outcome"`
	}
	if err := readJSON(r, &b); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	res, err := s.store.FinishAttempt(r.Context(), id, b.Outcome, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) discardAttempt(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	if err := s.store.DiscardAttempt(r.Context(), id); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logChess(w http.ResponseWriter, r *http.Request) {
	res, err := s.store.LogChess(r.Context(), s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	attempts, redemptions, err := s.store.History(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"attempts": attempts, "redemptions": redemptions})
}

func (s *Server) mediumStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.MediumStats(r.Context(), s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) rewards(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.Rewards(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) createReward(w http.ResponseWriter, r *http.Request) {
	var in store.RewardInput
	if err := readJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := s.store.CreateReward(r.Context(), in); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (s *Server) updateReward(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	var in store.RewardInput
	if err := readJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := s.store.UpdateReward(r.Context(), id, in); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) removeReward(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	if err := s.store.RemoveReward(r.Context(), id); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) buy(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	res, err := s.store.Buy(r.Context(), id, s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) mystery(w http.ResponseWriter, r *http.Request) {
	res, err := s.store.Mystery(r.Context(), s.now())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) claimMilestone(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	if err := s.store.ClaimMilestone(r.Context(), id, s.now()); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---------- frontend ----------

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" {
		name = "index.html"
	}
	if st, err := fs.Stat(s.static, name); err != nil || st.IsDir() {
		name = "index.html" // single-page app: unknown paths get the app
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeFileFS(w, r, s.static, name)
}
