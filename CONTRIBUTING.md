# Contributing

## Development

```bash
go test ./...
go run ./cmd/memd doctor
go run ./cmd/memd serve
```

## Project principles

- Keep `memd` local-first and lightweight by default.
- Do not make embeddings mandatory for correctness.
- Prefer explicit correction and exact-dedup over aggressive automatic merge.
- Keep REST and MCP thin adapters over the shared core service.

## Release checklist

- Run `gofmt -w ./cmd ./internal`
- Run `go test ./...`
- Build `go build ./...`
- Update docs if public API or CLI flags change
