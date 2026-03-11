# Codex smoke test

1. Start MCP mode:

```bash
go run ./cmd/memd mcp --agent codex
```

2. Configure Codex to point at the `memd mcp` command.

3. Run:

- `memory_store`
- `memory_retrieve`
- `memory_dedup_preview`
- `memory_dedup_apply`

4. Verify that identical memories collapse to one active canonical record.

