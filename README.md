# memd

Local-first memory runtime for AI agents.

## Features

- SQLite default store (`~/.memd/memd.db`)
- REST API
- stdio MCP server for Codex and similar agents
- explicit correction chain
- exact duplicate preview/apply
- optional OpenAI-compatible embedding rerank

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

## Notes

- `workspace_id` is required for REST callers
- MCP defaults `workspace_id` from the current working directory hash
- v1 only auto-applies exact duplicates

## Repository Docs

- design: `docs/DESIGN.md`
- publishing: `docs/PUBLISHING.md`
- contributing: `CONTRIBUTING.md`

