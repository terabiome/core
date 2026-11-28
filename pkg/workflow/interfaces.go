package workflow

import (
	"context"

	operationspkg "github.com/terabiome/core/pkg/operations"
)

type Registry interface {
	Get(name string) (operationspkg.Operation, error)
}

// EventHandler is called from sequence goroutines and must be safe for concurrent use.
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}
