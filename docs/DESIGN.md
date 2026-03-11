# memd v1 Design

`memd` is a lightweight, local-first memory runtime for AI agents.

## Core properties

- Single Go binary
- SQLite local default store
- REST + MCP adapters over one shared core service
- Workspace + agent isolation
- Explicit correction and exact-duplicate cleanup
- Optional embedding-based rerank, never required

## v1 scope

- `store`
- `retrieve`
- `search`
- `correct`
- `delete`
- `dedup preview`
- `dedup apply`
- `profile`
- `health`

## Out of scope

- automatic ingest
- snapshot / branch / reflection
- graph retrieval
- multi-tenant shared service
- web UI

## Default identity model

- `workspace_id`: required for all data access
- `agent_id`: required on writes, optional on reads
- `session_id`: metadata only, not a primary isolation key

## Default storage layout

- `memories`
- `memory_events`
- `memories_fts` (SQLite FTS5)

## Search behavior

- FTS5 is the primary candidate generator
- recency and same-agent affinity adjust ranking
- optional embeddings rerank keyword candidates in-process

## Dedup behavior

- exact duplicate = same `workspace_id + kind + content_norm`
- near duplicate = same topic with high lexical/semantic overlap
- only exact duplicates are auto-applied in v1

