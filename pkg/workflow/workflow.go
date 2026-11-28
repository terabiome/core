package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type WorkflowExecutor struct {
	env            *operationspkg.OperationEnvironment
	registry       Registry
	eventHandler   EventHandler
	maxConcurrency int
}

func NewWorkflowExecutor(
	env *operationspkg.OperationEnvironment,
	registry Registry,
	maxConcurrency int,
	handler EventHandler,
) *WorkflowExecutor {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	return &WorkflowExecutor{
		env:            env,
		registry:       registry,
		eventHandler:   handler,
		maxConcurrency: maxConcurrency,
	}
}

func (w *WorkflowExecutor) WithDispatch(ctx context.Context, dispatch Dispatch) context.Context {
	return WithDispatch(ctx, dispatch)
}

func (w *WorkflowExecutor) ExecuteAsDAG(
	ctx context.Context,
	seqs []manifestcontracts.UnitOfWorkSequence,
) error {
	if len(seqs) == 0 {
		return nil
	}

	if err := w.emit(ctx, Event{Scope: ScopeDispatch, Stage: StageStart}); err != nil {
		return err
	}

	return w.finish(
		ctx,
		Event{Scope: ScopeDispatch},
		w.executeAsDAG(ctx, seqs),
	)
}

func (w *WorkflowExecutor) executeAsDAG(
	ctx context.Context,
	seqs []manifestcontracts.UnitOfWorkSequence,
) error {
	if err := ValidateSequenceGraph(seqs); err != nil {
		return err
	}

	limit := w.maxConcurrency
	if limit <= 0 {
		limit = 1
	}

	pending := make(map[string]manifestcontracts.UnitOfWorkSequence, len(seqs))
	order := make([]string, 0, len(seqs))
	for _, seq := range seqs {
		pending[seq.Name] = seq
		order = append(order, seq.Name)
	}
	slices.Sort(order)

	completed := make(map[string]struct{}, len(seqs))

	type outcome struct {
		name string
		err  error
	}
	outcomes := make(chan outcome, len(seqs))

	inFlight := 0
	var runErrs []error

	launchReady := func() {
		for _, name := range order {
			if inFlight >= limit {
				return
			}

			seq, stillPending := pending[name]
			if !stillPending || !dependenciesSatisfied(seq, completed) {
				continue
			}

			delete(pending, name)
			inFlight++

			go func(s manifestcontracts.UnitOfWorkSequence) {
				outcomes <- outcome{
					name: s.Name,
					err:  w.runSequence(ctx, s),
				}
			}(seq)
		}
	}

	drain := func() {
		for ; inFlight > 0; inFlight-- {
			out := <-outcomes
			if out.err != nil {
				runErrs = append(runErrs, fmt.Errorf("sequence %s failed: %w", out.name, out.err))
				continue
			}
			completed[out.name] = struct{}{}
		}
	}

	for len(pending) > 0 || inFlight > 0 {
		launchReady()

		if inFlight == 0 {
			break
		}

		select {
		case <-ctx.Done():
			drain()
			return errors.Join(append(runErrs, ctx.Err())...)
		case out := <-outcomes:
			inFlight--
			if out.err != nil {
				runErrs = append(runErrs, fmt.Errorf("sequence %s failed: %w", out.name, out.err))
				continue
			}
			completed[out.name] = struct{}{}
		}
	}

	if len(pending) > 0 {
		blocked := make([]string, 0, len(pending))
		for name := range pending {
			blocked = append(blocked, name)
		}
		slices.Sort(blocked)

		status := StatusBlocked
		if ctx.Err() != nil {
			status = StatusCancelled
		}
		for _, name := range blocked {
			_ = w.emit(ctx, Event{
				Scope:    ScopeSequence,
				Sequence: name,
				Stage:    StageFinish,
				Status:   status,
			})
		}

		if ctx.Err() == nil {
			runErrs = append(runErrs, fmt.Errorf(
				"sequences not executed due to unresolved dependencies: %s",
				strings.Join(blocked, ", "),
			))
		}
	}

	return errors.Join(runErrs...)
}

func (w *WorkflowExecutor) runSequence(
	ctx context.Context,
	seq manifestcontracts.UnitOfWorkSequence,
) error {
	if err := w.emit(ctx, Event{Scope: ScopeSequence, Sequence: seq.Name, Stage: StageStart}); err != nil {
		return err
	}

	err := NewSequenceExecutor(w.env, w.registry, w.eventHandler).Execute(ctx, seq)
	return w.finish(
		ctx,
		Event{Scope: ScopeSequence, Sequence: seq.Name},
		err,
	)
}

func (w *WorkflowExecutor) finish(ctx context.Context, event Event, runErr error) error {
	event.Stage = StageFinish
	event.Status = statusFor(runErr)
	event.Err = runErr
	if handleErr := w.emit(ctx, event); runErr == nil {
		runErr = handleErr
	}
	return runErr
}

func (w *WorkflowExecutor) emit(ctx context.Context, event Event) error {
	if w.eventHandler == nil {
		return nil
	}
	return w.eventHandler.Handle(ctx, event)
}

func ValidateSequenceGraph(seqs []manifestcontracts.UnitOfWorkSequence) error {
	names := make(map[string]struct{}, len(seqs))
	for _, seq := range seqs {
		if _, exists := names[seq.Name]; exists {
			return fmt.Errorf("duplicate sequence name: %s", seq.Name)
		}
		names[seq.Name] = struct{}{}
	}

	for _, seq := range seqs {
		for _, dependency := range seq.DependsOn {
			if _, exists := names[dependency]; !exists {
				return fmt.Errorf("sequence %s depends on unknown sequence %s", seq.Name, dependency)
			}
		}
	}

	if (manifestcontracts.UnitOfWork{Sequences: seqs}).CheckIfContainCycle() {
		return errors.New("sequence graph contains a cycle")
	}

	return nil
}

func dependenciesSatisfied(
	seq manifestcontracts.UnitOfWorkSequence,
	completed map[string]struct{},
) bool {
	for _, dependency := range seq.DependsOn {
		if _, ok := completed[dependency]; !ok {
			return false
		}
	}
	return true
}

func (w *WorkflowExecutor) Execute(ctx context.Context, batches []manifestcontracts.UnitOfWorkSequenceBatch) error {
	for _, batch := range batches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.runBatch(ctx, batch); err != nil {
			return fmt.Errorf("batch %d failed: %w", batch.ID, err)
		}
	}
	return nil
}

func (w *WorkflowExecutor) runBatch(ctx context.Context, batch manifestcontracts.UnitOfWorkSequenceBatch) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(batch.Sequences))

	for _, seq := range batch.Sequences {
		wg.Add(1)
		go func(s manifestcontracts.UnitOfWorkSequence) {
			defer wg.Done()

			// Check context before starting
			if ctx.Err() != nil {
				errChan <- ctx.Err()
				return
			}

			exec := NewSequenceExecutor(w.env, w.registry, w.eventHandler)
			if err := exec.Execute(ctx, s); err != nil {
				errChan <- fmt.Errorf("sequence %s failed: %w", s.Name, err)
			}
		}(seq)
	}

	wg.Wait()
	close(errChan)

	// Return the first error encountered, if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
