package core

import (
	"context"
	"time"
)

type MemoryKind string

const (
	KindProfile    MemoryKind = "profile"
	KindFact       MemoryKind = "fact"
	KindPreference MemoryKind = "preference"
	KindConstraint MemoryKind = "constraint"
	KindDecision   MemoryKind = "decision"
	KindProcedure  MemoryKind = "procedure"
	KindWorking    MemoryKind = "working"
)

func AllKinds() []MemoryKind {
	return []MemoryKind{
		KindProfile,
		KindFact,
		KindPreference,
		KindConstraint,
		KindDecision,
		KindProcedure,
		KindWorking,
	}
}

type MemoryState string

const (
	StateActive      MemoryState = "active"
	StateSuperseded  MemoryState = "superseded"
	StateDeleted     MemoryState = "deleted"
	StateQuarantined MemoryState = "quarantined"
)

type Memory struct {
	ID           string      `json:"id"`
	WorkspaceID  string      `json:"workspace_id"`
	AgentID      string      `json:"agent_id"`
	SessionID    string      `json:"session_id,omitempty"`
	Kind         MemoryKind  `json:"kind"`
	Content      string      `json:"content"`
	ContentNorm  string      `json:"content_norm,omitempty"`
	Embedding    []float32   `json:"-"`
	State        MemoryState `json:"state"`
	Confidence   float64     `json:"confidence"`
	Source       string      `json:"source,omitempty"`
	SupersededBy string      `json:"superseded_by,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty"`
}

type MemoryEvent struct {
	ID          string    `json:"id"`
	MemoryID    string    `json:"memory_id,omitempty"`
	WorkspaceID string    `json:"workspace_id"`
	AgentID     string    `json:"agent_id,omitempty"`
	Action      string    `json:"action"`
	PayloadJSON string    `json:"payload_json"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateMemoryRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	AgentID     string     `json:"agent_id"`
	SessionID   string     `json:"session_id,omitempty"`
	Kind        MemoryKind `json:"kind"`
	Content     string     `json:"content"`
	Source      string     `json:"source,omitempty"`
	Confidence  float64    `json:"confidence,omitempty"`
}

type SearchRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	AgentID     string     `json:"agent_id,omitempty"`
	SessionID   string     `json:"session_id,omitempty"`
	Kind        MemoryKind `json:"kind,omitempty"`
	Query       string     `json:"query,omitempty"`
	Limit       int        `json:"limit,omitempty"`
}

type SearchOptions struct {
	WorkspaceID string
	AgentID     string
	Kind        MemoryKind
	Query       string
	Limit       int
}

type SearchHit struct {
	Memory        Memory  `json:"memory"`
	Score         float64 `json:"score"`
	LexicalScore  float64 `json:"lexical_score"`
	SemanticScore float64 `json:"semantic_score"`
	RecencyScore  float64 `json:"recency_score"`
	SameAgent     bool    `json:"same_agent"`
}

type SearchResponse struct {
	Items []SearchHit `json:"items"`
}

type CorrectMemoryRequest struct {
	WorkspaceID string  `json:"workspace_id"`
	AgentID     string  `json:"agent_id,omitempty"`
	SessionID   string  `json:"session_id,omitempty"`
	Content     string  `json:"content"`
	Source      string  `json:"source,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
}

type CorrectMemoryResult struct {
	OldMemory Memory `json:"old_memory"`
	NewMemory Memory `json:"new_memory"`
}

type DeleteMemoryRequest struct {
	WorkspaceID string `json:"workspace_id"`
	AgentID     string `json:"agent_id,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type DeleteMemoryResult struct {
	Memory Memory `json:"memory"`
}

type DedupPreviewRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	AgentID     string     `json:"agent_id,omitempty"`
	Kind        MemoryKind `json:"kind,omitempty"`
	Limit       int        `json:"limit,omitempty"`
}

type DedupExactGroup struct {
	Kind               MemoryKind `json:"kind"`
	ContentNorm        string     `json:"content_norm"`
	SuggestedCanonical string     `json:"suggested_canonical"`
	Items              []Memory   `json:"items"`
}

type DedupNearGroup struct {
	Kind   MemoryKind `json:"kind"`
	Left   Memory     `json:"left"`
	Right  Memory     `json:"right"`
	Score  float64    `json:"score"`
	Method string     `json:"method"`
}

type DedupPreviewResult struct {
	Exact []DedupExactGroup `json:"exact_duplicates"`
	Near  []DedupNearGroup  `json:"near_duplicates"`
}

type DedupApplyRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	AgentID     string   `json:"agent_id,omitempty"`
	CanonicalID string   `json:"canonical_id,omitempty"`
	MemoryIDs   []string `json:"memory_ids"`
}

type DedupApplyResult struct {
	CanonicalID string   `json:"canonical_id"`
	DeletedIDs  []string `json:"deleted_ids"`
}

type ProfileRequest struct {
	WorkspaceID string `json:"workspace_id"`
	AgentID     string `json:"agent_id,omitempty"`
}

type ProfileResult struct {
	WorkspaceID string         `json:"workspace_id"`
	AgentID     string         `json:"agent_id,omitempty"`
	Counts      map[string]int `json:"counts"`
	Recent      []Memory       `json:"recent"`
	Summary     string         `json:"summary"`
}

type HealthRequest struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type HealthResult struct {
	DBPath           string         `json:"db_path"`
	EmbeddingEnabled bool           `json:"embedding_enabled"`
	Total            int            `json:"total"`
	States           map[string]int `json:"states"`
	FTSRows          int            `json:"fts_rows"`
	ExactGroups      int            `json:"exact_duplicate_groups"`
}

type Repository interface {
	Path() string
	CreateMemory(context.Context, Memory) (Memory, error)
	GetMemory(context.Context, string, string) (Memory, error)
	SearchFTS(context.Context, SearchOptions) ([]SearchHit, error)
	ListRecent(context.Context, string, string, MemoryKind, int) ([]SearchHit, error)
	Supersede(context.Context, Memory, Memory) (Memory, error)
	SoftDelete(context.Context, string, string) (Memory, error)
	ListActiveMemories(context.Context, string, MemoryKind, int) ([]Memory, error)
	ListActiveMemoriesByIDs(context.Context, string, []string) ([]Memory, error)
	CountStates(context.Context, string) (map[string]int, int, error)
	CountExactDuplicateGroups(context.Context, string) (int, error)
	LogEvent(context.Context, MemoryEvent) error
}
