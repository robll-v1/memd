# Codex Setup

This document explains how to use `memd` with both:

- Codex CLI
- Codex for VS Code

Both clients use the same MCP config file:

- `~/.codex/config.toml`

## 1. Build or install `memd`

If you build from source:

```bash
go build -o memd ./cmd/memd
```

Or download a release binary from:

- `https://github.com/robll-v1/memd/releases`

## 2. Add MCP config

Edit `~/.codex/config.toml` and add:

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

## 3. Start a new Codex session

Important:

- existing sessions do not hot-reload newly added MCP servers
- after changing config, open a brand new Codex session

## 4. Smoke test

Run these prompts in a new Codex session:

```text
Please call memory_health first.
```

```text
Please store this fact with memd: my MatrixOne port is 6001.
```

```text
Please retrieve what port my MatrixOne uses.
```

## 5. Exact duplicate test

Store the same fact twice, then ask:

```text
Please run memory_dedup_preview and show exact duplicate groups.
```

Then:

```text
Please run memory_dedup_apply on the exact duplicate group.
```

## 6. Optional embedding configuration

If you want semantic rerank, use embedding flags in the MCP command.

Example:

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = [
  "mcp",
  "--agent", "codex",
  "--embed-provider", "openai",
  "--embed-api-key", "sk-...",
  "--embed-model", "text-embedding-3-small"
]
```

## Notes

- MCP mode derives default `workspace_id` from the current working directory
- REST mode requires explicit `workspace_id`
- `memory_dedup_apply` only auto-applies exact duplicates in v1
