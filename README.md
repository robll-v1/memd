# memd

[![CI](https://github.com/robll-v1/memd/actions/workflows/ci.yml/badge.svg)](https://github.com/robll-v1/memd/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/robll-v1/memd)](./LICENSE)

Local-first memory runtime for AI agents.

## Features

- SQLite default store (`~/.memd/memd.db`)
- REST API
- stdio MCP server for Codex and similar agents
- explicit correction chain
- exact duplicate preview/apply
- optional OpenAI-compatible embedding rerank

## Quick Start

### Build

```bash
go build -o memd ./cmd/memd
```

### Inspect local health

```bash
./memd doctor
```

### Start the local REST server

```bash
./memd serve
```

### Store and retrieve one memory over REST

```bash
curl -X POST http://127.0.0.1:8081/v1/memories \
  -H 'Content-Type: application/json' \
  -d '{
    "workspace_id": "demo",
    "agent_id": "codex",
    "kind": "fact",
    "content": "MatrixOne runs locally on port 6001"
  }'

curl -X POST http://127.0.0.1:8081/v1/memories/retrieve \
  -H 'Content-Type: application/json' \
  -d '{
    "workspace_id": "demo",
    "agent_id": "codex",
    "query": "what port does MatrixOne use",
    "limit": 5
  }'
```

## Commands

```bash
go run ./cmd/memd serve
go run ./cmd/memd mcp --agent codex
go run ./cmd/memd doctor
```

## REST endpoints

- `POST /v1/memories`
- `POST /v1/memories/retrieve`
- `GET /v1/memories/search`
- `PUT /v1/memories/{id}/correct`
- `DELETE /v1/memories/{id}`
- `POST /v1/memories/dedup/preview`
- `POST /v1/memories/dedup/apply`
- `GET /v1/profile`
- `GET /v1/health`

## Codex MCP example

Add this to `~/.codex/config.toml`:

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = ["mcp", "--agent", "codex"]
```

Then start a new Codex session and use the MCP tools directly:

- `memory_store`
- `memory_retrieve`
- `memory_search`
- `memory_dedup_preview`
- `memory_dedup_apply`

## Release Notes

This repository includes:

- CI workflow: `.github/workflows/ci.yml`
- release workflow: `.github/workflows/release.yml`
- release note categories: `.github/release.yml`

Pushing a tag like `v0.1.1` will build binaries and attach them to the GitHub Release automatically.

## Notes

- `workspace_id` is required for REST callers
- MCP defaults `workspace_id` from the current working directory hash
- v1 only auto-applies exact duplicates

## Repository Docs

- design: `docs/DESIGN.md`
- publishing: `docs/PUBLISHING.md`
- releasing: `docs/RELEASING.md`
- contributing: `CONTRIBUTING.md`

