package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/moukeyu/memd/internal/embed"
)

type Service struct {
	repo     Repository
	embedder embed.Embedder
}

func NewService(repo Repository, embedder embed.Embedder) *Service {
	if embedder == nil {
		embedder = embed.New(embed.Config{})
	}
	return &Service{repo: repo, embedder: embedder}
}

func (s *Service) DBPath() string {
	return s.repo.Path()
}

func (s *Service) Store(ctx context.Context, req CreateMemoryRequest) (Memory, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return Memory{}, err
	}
	if err := ValidateAgentID(req.AgentID); err != nil {
		return Memory{}, err
	}
	if err := ValidateKind(req.Kind); err != nil {
		return Memory{}, err
	}
	if strings.TrimSpace(req.Content) == "" {
		return Memory{}, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}
	memory := Memory{
		ID:          uuid.NewString(),
		WorkspaceID: req.WorkspaceID,
		AgentID:     req.AgentID,
		SessionID:   req.SessionID,
		Kind:        req.Kind,
		Content:     strings.TrimSpace(req.Content),
		ContentNorm: NormalizeContent(req.Content),
		Confidence:  req.Confidence,
		Source:      req.Source,
	}
	if memory.Source == "" {
		memory.Source = req.AgentID
	}
	if s.embedder.Enabled() {
		embs, err := s.embedder.BatchEmbed(ctx, []string{memory.Content})
		if err != nil {
			return Memory{}, err
		}
		if len(embs) == 1 {
			memory.Embedding = embs[0]
		}
	}
	stored, err := s.repo.CreateMemory(ctx, memory)
	if err != nil {
		return Memory{}, err
	}
	payload, _ := json.Marshal(map[string]any{"kind": stored.Kind, "source": stored.Source})
	_ = s.repo.LogEvent(ctx, MemoryEvent{
		MemoryID:    stored.ID,
		WorkspaceID: stored.WorkspaceID,
		AgentID:     stored.AgentID,
		Action:      "create",
		PayloadJSON: string(payload),
	})
	return stored, nil
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	return s.rankAndSearch(ctx, req, false)
}

func (s *Service) Retrieve(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	return s.rankAndSearch(ctx, req, true)
}

func (s *Service) rankAndSearch(ctx context.Context, req SearchRequest, includeRetrieveEvent bool) (SearchResponse, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return SearchResponse{}, err
	}
	req.Limit = NormalizeLimit(req.Limit, 5, 50)
	if req.Kind != "" {
		if err := ValidateKind(req.Kind); err != nil {
			return SearchResponse{}, err
		}
	}
	hits, err := s.repo.SearchFTS(ctx, SearchOptions{
		WorkspaceID: req.WorkspaceID,
		AgentID:     req.AgentID,
		Kind:        req.Kind,
		Query:       req.Query,
		Limit:       req.Limit * 5,
	})
	if err != nil {
		return SearchResponse{}, err
	}
	if len(hits) == 0 {
		hits, err = s.repo.ListRecent(ctx, req.WorkspaceID, req.AgentID, req.Kind, req.Limit)
		if err != nil {
			return SearchResponse{}, err
		}
	}
	if len(hits) == 0 {
		return SearchResponse{}, nil
	}
	var queryEmbedding []float32
	if s.embedder.Enabled() && strings.TrimSpace(req.Query) != "" {
		embs, err := s.embedder.BatchEmbed(ctx, []string{req.Query})
		if err == nil && len(embs) == 1 {
			queryEmbedding = embs[0]
		}
	}
	now := time.Now().UTC()
	for i := range hits {
		hits[i].SameAgent = req.AgentID != "" && hits[i].Memory.AgentID == req.AgentID
		hits[i].RecencyScore = recencyScore(hits[i].Memory.UpdatedAt, now)
		hits[i].LexicalScore = lexicalScore(hits[i].LexicalScore, strings.TrimSpace(req.Query) == "")
		if len(queryEmbedding) > 0 && len(hits[i].Memory.Embedding) > 0 {
			hits[i].SemanticScore = cosine(queryEmbedding, hits[i].Memory.Embedding)
		}
		hits[i].Score = finalScore(hits[i])
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})
	if len(hits) > req.Limit {
		hits = hits[:req.Limit]
	}
	if includeRetrieveEvent {
		payload, _ := json.Marshal(map[string]any{"query": req.Query, "count": len(hits)})
		_ = s.repo.LogEvent(ctx, MemoryEvent{
			WorkspaceID: req.WorkspaceID,
			AgentID:     req.AgentID,
			Action:      "retrieve",
			PayloadJSON: string(payload),
		})
	}
	return SearchResponse{Items: hits}, nil
}

func (s *Service) Correct(ctx context.Context, memoryID string, req CorrectMemoryRequest) (CorrectMemoryResult, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return CorrectMemoryResult{}, err
	}
	if strings.TrimSpace(req.Content) == "" {
		return CorrectMemoryResult{}, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}
	oldMemory, err := s.repo.GetMemory(ctx, req.WorkspaceID, memoryID)
	if err != nil {
		return CorrectMemoryResult{}, err
	}
	agentID := req.AgentID
	if strings.TrimSpace(agentID) == "" {
		agentID = oldMemory.AgentID
	}
	newMemory := Memory{
		ID:          uuid.NewString(),
		WorkspaceID: oldMemory.WorkspaceID,
		AgentID:     agentID,
		SessionID:   req.SessionID,
		Kind:        oldMemory.Kind,
		Content:     strings.TrimSpace(req.Content),
		ContentNorm: NormalizeContent(req.Content),
		Confidence:  req.Confidence,
		Source:      req.Source,
	}
	if newMemory.Source == "" {
		newMemory.Source = agentID
	}
	if newMemory.Confidence <= 0 {
		newMemory.Confidence = oldMemory.Confidence
	}
	if s.embedder.Enabled() {
		embs, err := s.embedder.BatchEmbed(ctx, []string{newMemory.Content})
		if err != nil {
			return CorrectMemoryResult{}, err
		}
		if len(embs) == 1 {
			newMemory.Embedding = embs[0]
		}
	}
	stored, err := s.repo.Supersede(ctx, oldMemory, newMemory)
	if err != nil {
		return CorrectMemoryResult{}, err
	}
	payload, _ := json.Marshal(map[string]any{"old_memory_id": oldMemory.ID, "new_memory_id": stored.ID})
	_ = s.repo.LogEvent(ctx, MemoryEvent{
		MemoryID:    stored.ID,
		WorkspaceID: stored.WorkspaceID,
		AgentID:     stored.AgentID,
		Action:      "correct",
		PayloadJSON: string(payload),
	})
	oldMemory.State = StateSuperseded
	oldMemory.SupersededBy = stored.ID
	return CorrectMemoryResult{OldMemory: oldMemory, NewMemory: stored}, nil
}

func (s *Service) Delete(ctx context.Context, memoryID string, req DeleteMemoryRequest) (DeleteMemoryResult, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return DeleteMemoryResult{}, err
	}
	memory, err := s.repo.SoftDelete(ctx, req.WorkspaceID, memoryID)
	if err != nil {
		return DeleteMemoryResult{}, err
	}
	payload, _ := json.Marshal(map[string]any{"reason": req.Reason})
	_ = s.repo.LogEvent(ctx, MemoryEvent{
		MemoryID:    memory.ID,
		WorkspaceID: memory.WorkspaceID,
		AgentID:     req.AgentID,
		Action:      "delete",
		PayloadJSON: string(payload),
	})
	return DeleteMemoryResult{Memory: memory}, nil
}

func (s *Service) DedupPreview(ctx context.Context, req DedupPreviewRequest) (DedupPreviewResult, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return DedupPreviewResult{}, err
	}
	if req.Kind != "" {
		if err := ValidateKind(req.Kind); err != nil {
			return DedupPreviewResult{}, err
		}
	}
	limit := NormalizeLimit(req.Limit, 100, 500)
	memories, err := s.repo.ListActiveMemories(ctx, req.WorkspaceID, req.Kind, limit)
	if err != nil {
		return DedupPreviewResult{}, err
	}
	result := DedupPreviewResult{}
	grouped := make(map[string][]Memory)
	for _, memory := range memories {
		key := fmt.Sprintf("%s|%s", memory.Kind, memory.ContentNorm)
		grouped[key] = append(grouped[key], memory)
	}
	for _, group := range grouped {
		if len(group) < 2 {
			continue
		}
		sort.SliceStable(group, func(i, j int) bool {
			li, lj := len(group[i].Content), len(group[j].Content)
			if li == lj {
				return group[i].UpdatedAt.After(group[j].UpdatedAt)
			}
			return li > lj
		})
		result.Exact = append(result.Exact, DedupExactGroup{
			Kind:               group[0].Kind,
			ContentNorm:        group[0].ContentNorm,
			SuggestedCanonical: group[0].ID,
			Items:              group,
		})
	}
	for i := 0; i < len(memories); i++ {
		for j := i + 1; j < len(memories); j++ {
			if memories[i].Kind != memories[j].Kind {
				continue
			}
			if memories[i].ContentNorm == memories[j].ContentNorm {
				continue
			}
			score, method := nearDuplicateScore(memories[i], memories[j])
			if score < 0.72 {
				continue
			}
			result.Near = append(result.Near, DedupNearGroup{
				Kind:   memories[i].Kind,
				Left:   memories[i],
				Right:  memories[j],
				Score:  score,
				Method: method,
			})
		}
	}
	sort.SliceStable(result.Near, func(i, j int) bool { return result.Near[i].Score > result.Near[j].Score })
	payload, _ := json.Marshal(map[string]any{"exact_groups": len(result.Exact), "near_groups": len(result.Near)})
	_ = s.repo.LogEvent(ctx, MemoryEvent{
		WorkspaceID: req.WorkspaceID,
		AgentID:     req.AgentID,
		Action:      "dedup_preview",
		PayloadJSON: string(payload),
	})
	return result, nil
}

func (s *Service) DedupApply(ctx context.Context, req DedupApplyRequest) (DedupApplyResult, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return DedupApplyResult{}, err
	}
	if len(req.MemoryIDs) < 2 {
		return DedupApplyResult{}, ErrNoMemoriesProvided
	}
	memories, err := s.repo.ListActiveMemoriesByIDs(ctx, req.WorkspaceID, req.MemoryIDs)
	if err != nil {
		return DedupApplyResult{}, err
	}
	if len(memories) != len(req.MemoryIDs) {
		return DedupApplyResult{}, ErrNotFound
	}
	baseKind := memories[0].Kind
	baseNorm := memories[0].ContentNorm
	for _, memory := range memories[1:] {
		if memory.Kind != baseKind || memory.ContentNorm != baseNorm {
			return DedupApplyResult{}, ErrNotExactDuplicates
		}
	}
	canonical := chooseCanonical(memories)
	if req.CanonicalID != "" {
		for _, memory := range memories {
			if memory.ID == req.CanonicalID {
				canonical = memory
				break
			}
		}
		if canonical.ID != req.CanonicalID {
			return DedupApplyResult{}, ErrNotFound
		}
	}
	deleted := make([]string, 0, len(memories)-1)
	for _, memory := range memories {
		if memory.ID == canonical.ID {
			continue
		}
		if _, err := s.repo.SoftDelete(ctx, req.WorkspaceID, memory.ID); err != nil {
			return DedupApplyResult{}, err
		}
		deleted = append(deleted, memory.ID)
		payload, _ := json.Marshal(map[string]any{"canonical_id": canonical.ID})
		_ = s.repo.LogEvent(ctx, MemoryEvent{
			MemoryID:    memory.ID,
			WorkspaceID: req.WorkspaceID,
			AgentID:     req.AgentID,
			Action:      "dedup_apply",
			PayloadJSON: string(payload),
		})
	}
	return DedupApplyResult{CanonicalID: canonical.ID, DeletedIDs: deleted}, nil
}

func (s *Service) Profile(ctx context.Context, req ProfileRequest) (ProfileResult, error) {
	if err := ValidateWorkspaceID(req.WorkspaceID); err != nil {
		return ProfileResult{}, err
	}
	memories, err := s.repo.ListActiveMemories(ctx, req.WorkspaceID, "", 50)
	if err != nil {
		return ProfileResult{}, err
	}
	counts := map[string]int{}
	recent := make([]Memory, 0, 10)
	var builder []string
	for _, memory := range memories {
		if req.AgentID != "" && memory.AgentID != req.AgentID {
			continue
		}
		counts[string(memory.Kind)]++
		if len(recent) < 10 {
			recent = append(recent, memory)
		}
		if len(builder) < 5 {
			builder = append(builder, fmt.Sprintf("[%s] %s", memory.Kind, memory.Content))
		}
	}
	return ProfileResult{
		WorkspaceID: req.WorkspaceID,
		AgentID:     req.AgentID,
		Counts:      counts,
		Recent:      recent,
		Summary:     strings.Join(builder, "\n"),
	}, nil
}

func (s *Service) Health(ctx context.Context, req HealthRequest) (HealthResult, error) {
	counts, ftsRows, err := s.repo.CountStates(ctx, req.WorkspaceID)
	if err != nil {
		return HealthResult{}, err
	}
	exactGroups, err := s.repo.CountExactDuplicateGroups(ctx, req.WorkspaceID)
	if err != nil {
		return HealthResult{}, err
	}
	total := 0
	for _, count := range counts {
		total += count
	}
	return HealthResult{
		DBPath:           s.repo.Path(),
		EmbeddingEnabled: s.embedder.Enabled(),
		Total:            total,
		States:           counts,
		FTSRows:          ftsRows,
		ExactGroups:      exactGroups,
	}, nil
}

func lexicalScore(raw float64, recentFallback bool) float64 {
	if recentFallback {
		return 0.25
	}
	if raw <= 0 {
		return 1
	}
	return 1 / (1 + raw)
}

func recencyScore(ts, now time.Time) float64 {
	ageHours := now.Sub(ts).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	return 1 / (1 + ageHours/24)
}

func finalScore(hit SearchHit) float64 {
	score := hit.LexicalScore*0.55 + hit.SemanticScore*0.35 + hit.RecencyScore*0.10
	if hit.SameAgent {
		score += 0.15
	}
	return score
}

func cosine(left, right []float32) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		lv := float64(left[i])
		rv := float64(right[i])
		dot += lv * rv
		leftNorm += lv * lv
		rightNorm += rv * rv
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func nearDuplicateScore(left, right Memory) (float64, string) {
	if len(left.Embedding) > 0 && len(left.Embedding) == len(right.Embedding) {
		return cosine(left.Embedding, right.Embedding), "embedding"
	}
	leftTokens := TokenizeNormalized(left.Content)
	rightTokens := TokenizeNormalized(right.Content)
	if len(leftTokens) == 0 || len(rightTokens) == 0 {
		return 0, "token"
	}
	leftSet := make(map[string]struct{}, len(leftTokens))
	for _, token := range leftTokens {
		leftSet[token] = struct{}{}
	}
	intersections := 0
	union := len(leftSet)
	for _, token := range rightTokens {
		if _, ok := leftSet[token]; ok {
			intersections++
			continue
		}
		union++
	}
	if union == 0 {
		return 0, "token"
	}
	score := float64(intersections) / float64(union)
	if strings.Contains(left.ContentNorm, right.ContentNorm) || strings.Contains(right.ContentNorm, left.ContentNorm) {
		score = maxFloat(score, 0.75)
	}
	return score, "token"
}

func chooseCanonical(memories []Memory) Memory {
	best := memories[0]
	for _, memory := range memories[1:] {
		if len(memory.Content) > len(best.Content) {
			best = memory
			continue
		}
		if len(memory.Content) == len(best.Content) && memory.UpdatedAt.After(best.UpdatedAt) {
			best = memory
		}
	}
	return best
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
