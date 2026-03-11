# memd

[![CI](https://github.com/robll-v1/memd/actions/workflows/ci.yml/badge.svg)](https://github.com/robll-v1/memd/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/robll-v1/memd)](./LICENSE)

Local-first memory runtime for AI agents.

中文 README: `README.zh-CN.md`

`memd` is a lightweight memory layer for coding agents. It stores durable memories in local SQLite, exposes a local REST API, and can be mounted into agent clients through MCP.

## Features

- SQLite default store (`~/.memd/memd.db`)
- single local binary
- REST API
- stdio MCP server for Codex and similar agents
- explicit correction chain
- exact duplicate preview/apply
- optional OpenAI-compatible embedding rerank
- no LLM required for core functionality

## Install

### Option A: Download a release binary

Download a binary from the GitHub Releases page:

- Releases: `https://github.com/robll-v1/memd/releases`

Then make it executable:

```bash
chmod +x ./memd-darwin-amd64
mv ./memd-darwin-amd64 ./memd
```

### Option B: Build from source

Requirements:

- Go 1.24+

```bash
git clone https://github.com/robll-v1/memd.git
cd memd
go build -o memd ./cmd/memd
```

## Quick Start

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

## Codex Setup

`Codex CLI` and `Codex for VS Code` share the same MCP configuration file:

- `~/.codex/config.toml`

You only need to configure `memd` once.

### Recommended MCP config

Add this to `~/.codex/config.toml`:

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = ["mcp", "--agent", "codex"]
```

Example:

```toml
[mcp_servers.memd]
command = "/Users/you/bin/memd"
args = ["mcp", "--agent", "codex"]
```

### Important behavior

- Do **not** manually start a long-running `memd mcp` process first.
- Let Codex launch the MCP server from `~/.codex/config.toml`.
- After editing `~/.codex/config.toml`, start a **new Codex session**.
- Existing Codex sessions will not hot-reload new MCP servers.

### Codex smoke test

After adding the MCP config, open a new Codex session and run these prompts in order:

1. health check

```text
Please call memory_health first and show me the result.
```

2. store one fact

```text
Please store this fact with memd: my MatrixOne port is 6001.
```

3. retrieve it

```text
Please retrieve what port my MatrixOne uses.
```

4. duplicate preview

Store the same fact twice, then ask:

```text
Please run memory_dedup_preview and tell me whether exact duplicates exist.
```

5. exact duplicate cleanup

```text
Please run memory_dedup_apply on the exact duplicate group and show which record was kept.
```

### Workspace behavior

In MCP mode, `memd` derives a default `workspace_id` from the current working directory.

That means:

- different project directories get different default memory spaces
- the same project directory will consistently reuse the same memory space
- REST callers must still pass `workspace_id` explicitly

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

## Notes

- `workspace_id` is required for REST callers
- MCP defaults `workspace_id` from the current working directory hash
- `agent_id` defaults to the `--agent` CLI flag in MCP mode
- v1 only auto-applies exact duplicates
- near-duplicate groups are preview-only in v1
- embedding is optional; core functionality works without it

## Troubleshooting

### Codex cannot see `memd`

- check `~/.codex/config.toml`
- make sure `command` points to the actual absolute binary path
- start a **new** Codex session after editing config

### `memory_health` works but retrieval is empty

- store at least one memory first
- make sure you are in the same project directory if relying on default MCP `workspace_id`
- for REST calls, verify `workspace_id` matches the one used when storing

### Duplicate cleanup does not delete near duplicates

This is expected in v1.

- `memory_dedup_apply` only auto-applies exact duplicates
- near duplicates are previewed, not automatically removed

### I want semantic rerank

Set embedding flags when running `memd` directly:

```bash
./memd serve \
  --embed-provider openai \
  --embed-api-key sk-... \
  --embed-model text-embedding-3-small
```

Use the same flags in your MCP command if you want Codex sessions to use embeddings.

## Repository Docs

- design: `docs/DESIGN.md`
- design (zh-CN): `docs/DESIGN.zh-CN.md`
- install: `docs/INSTALL.md`
- install (zh-CN): `docs/INSTALL.zh-CN.md`
- publishing: `docs/PUBLISHING.md`
- publishing (zh-CN): `docs/PUBLISHING.zh-CN.md`
- releasing: `docs/RELEASING.md`
- releasing (zh-CN): `docs/RELEASING.zh-CN.md`
- codex setup: `docs/CODEX.md`
- codex setup (zh-CN): `docs/CODEX.zh-CN.md`
- contributing: `CONTRIBUTING.md`
- contributing (zh-CN): `CONTRIBUTING.zh-CN.md`
