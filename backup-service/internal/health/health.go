package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"fcstask-backend/backup-service/internal/logging"
)

type Server struct {
	srv    *http.Server
	logger *logging.Logger

	mu          sync.RWMutex
	ready       bool
	lastSuccess time.Time
	lastError   string
}

type status struct {
	Status      string `json:"status"`
	Ready       bool   `json:"ready"`
	LastSuccess string `json:"last_success,omitempty"`
	Degraded    bool   `json:"degraded"`
}

func New(addr string, logger *logging.Logger) *Server {
	s := &Server{logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

func (s *Server) SetReady(v bool) {
	s.mu.Lock()
	s.ready = v
	s.mu.Unlock()
}

func (s *Server) RecordSuccess(t time.Time) {
	s.mu.Lock()
	s.lastSuccess = t
	s.lastError = ""
	s.mu.Unlock()
}

func (s *Server) RecordError(msg string) {
	s.mu.Lock()
	s.lastError = msg
	s.mu.Unlock()
}

func (s *Server) snapshot() status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := status{Ready: s.ready, Degraded: s.lastError != ""}
	if !s.lastSuccess.IsZero() {
		st.LastSuccess = s.lastSuccess.UTC().Format(time.RFC3339)
	}
	if s.ready {
		st.Status = "ok"
	} else {
		st.Status = "starting"
	}
	return st
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	st := s.snapshot()
	code := http.StatusOK
	if !st.Ready {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, st)
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	st := s.snapshot()
	code := http.StatusOK
	if !st.Ready {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, st)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Health server error: %v", err)
		}
	}()
	s.logger.Info("Health endpoint listening on %s", s.srv.Addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
