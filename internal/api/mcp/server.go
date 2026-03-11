package mcpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/robll-v1/memd/internal/config"
	"github.com/robll-v1/memd/internal/core"
)

type Options struct {
	AgentID      string
	WorkspaceDir string
	WorkspaceID  string
}

func Run(ctx context.Context, service *core.Service, opts Options) error {
	server, err := BuildServer(service, opts)
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}

func BuildServer(service *core.Service, opts Options) (*mcp.Server, error) {
	workspaceDir := opts.WorkspaceDir
	if workspaceDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workspaceDir = cwd
	}
	defaultWorkspace := config.ResolveWorkspaceID(opts.WorkspaceID, workspaceDir)
	defaultAgent := opts.AgentID
	if defaultAgent == "" {
		defaultAgent = "codex"
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "memd", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "memory_store", Description: "Store a new memory"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string          `json:"workspace_id,omitempty"`
		AgentID     string          `json:"agent_id,omitempty"`
		SessionID   string          `json:"session_id,omitempty"`
		Kind        core.MemoryKind `json:"kind"`
		Content     string          `json:"content"`
		Source      string          `json:"source,omitempty"`
		Confidence  float64         `json:"confidence,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		memory, err := service.Store(ctx, core.CreateMemoryRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			SessionID:   args.SessionID,
			Kind:        args.Kind,
			Content:     args.Content,
			Source:      args.Source,
			Confidence:  args.Confidence,
		})
		return jsonResult(memory), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_retrieve", Description: "Retrieve relevant memories for a query"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string          `json:"workspace_id,omitempty"`
		AgentID     string          `json:"agent_id,omitempty"`
		SessionID   string          `json:"session_id,omitempty"`
		Kind        core.MemoryKind `json:"kind,omitempty"`
		Query       string          `json:"query"`
		Limit       int             `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Retrieve(ctx, core.SearchRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			SessionID:   args.SessionID,
			Kind:        args.Kind,
			Query:       args.Query,
			Limit:       args.Limit,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_search", Description: "Search memories by keyword"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string          `json:"workspace_id,omitempty"`
		AgentID     string          `json:"agent_id,omitempty"`
		Kind        core.MemoryKind `json:"kind,omitempty"`
		Query       string          `json:"query,omitempty"`
		Limit       int             `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Search(ctx, core.SearchRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			Kind:        args.Kind,
			Query:       args.Query,
			Limit:       args.Limit,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_correct", Description: "Correct an existing memory"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		MemoryID    string  `json:"memory_id"`
		WorkspaceID string  `json:"workspace_id,omitempty"`
		AgentID     string  `json:"agent_id,omitempty"`
		SessionID   string  `json:"session_id,omitempty"`
		Content     string  `json:"content"`
		Source      string  `json:"source,omitempty"`
		Confidence  float64 `json:"confidence,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Correct(ctx, args.MemoryID, core.CorrectMemoryRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			SessionID:   args.SessionID,
			Content:     args.Content,
			Source:      args.Source,
			Confidence:  args.Confidence,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_delete", Description: "Soft-delete a memory"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		MemoryID    string `json:"memory_id"`
		WorkspaceID string `json:"workspace_id,omitempty"`
		AgentID     string `json:"agent_id,omitempty"`
		Reason      string `json:"reason,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Delete(ctx, args.MemoryID, core.DeleteMemoryRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			Reason:      args.Reason,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_dedup_preview", Description: "Preview duplicate memory groups"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string          `json:"workspace_id,omitempty"`
		AgentID     string          `json:"agent_id,omitempty"`
		Kind        core.MemoryKind `json:"kind,omitempty"`
		Limit       int             `json:"limit,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.DedupPreview(ctx, core.DedupPreviewRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			Kind:        args.Kind,
			Limit:       args.Limit,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_dedup_apply", Description: "Apply exact duplicate cleanup"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string   `json:"workspace_id,omitempty"`
		AgentID     string   `json:"agent_id,omitempty"`
		CanonicalID string   `json:"canonical_id,omitempty"`
		MemoryIDs   []string `json:"memory_ids"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.DedupApply(ctx, core.DedupApplyRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
			CanonicalID: args.CanonicalID,
			MemoryIDs:   args.MemoryIDs,
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_profile", Description: "Summarize memories for the workspace and agent"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string `json:"workspace_id,omitempty"`
		AgentID     string `json:"agent_id,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Profile(ctx, core.ProfileRequest{
			WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace),
			AgentID:     defaultString(args.AgentID, defaultAgent),
		})
		return jsonResult(result), nil, err
	})

	mcp.AddTool(server, &mcp.Tool{Name: "memory_health", Description: "Show health information for the local memory store"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		WorkspaceID string `json:"workspace_id,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		result, err := service.Health(ctx, core.HealthRequest{WorkspaceID: defaultString(args.WorkspaceID, defaultWorkspace)})
		return jsonResult(result), nil, err
	})

	return server, nil
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func jsonResult(payload any) *mcp.CallToolResult {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		raw = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}
}
