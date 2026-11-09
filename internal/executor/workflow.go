package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	yamlcontracts "github.com/terabiome/core/internal/contracts/yaml"
	"github.com/terabiome/core/internal/operations"
)

// WorkflowExecutor handles the top-level coordination of a UnitOfWork.
type WorkflowExecutor struct {
	env            *operations.OperationEnvironment
	maxConcurrency int
}

func NewWorkflowExecutor(env *operations.OperationEnvironment, maxConcurrency int) *WorkflowExecutor {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	return &WorkflowExecutor{
		env:            env,
		maxConcurrency: maxConcurrency,
	}
}

// ExecuteAsDAG dynamically resolves dependencies and executes sequences concurrently.
func (w *WorkflowExecutor) ExecuteAsDAG(ctx context.Context, seqs []yamlcontracts.UnitOfWorkSequence) error {
	if len(seqs) == 0 {
		return nil
	}

	var (
		mu                     sync.Mutex
		stagedSequences        = make(map[string]yamlcontracts.UnitOfWorkSequence, len(seqs))
		completedSequenceNames = make(map[string]bool, len(seqs))
		activeGoroutines       = 0
		errChan                = make(chan error, len(seqs))
		wg                     sync.WaitGroup
	)

	for _, seq := range seqs {
		stagedSequences[seq.Name] = seq
	}

	// Worker pool to process ready sequences
	for {
		mu.Lock()

		// Check for completion or deadlock
		if len(completedSequenceNames) == len(seqs) {
			mu.Unlock()
			break
		}

		// Find sequences that are ready to run
		var readyToRun []yamlcontracts.UnitOfWorkSequence
		for name, seq := range stagedSequences {
			if completedSequenceNames[name] {
				continue
			}

			// Check dependencies
			allDepsMet := true
			for _, dep := range seq.DependsOn {
				if !completedSequenceNames[dep] {
					allDepsMet = false
					break
				}
			}

			if allDepsMet && activeGoroutines < w.maxConcurrency {
				readyToRun = append(readyToRun, seq)
				activeGoroutines++
				delete(stagedSequences, name)
			}
		}
		mu.Unlock()

		// If nothing is ready and we have active goroutines, wait for one to finish
		if len(readyToRun) == 0 {
			if activeGoroutines <= 0 {
				return fmt.Errorf("deadlock detected: unresolved dependencies in DAG")
			}
			// Briefly sleep to prevent 100% CPU busy-waiting
			mu.Unlock()
			// Note: A proper implementation would use a sync.Cond here to wait for a goroutine to signal completion.
			// For simplicity, yield.
			continue
		}

		// Launch ready sequences
		for _, seq := range readyToRun {
			wg.Add(1)
			go func(s yamlcontracts.UnitOfWorkSequence) {
				defer wg.Done()

				// Execute the sequence
				err := NewSequenceExecutor(w.env).Execute(ctx, s)

				// Record result safely
				mu.Lock()
				if err != nil {
					errChan <- fmt.Errorf("sequence %s failed: %w", s.Name, err)
				}
				completedSequenceNames[s.Name] = true
				activeGoroutines--
				mu.Unlock()
			}(seq)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	var err error
	for subErr := range errChan {
		if subErr != nil {
			err = errors.Join(err, subErr)
		}
	}

	return err
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

			// Check context before starting
			if ctx.Err() != nil {
				errChan <- ctx.Err()
				return
			}

			exec := NewSequenceExecutor(w.env)
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
