package workflow

import (
	"context"

	operationspkg "github.com/terabiome/core/pkg/operations"
)

type Registry interface {
	Get(name string) (operationspkg.Operation, error)
}

// StepHandler is called from sequence goroutines and must be safe for concurrent use.
type StepHandler interface {
	Handle(ctx context.Context, event StepEvent) error
}
