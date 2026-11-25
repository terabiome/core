package workflow

import (
	"context"
	"strings"
	"testing"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

const failingStep = "does.not.exist"

func TestExecuteAsDAGEmpty(t *testing.T) {
	t.Parallel()

	exec := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2)
	if err := exec.ExecuteAsDAG(context.Background(), nil); err != nil {
		t.Fatalf("ExecuteAsDAG(nil) = %v, want nil", err)
	}
}

func TestExecuteAsDAGChain(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "a"},
		{Name: "b", DependsOn: []string{"a"}},
		{Name: "c", DependsOn: []string{"b"}},
	}

	if err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs); err != nil {
		t.Fatalf("ExecuteAsDAG(chain) = %v, want nil", err)
	}
}

func TestExecuteAsDAGIndependent(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "a"},
		{Name: "b"},
		{Name: "c", DependsOn: []string{"a", "b"}},
	}

	if err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs); err != nil {
		t.Fatalf("ExecuteAsDAG(independent) = %v, want nil", err)
	}
}

func TestExecuteAsDAGFailureBlocksDependentsOnly(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "bad", Steps: []manifestcontracts.UnitOfWorkStep{{Name: failingStep}}},
		{Name: "after", DependsOn: []string{"bad"}},
		{Name: "good"},
	}

	err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs)
	if err == nil {
		t.Fatal("ExecuteAsDAG() = nil, want an error for the failed sequence")
	}
	if !strings.Contains(err.Error(), "sequence bad failed") {
		t.Errorf("error does not mention the failed sequence: %v", err)
	}
	if !strings.Contains(err.Error(), "after") {
		t.Errorf("error does not mention the blocked dependent: %v", err)
	}
	if strings.Contains(err.Error(), "good") {
		t.Errorf("independent sequence should have run, but error mentions it: %v", err)
	}
}

func TestExecuteAsDAGRejectsUnknownDependency(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "a", DependsOn: []string{"missing"}},
	}

	err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs)
	if err == nil || !strings.Contains(err.Error(), "unknown sequence") {
		t.Fatalf("ExecuteAsDAG() = %v, want unknown sequence error", err)
	}
}

func TestExecuteAsDAGRejectsCycle(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "a", DependsOn: []string{"b"}},
		{Name: "b", DependsOn: []string{"a"}},
	}

	err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("ExecuteAsDAG() = %v, want cycle error", err)
	}
}

func TestExecuteAsDAGRejectsDuplicateName(t *testing.T) {
	t.Parallel()

	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "a"},
		{Name: "a"},
	}

	err := NewWorkflowExecutor(nil, operationspkg.NewRegistry(), 2).ExecuteAsDAG(context.Background(), seqs)
	if err == nil || !strings.Contains(err.Error(), "duplicate sequence name") {
		t.Fatalf("ExecuteAsDAG() = %v, want duplicate name error", err)
	}
}
