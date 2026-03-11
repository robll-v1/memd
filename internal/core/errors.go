package core

import "errors"

var (
	ErrNotFound            = errors.New("memory not found")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrNoMemoriesProvided  = errors.New("no memories provided")
	ErrNotExactDuplicates  = errors.New("memories are not exact duplicates")
	ErrCrossWorkspaceApply = errors.New("memories must belong to the same workspace")
)
