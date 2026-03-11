package core

import (
	"fmt"
	"slices"
	"strings"
)

func ValidateKind(kind MemoryKind) error {
	if kind == "" {
		return fmt.Errorf("%w: kind is required", ErrInvalidArgument)
	}
	if !slices.Contains(AllKinds(), kind) {
		return fmt.Errorf("%w: unsupported kind %q", ErrInvalidArgument, kind)
	}
	return nil
}

func ValidateWorkspaceID(workspaceID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return fmt.Errorf("%w: workspace_id is required", ErrInvalidArgument)
	}
	return nil
}

func ValidateAgentID(agentID string) error {
	if strings.TrimSpace(agentID) == "" {
		return fmt.Errorf("%w: agent_id is required", ErrInvalidArgument)
	}
	return nil
}

func NormalizeLimit(limit, fallback, max int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}
