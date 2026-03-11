package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/robll-v1/memd/internal/core"
)

type Server struct {
	service *core.Service
	mux     *http.ServeMux
}

func New(service *core.Service) *Server {
	s := &Server{service: service, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/v1/health", s.handleHealth)
	s.mux.HandleFunc("/v1/profile", s.handleProfile)
	s.mux.HandleFunc("/v1/memories", s.handleMemories)
	s.mux.HandleFunc("/v1/memories/search", s.handleSearch)
	s.mux.HandleFunc("/v1/memories/retrieve", s.handleRetrieve)
	s.mux.HandleFunc("/v1/memories/dedup/preview", s.handleDedupPreview)
	s.mux.HandleFunc("/v1/memories/dedup/apply", s.handleDedupApply)
	s.mux.HandleFunc("/v1/memories/", s.handleMemoryByID)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	result, err := s.service.Health(r.Context(), core.HealthRequest{WorkspaceID: r.URL.Query().Get("workspace_id")})
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	result, err := s.service.Profile(r.Context(), core.ProfileRequest{
		WorkspaceID: r.URL.Query().Get("workspace_id"),
		AgentID:     r.URL.Query().Get("agent_id"),
	})
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleMemories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req core.CreateMemoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.service.Store(r.Context(), req)
	writeResult(w, map[string]any{"memory": result}, err, http.StatusCreated)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := s.service.Search(r.Context(), core.SearchRequest{
		WorkspaceID: r.URL.Query().Get("workspace_id"),
		AgentID:     r.URL.Query().Get("agent_id"),
		Query:       r.URL.Query().Get("q"),
		Kind:        core.MemoryKind(r.URL.Query().Get("kind")),
		Limit:       limit,
	})
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req core.SearchRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.service.Retrieve(r.Context(), req)
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleDedupPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req core.DedupPreviewRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.service.DedupPreview(r.Context(), req)
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleDedupApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req core.DedupApplyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := s.service.DedupApply(r.Context(), req)
	writeResult(w, result, err, http.StatusOK)
}

func (s *Server) handleMemoryByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/memories/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]
	if len(parts) == 2 && parts[1] == "correct" {
		if r.Method != http.MethodPut {
			writeMethodNotAllowed(w)
			return
		}
		var req core.CorrectMemoryRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		result, err := s.service.Correct(r.Context(), id, req)
		writeResult(w, result, err, http.StatusOK)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		result, err := s.service.Delete(r.Context(), id, core.DeleteMemoryRequest{
			WorkspaceID: r.URL.Query().Get("workspace_id"),
			AgentID:     r.URL.Query().Get("agent_id"),
			Reason:      r.URL.Query().Get("reason"),
		})
		writeResult(w, result, err, http.StatusOK)
		return
	}
	writeMethodNotAllowed(w)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeResult(w http.ResponseWriter, payload any, err error, successStatus int) {
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrInvalidArgument):
			status = http.StatusBadRequest
		case errors.Is(err, core.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, core.ErrNotExactDuplicates), errors.Is(err, core.ErrNoMemoriesProvided):
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, successStatus, payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Run(ctx context.Context, addr string, server *Server) error {
	httpServer := &http.Server{Addr: addr, Handler: server.Handler()}
	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()
	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
