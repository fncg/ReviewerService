package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fncg/ReviewerService/internal/github"
	"github.com/fncg/ReviewerService/internal/storage"
	"github.com/fncg/ReviewerService/internal/telegram"
)

type Server struct {
	mux *http.ServeMux
	db  *storage.Postgres
	bot telegram.Notifier
}

func NewServer(db *storage.Postgres, bot telegram.Notifier) *Server {
	mux := http.NewServeMux()

	s := &Server{
		mux: mux,
		db:  db,
		bot: bot,
	}

	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/github/webhook", s.githubWebhook)

	return s
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "pull_request" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var event github.PullRequestEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Action != "opened" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.db.HandlePR(event, s.bot)

	log.Printf(
		"New PR opened: repo=%s title=%s author=%s",
		event.Repository.FullName,
		event.PullRequest.Title,
		event.PullRequest.User.Login,
	)

	w.WriteHeader(http.StatusOK)
}
