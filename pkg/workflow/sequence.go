package workflow

import (
	"context"
	"fmt"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type SequenceExecutor struct {
	env         *operationspkg.OperationEnvironment
	registry    Registry
	stepHandler StepHandler
}

func NewSequenceExecutor(
	env *operationspkg.OperationEnvironment,
	registry Registry,
	stepHandler StepHandler,
) *SequenceExecutor {
	return &SequenceExecutor{
		env:         env,
		registry:    registry,
		stepHandler: stepHandler,
	}
}

func (e *SequenceExecutor) Execute(ctx context.Context, sequence manifestcontracts.UnitOfWorkSequence) error {
	for i, step := range sequence.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.runStep(ctx, sequence.Name, i, step); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", i, step.Name, err)
		}
	}
	return nil
}

func (e *SequenceExecutor) runStep(
	ctx context.Context,
	sequence string,
	index int,
	step manifestcontracts.UnitOfWorkStep,
) error {
	current := Step{Sequence: sequence, Name: step.Name, Index: index}
	if err := e.handle(ctx, StepEvent{Step: current, Stage: StepStageStart}); err != nil {
		return err
	}

	err := e.executeStep(ctx, step)
	if handleErr := e.handle(ctx, StepEvent{Step: current, Stage: StepStageFinish, Err: err}); err == nil {
		err = handleErr
	}
	return err
}

func (e *SequenceExecutor) executeStep(ctx context.Context, step manifestcontracts.UnitOfWorkStep) error {
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

func (e *SequenceExecutor) handle(ctx context.Context, event StepEvent) error {
	if e.stepHandler == nil {
		return nil
	}
	return e.stepHandler.Handle(ctx, event)
}
