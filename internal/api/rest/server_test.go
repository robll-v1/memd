package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/moukeyu/memd/internal/core"
	"github.com/moukeyu/memd/internal/embed"
	"github.com/moukeyu/memd/internal/store/sqlite"
)

func newRESTServer(t *testing.T) *Server {
	t.Helper()
	repo, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "memd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return New(core.NewService(repo, embed.New(embed.Config{})))
}

func TestCreateAndRetrieveOverREST(t *testing.T) {
	srv := newRESTServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	createBody := map[string]any{"workspace_id": "ws1", "agent_id": "codex", "kind": "fact", "content": "MatrixOne build command is make build"}
	raw, _ := json.Marshal(createBody)
	resp, err := http.Post(ts.URL+"/v1/memories", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}

	retrieveBody := map[string]any{"workspace_id": "ws1", "agent_id": "codex", "query": "build command", "limit": 5}
	raw, _ = json.Marshal(retrieveBody)
	resp, err = http.Post(ts.URL+"/v1/memories/retrieve", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("retrieve status = %d, want 200", resp.StatusCode)
	}
	var parsed core.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("retrieve items = %d, want 1", len(parsed.Items))
	}
}
