package mcpapi

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moukeyu/memd/internal/core"
	"github.com/moukeyu/memd/internal/embed"
	"github.com/moukeyu/memd/internal/store/sqlite"
)

func TestMCPStoreAndRetrieve(t *testing.T) {
	ctx := context.Background()
	repo, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "memd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	svc := core.NewService(repo, embed.New(embed.Config{}))
	server, err := BuildServer(svc, Options{AgentID: "codex", WorkspaceID: "ws_test"})
	if err != nil {
		t.Fatal(err)
	}
	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "v0.0.1"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "memory_store", Arguments: map[string]any{"kind": "fact", "content": "local MO port 6001"}}); err != nil {
		t.Fatal(err)
	}
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "memory_retrieve", Arguments: map[string]any{"query": "port", "limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Content) == 0 {
		t.Fatal("expected MCP content")
	}
}
