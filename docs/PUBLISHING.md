# Publishing

## Build

```bash
go build -o memd ./cmd/memd
```

## Local smoke checks

```bash
go test ./...
./memd doctor
./memd serve
./memd mcp --agent codex
```

## Codex MCP config example

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = ["mcp", "--agent", "codex"]
```

## GitHub release flow

1. Commit all changes.
2. Create or connect a public GitHub repository.
3. Push `main`.
4. Tag versions as `v0.x.y`.
5. Attach the compiled binary or publish a Homebrew / release artifact later.
