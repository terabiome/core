package workflow

import (
	"context"
	"errors"
)

type Event struct {
	Scope    Scope
	Sequence string
	Step     string
	Index    int
	Stage    Stage
	Status   Status
	Err      error
}

type Scope string

const (
	ScopeDispatch Scope = "dispatch"
	ScopeSequence Scope = "sequence"
	ScopeStep     Scope = "step"
)

type Stage string

const (
	StageStart  Stage = "start"
	StageFinish Stage = "finish"
)

type Status string

const (
	StatusOK        Status = "ok"
	StatusFailed    Status = "failed"
	StatusBlocked   Status = "blocked"
	StatusCancelled Status = "cancelled"
)

func statusFor(err error) Status {
	switch {
	case err == nil:
		return StatusOK
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return StatusCancelled
	default:
		return StatusFailed
	}
}
