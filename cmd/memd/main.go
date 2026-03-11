package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mcpapi "github.com/robll-v1/memd/internal/api/mcp"
	"github.com/robll-v1/memd/internal/api/rest"
	"github.com/robll-v1/memd/internal/config"
	"github.com/robll-v1/memd/internal/core"
	"github.com/robll-v1/memd/internal/embed"
	"github.com/robll-v1/memd/internal/store/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch os.Args[1] {
	case "serve":
		must(runServe(ctx, os.Args[2:]))
	case "mcp":
		must(runMCP(ctx, os.Args[2:]))
	case "doctor":
		must(runDoctor(ctx, os.Args[2:]))
	default:
		usage()
		os.Exit(2)
	}
}

func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	dbPath := fs.String("db-path", config.DefaultDBPath(), "SQLite database path")
	addr := fs.String("addr", "127.0.0.1:8081", "REST listen address")
	embedFlags := bindEmbedFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	svc, cleanup, err := buildService(ctx, *dbPath, embedFlags.Config())
	if err != nil {
		return err
	}
	defer cleanup()
	server := rest.New(svc)
	fmt.Fprintf(os.Stdout, "memd REST listening on %s\n", *addr)
	return rest.Run(ctx, *addr, server)
}

func runMCP(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	dbPath := fs.String("db-path", config.DefaultDBPath(), "SQLite database path")
	agent := fs.String("agent", "codex", "Default agent ID")
	workspaceID := fs.String("workspace-id", "", "Explicit workspace ID override")
	workspaceDir := fs.String("workspace-dir", ".", "Workspace directory used to derive default workspace ID")
	embedFlags := bindEmbedFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	svc, cleanup, err := buildService(ctx, *dbPath, embedFlags.Config())
	if err != nil {
		return err
	}
	defer cleanup()
	return mcpapi.Run(ctx, svc, mcpapi.Options{AgentID: *agent, WorkspaceDir: *workspaceDir, WorkspaceID: *workspaceID})
}

func runDoctor(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	dbPath := fs.String("db-path", config.DefaultDBPath(), "SQLite database path")
	workspaceID := fs.String("workspace-id", "", "Optional workspace ID")
	embedFlags := bindEmbedFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	svc, cleanup, err := buildService(ctx, *dbPath, embedFlags.Config())
	if err != nil {
		return err
	}
	defer cleanup()
	health, err := svc.Health(ctx, core.HealthRequest{WorkspaceID: *workspaceID})
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(append(raw, '\n'))
	return err
}

func buildService(ctx context.Context, dbPath string, embedCfg embed.Config) (*core.Service, func(), error) {
	repo, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		return nil, nil, err
	}
	embedder := embed.New(embedCfg)
	return core.NewService(repo, embedder), func() { _ = repo.Close() }, nil
}

type embedFlagSet struct {
	provider *string
	baseURL  *string
	apiKey   *string
	model    *string
	dims     *int
}

func bindEmbedFlags(fs *flag.FlagSet) embedFlagSet {
	return embedFlagSet{
		provider: fs.String("embed-provider", "", "Embedding provider (openai)"),
		baseURL:  fs.String("embed-base-url", "", "OpenAI-compatible embeddings base URL"),
		apiKey:   fs.String("embed-api-key", "", "Embedding API key"),
		model:    fs.String("embed-model", "", "Embedding model"),
		dims:     fs.Int("embed-dims", 0, "Embedding dimensions"),
	}
}

func (e embedFlagSet) Config() embed.Config {
	return embed.Config{
		Provider: value(e.provider),
		BaseURL:  value(e.baseURL),
		APIKey:   value(e.apiKey),
		Model:    value(e.model),
		Dims:     valueInt(e.dims),
	}
}

func value(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func valueInt(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, "memd <serve|mcp|doctor> [flags]")
}
