package sqlite

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/moukeyu/memd/internal/config"
	"github.com/moukeyu/memd/internal/core"
)

type Repository struct {
	db   *sql.DB
	path string
}

func Open(ctx context.Context, path string) (*Repository, error) {
	if path == "" {
		path = config.DefaultDBPath()
	}
	if err := config.EnsureParentDir(path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	repo := &Repository{db: db, path: path}
	if err := repo.Init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Path() string { return r.path }

func (r *Repository) Close() error { return r.db.Close() }

func (r *Repository) Init(ctx context.Context) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA busy_timeout=5000;`,
		`PRAGMA foreign_keys=ON;`,
		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			workspace_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			session_id TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL,
			content TEXT NOT NULL,
			content_norm TEXT NOT NULL,
			embedding BLOB NULL,
			state TEXT NOT NULL,
			confidence REAL NOT NULL DEFAULT 0.8,
			source TEXT NOT NULL DEFAULT '',
			superseded_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_memories_ws_state_updated ON memories(workspace_id, state, updated_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_memories_ws_kind_norm_state ON memories(workspace_id, kind, content_norm, state);`,
		`CREATE INDEX IF NOT EXISTS idx_memories_ws_agent_state ON memories(workspace_id, agent_id, state, updated_at DESC);`,
		`CREATE TABLE IF NOT EXISTS memory_events (
			id TEXT PRIMARY KEY,
			memory_id TEXT NULL,
			workspace_id TEXT NOT NULL,
			agent_id TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_events_ws_created ON memory_events(workspace_id, created_at DESC);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
			memory_id UNINDEXED,
			workspace_id UNINDEXED,
			agent_id UNINDEXED,
			kind UNINDEXED,
			state UNINDEXED,
			content,
			content_norm,
			tokenize='unicode61'
		);`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateMemory(ctx context.Context, memory core.Memory) (core.Memory, error) {
	now := time.Now().UTC()
	if memory.ID == "" {
		memory.ID = uuid.NewString()
	}
	memory.CreatedAt = now
	memory.UpdatedAt = now
	memory.State = core.StateActive
	if memory.Confidence <= 0 {
		memory.Confidence = 0.8
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return core.Memory{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO memories (
		id, workspace_id, agent_id, session_id, kind, content, content_norm,
		embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		memory.ID, memory.WorkspaceID, memory.AgentID, memory.SessionID, string(memory.Kind), memory.Content,
		memory.ContentNorm, encodeEmbedding(memory.Embedding), string(memory.State), memory.Confidence,
		memory.Source, memory.SupersededBy, memory.CreatedAt, memory.UpdatedAt,
	); err != nil {
		return core.Memory{}, err
	}
	if err := r.upsertFTS(ctx, tx, memory); err != nil {
		return core.Memory{}, err
	}
	if err := tx.Commit(); err != nil {
		return core.Memory{}, err
	}
	return memory, nil
}

func (r *Repository) GetMemory(ctx context.Context, workspaceID, id string) (core.Memory, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, workspace_id, agent_id, session_id, kind, content, content_norm, embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at FROM memories WHERE workspace_id = ? AND id = ?`, workspaceID, id)
	memory, err := scanMemory(row)
	if err == sql.ErrNoRows {
		return core.Memory{}, core.ErrNotFound
	}
	return memory, err
}

func (r *Repository) SearchFTS(ctx context.Context, opts core.SearchOptions) ([]core.SearchHit, error) {
	query := buildFTSQuery(opts.Query)
	if query == "" {
		return nil, nil
	}
	args := []any{query, opts.WorkspaceID, string(core.StateActive)}
	whereKind := ""
	if opts.Kind != "" {
		whereKind = " AND m.kind = ?"
		args = append(args, string(opts.Kind))
	}
	q := `SELECT m.id, m.workspace_id, m.agent_id, m.session_id, m.kind, m.content, m.content_norm, m.embedding,
		m.state, m.confidence, m.source, m.superseded_by, m.created_at, m.updated_at, m.deleted_at,
		bm25(memories_fts, 1.0, 0.5) AS lexical
		FROM memories_fts
		JOIN memories m ON m.id = memories_fts.memory_id
		WHERE memories_fts MATCH ? AND m.workspace_id = ? AND m.state = ?` + whereKind + `
		ORDER BY lexical
		LIMIT ?`
	args = append(args, opts.Limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []core.SearchHit
	for rows.Next() {
		memory, lexical, err := scanMemoryWithLexical(rows)
		if err != nil {
			return nil, err
		}
		hits = append(hits, core.SearchHit{Memory: memory, LexicalScore: lexical})
	}
	return hits, rows.Err()
}

func (r *Repository) ListRecent(ctx context.Context, workspaceID, agentID string, kind core.MemoryKind, limit int) ([]core.SearchHit, error) {
	args := []any{workspaceID, string(core.StateActive)}
	filters := ""
	if kind != "" {
		filters += " AND kind = ?"
		args = append(args, string(kind))
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `SELECT id, workspace_id, agent_id, session_id, kind, content, content_norm, embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at FROM memories WHERE workspace_id = ? AND state = ?`+filters+` ORDER BY updated_at DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []core.SearchHit
	for rows.Next() {
		memory, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		hits = append(hits, core.SearchHit{Memory: memory})
	}
	return hits, rows.Err()
}

func (r *Repository) Supersede(ctx context.Context, oldMemory core.Memory, newMemory core.Memory) (core.Memory, error) {
	now := time.Now().UTC()
	if newMemory.ID == "" {
		newMemory.ID = uuid.NewString()
	}
	newMemory.CreatedAt = now
	newMemory.UpdatedAt = now
	newMemory.State = core.StateActive
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return core.Memory{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE memories SET state = ?, superseded_by = ?, updated_at = ? WHERE id = ? AND workspace_id = ?`, string(core.StateSuperseded), newMemory.ID, now, oldMemory.ID, oldMemory.WorkspaceID); err != nil {
		return core.Memory{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM memories_fts WHERE memory_id = ?`, oldMemory.ID); err != nil {
		return core.Memory{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO memories (
		id, workspace_id, agent_id, session_id, kind, content, content_norm,
		embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		newMemory.ID, newMemory.WorkspaceID, newMemory.AgentID, newMemory.SessionID, string(newMemory.Kind), newMemory.Content,
		newMemory.ContentNorm, encodeEmbedding(newMemory.Embedding), string(newMemory.State), newMemory.Confidence,
		newMemory.Source, newMemory.SupersededBy, newMemory.CreatedAt, newMemory.UpdatedAt,
	); err != nil {
		return core.Memory{}, err
	}
	if err := r.upsertFTS(ctx, tx, newMemory); err != nil {
		return core.Memory{}, err
	}
	if err := tx.Commit(); err != nil {
		return core.Memory{}, err
	}
	return newMemory, nil
}

func (r *Repository) SoftDelete(ctx context.Context, workspaceID, id string) (core.Memory, error) {
	memory, err := r.GetMemory(ctx, workspaceID, id)
	if err != nil {
		return core.Memory{}, err
	}
	now := time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, `UPDATE memories SET state = ?, updated_at = ?, deleted_at = ? WHERE workspace_id = ? AND id = ?`, string(core.StateDeleted), now, now, workspaceID, id); err != nil {
		return core.Memory{}, err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM memories_fts WHERE memory_id = ?`, id); err != nil {
		return core.Memory{}, err
	}
	memory.State = core.StateDeleted
	memory.UpdatedAt = now
	memory.DeletedAt = &now
	return memory, nil
}

func (r *Repository) ListActiveMemories(ctx context.Context, workspaceID string, kind core.MemoryKind, limit int) ([]core.Memory, error) {
	args := []any{workspaceID, string(core.StateActive)}
	filter := ""
	if kind != "" {
		filter = " AND kind = ?"
		args = append(args, string(kind))
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `SELECT id, workspace_id, agent_id, session_id, kind, content, content_norm, embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at FROM memories WHERE workspace_id = ? AND state = ?`+filter+` ORDER BY updated_at DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []core.Memory
	for rows.Next() {
		memory, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, memory)
	}
	return items, rows.Err()
}

func (r *Repository) ListActiveMemoriesByIDs(ctx context.Context, workspaceID string, ids []string) ([]core.Memory, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := []any{workspaceID, string(core.StateActive)}
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, workspace_id, agent_id, session_id, kind, content, content_norm, embedding, state, confidence, source, superseded_by, created_at, updated_at, deleted_at FROM memories WHERE workspace_id = ? AND state = ? AND id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []core.Memory
	for rows.Next() {
		memory, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, memory)
	}
	return items, rows.Err()
}

func (r *Repository) CountStates(ctx context.Context, workspaceID string) (map[string]int, int, error) {
	counts := map[string]int{}
	var rows *sql.Rows
	var err error
	if strings.TrimSpace(workspaceID) == "" {
		rows, err = r.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM memories GROUP BY state`)
	} else {
		rows, err = r.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM memories WHERE workspace_id = ? GROUP BY state`, workspaceID)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, 0, err
		}
		counts[state] = count
		total += count
	}
	var ftsRows int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM memories_fts`).Scan(&ftsRows); err != nil {
		return nil, 0, err
	}
	return counts, ftsRows, nil
}

func (r *Repository) CountExactDuplicateGroups(ctx context.Context, workspaceID string) (int, error) {
	query := `SELECT COUNT(*) FROM (
		SELECT kind, content_norm FROM memories
		WHERE state = ?`
	args := []any{string(core.StateActive)}
	if workspaceID != "" {
		query += ` AND workspace_id = ?`
		args = append(args, workspaceID)
	}
	query += ` GROUP BY workspace_id, kind, content_norm HAVING COUNT(*) > 1
	)`
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) LogEvent(ctx context.Context, event core.MemoryEvent) error {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO memory_events (id, memory_id, workspace_id, agent_id, action, payload_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, event.ID, nullableString(event.MemoryID), event.WorkspaceID, event.AgentID, event.Action, event.PayloadJSON, event.CreatedAt)
	return err
}

func (r *Repository) FileExists() bool {
	_, err := os.Stat(filepath.Clean(r.path))
	return err == nil
}

func (r *Repository) upsertFTS(ctx context.Context, tx *sql.Tx, memory core.Memory) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM memories_fts WHERE memory_id = ?`, memory.ID); err != nil {
		return err
	}
	if memory.State != core.StateActive {
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO memories_fts (memory_id, workspace_id, agent_id, kind, state, content, content_norm) VALUES (?, ?, ?, ?, ?, ?, ?)`, memory.ID, memory.WorkspaceID, memory.AgentID, string(memory.Kind), string(memory.State), memory.Content, memory.ContentNorm)
	return err
}

func scanMemory(scanner interface{ Scan(dest ...any) error }) (core.Memory, error) {
	var m core.Memory
	var kind, state string
	var embedding []byte
	var deletedAt sql.NullTime
	err := scanner.Scan(&m.ID, &m.WorkspaceID, &m.AgentID, &m.SessionID, &kind, &m.Content, &m.ContentNorm, &embedding, &state, &m.Confidence, &m.Source, &m.SupersededBy, &m.CreatedAt, &m.UpdatedAt, &deletedAt)
	if err != nil {
		return core.Memory{}, err
	}
	m.Kind = core.MemoryKind(kind)
	m.State = core.MemoryState(state)
	m.Embedding = decodeEmbedding(embedding)
	if deletedAt.Valid {
		ts := deletedAt.Time.UTC()
		m.DeletedAt = &ts
	}
	return m, nil
}

func scanMemoryWithLexical(scanner interface{ Scan(dest ...any) error }) (core.Memory, float64, error) {
	var m core.Memory
	var kind, state string
	var lexical float64
	var embedding []byte
	var deletedAt sql.NullTime
	err := scanner.Scan(&m.ID, &m.WorkspaceID, &m.AgentID, &m.SessionID, &kind, &m.Content, &m.ContentNorm, &embedding, &state, &m.Confidence, &m.Source, &m.SupersededBy, &m.CreatedAt, &m.UpdatedAt, &deletedAt, &lexical)
	if err != nil {
		return core.Memory{}, 0, err
	}
	m.Kind = core.MemoryKind(kind)
	m.State = core.MemoryState(state)
	m.Embedding = decodeEmbedding(embedding)
	if deletedAt.Valid {
		ts := deletedAt.Time.UTC()
		m.DeletedAt = &ts
	}
	return m, lexical, nil
}

func buildFTSQuery(query string) string {
	tokens := core.TokenizeNormalized(query)
	if len(tokens) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(tokens))
	for _, token := range tokens {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, strings.ReplaceAll(token, `"`, `""`)))
	}
	return strings.Join(quoted, " OR ")
}

func encodeEmbedding(values []float32) []byte {
	if len(values) == 0 {
		return nil
	}
	raw := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
	}
	return raw
}

func decodeEmbedding(raw []byte) []float32 {
	if len(raw) == 0 {
		return nil
	}
	if len(raw)%4 != 0 {
		return nil
	}
	values := make([]float32, len(raw)/4)
	for i := range values {
		values[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[i*4:]))
	}
	return values
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
