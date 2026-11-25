package workflow

import (
	"context"
	"fmt"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type SequenceExecutor struct {
	env      *operationspkg.OperationEnvironment
	registry Registry
}

func NewSequenceExecutor(
	env *operationspkg.OperationEnvironment,
	registry Registry,
) *SequenceExecutor {
	return &SequenceExecutor{
		env:      env,
		registry: registry,
	}
}

func (e *SequenceExecutor) Execute(ctx context.Context, sequence manifestcontracts.UnitOfWorkSequence) error {
	for i, step := range sequence.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.runStep(ctx, step); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", i, step.Name, err)
		}
	}
	return nil
}

func (e *SequenceExecutor) runStep(ctx context.Context, step manifestcontracts.UnitOfWorkStep) error {
	op, err := e.registry.Get(step.Name)
	if err != nil {
		return err
	}

	if err := op.SetParameters(step.Parameters); err != nil {
		return fmt.Errorf("params invalid for %s: %w", step.Name, err)
	}

	if err := op.PreCheck(e.env); err != nil {
		return fmt.Errorf("pre-check failed for %s: %w", step.Name, err)
	}

	return op.Run(ctx, e.env)
}
