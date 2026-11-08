package executor

import (
	"context"
	"fmt"
	"sync"

	yamlcontracts "github.com/terabiome/core/internal/contracts/yaml"
	"github.com/terabiome/core/internal/operations"
)

// WorkflowExecutor handles the top-level coordination of a UnitOfWork.
type WorkflowExecutor struct {
	env               *operations.OperationEnvironment
	concurrency       int
	concurrencyFactor int
}

func NewWorkflowExecutor(env *operations.OperationEnvironment) *WorkflowExecutor {
	return &WorkflowExecutor{
		env:               env,
		concurrency:       0,
		concurrencyFactor: 4,
	}
}

// Run in concurrent at best effort
func (w *WorkflowExecutor) ExecuteAsDAG(ctx context.Context, seqs []yamlcontracts.UnitOfWorkSequence) error {
	// atp, the sequence slice is guarantee to form a DAG
	var (
		stagedSequences        = make(map[string]yamlcontracts.UnitOfWorkSequence, len(seqs))
		completedSequenceNames = make(map[string]bool, len(seqs))
	)

	for _, seq := range seqs {
		stagedSequences[seq.Name] = seq
	}

	var (
		errChan = make(chan error, len(seqs))
		wg      = sync.WaitGroup{}
	)
	// sweep
	for len(completedSequenceNames) != len(seqs) && len(stagedSequences) != 0 {
		for _, stagedSequence := range stagedSequences {
			// just to make sure
			if _, ok := completedSequenceNames[stagedSequence.Name]; ok {
				delete(stagedSequences, stagedSequence.Name)
				continue
			}
			// check if sequence's dependencies are COMPLETED, else skip
			for _, dependencySequenceName := range stagedSequence.DependsOn {
				if _, ok := completedSequenceNames[dependencySequenceName]; !ok {
					continue
				}
			}
			// must not dispatch more than <concurrencyFactor>
			if w.concurrency >= w.concurrencyFactor {
				continue
			}

			// enlist for military duty
			wg.Add(1)
			w.concurrency++
			go func() {
				defer wg.Done()
				// Reuse our SequenceExecutor intrinsic
				errChan <- NewSequenceExecutor(w.env).Execute(ctx, stagedSequence)
				w.concurrency--
				completedSequenceNames[stagedSequence.Name] = true
				delete(stagedSequences, stagedSequence.Name)
			}()
		}
	}

	return nil
}

// Execute takes pre-sorted batches and runs them.
// Batches are sequential; sequences within a batch are concurrent.
func (w *WorkflowExecutor) Execute(ctx context.Context, batches []yamlcontracts.UnitOfWorkSequenceBatch) error {
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

func (w *WorkflowExecutor) runBatch(ctx context.Context, batch yamlcontracts.UnitOfWorkSequenceBatch) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(batch.Sequences))

	for _, seq := range batch.Sequences {
		wg.Add(1)
		go func(s yamlcontracts.UnitOfWorkSequence) {
			defer wg.Done()
			// Reuse our SequenceExecutor intrinsic
			exec := NewSequenceExecutor(w.env)
			if err := exec.Execute(ctx, s); err != nil {
				errChan <- err
			}
		}(seq)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
