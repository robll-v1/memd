package core_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/robll-v1/memd/internal/core"
	"github.com/robll-v1/memd/internal/embed"
	"github.com/robll-v1/memd/internal/store/sqlite"
)

func newTestService(t *testing.T) *core.Service {
	t.Helper()
	repo, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "memd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return core.NewService(repo, embed.New(embed.Config{}))
}

func TestCorrectMakesNewCanonical(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	created, err := svc.Store(ctx, core.CreateMemoryRequest{
		WorkspaceID: "ws1",
		AgentID:     "codex",
		Kind:        core.KindFact,
		Content:     "port is 6001",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.Correct(ctx, created.ID, core.CorrectMemoryRequest{
		WorkspaceID: "ws1",
		AgentID:     "codex",
		Content:     "port is 6002",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OldMemory.SupersededBy != result.NewMemory.ID {
		t.Fatalf("old memory superseded_by = %q, want %q", result.OldMemory.SupersededBy, result.NewMemory.ID)
	}
	search, err := svc.Retrieve(ctx, core.SearchRequest{WorkspaceID: "ws1", AgentID: "codex", Query: "port", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Items) == 0 {
		t.Fatal("expected retrieval results")
	}
	if search.Items[0].Memory.Content != "port is 6002" {
		t.Fatalf("top result = %q, want corrected value", search.Items[0].Memory.Content)
	}
}

func TestDedupPreviewAndApplyExactDuplicates(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	first, err := svc.Store(ctx, core.CreateMemoryRequest{WorkspaceID: "ws1", AgentID: "codex", Kind: core.KindFact, Content: "MatrixOne runs locally"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Store(ctx, core.CreateMemoryRequest{WorkspaceID: "ws1", AgentID: "codex", Kind: core.KindFact, Content: "MatrixOne runs locally"})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := svc.DedupPreview(ctx, core.DedupPreviewRequest{WorkspaceID: "ws1", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Exact) != 1 {
		t.Fatalf("exact groups = %d, want 1", len(preview.Exact))
	}
	apply, err := svc.DedupApply(ctx, core.DedupApplyRequest{WorkspaceID: "ws1", MemoryIDs: []string{first.ID, second.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if len(apply.DeletedIDs) != 1 {
		t.Fatalf("deleted ids = %d, want 1", len(apply.DeletedIDs))
	}
	previewAfter, err := svc.DedupPreview(ctx, core.DedupPreviewRequest{WorkspaceID: "ws1", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(previewAfter.Exact) != 0 {
		t.Fatalf("exact groups after apply = %d, want 0", len(previewAfter.Exact))
	}
}

func TestRetrievePrefersSameAgent(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	if _, err := svc.Store(ctx, core.CreateMemoryRequest{WorkspaceID: "ws1", AgentID: "alpha", Kind: core.KindFact, Content: "database port 6001"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Store(ctx, core.CreateMemoryRequest{WorkspaceID: "ws1", AgentID: "beta", Kind: core.KindFact, Content: "database port 6001"}); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Retrieve(ctx, core.SearchRequest{WorkspaceID: "ws1", AgentID: "alpha", Query: "database port", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) < 2 {
		t.Fatalf("got %d results, want >= 2", len(result.Items))
	}
	if result.Items[0].Memory.AgentID != "alpha" {
		t.Fatalf("top agent = %q, want alpha", result.Items[0].Memory.AgentID)
	}
}
