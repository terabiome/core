package workflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"uuid"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type recordingHandler struct {
	mu         sync.Mutex
	events     []Event
	finishErr  error
	dispatches []Dispatch
	found      []bool
}

func (h *recordingHandler) Handle(ctx context.Context, event Event) error {
	dispatch, ok := DispatchFromContext(ctx)

	h.mu.Lock()
	defer h.mu.Unlock()
	h.dispatches = append(h.dispatches, dispatch)
	h.found = append(h.found, ok)
	h.events = append(h.events, event)
	if event.Stage == StageFinish {
		return h.finishErr
	}
	return nil
}

type stubOperation struct{}

func (stubOperation) Name() string                                       { return "stub.ok" }
func (stubOperation) SetParameters(map[string]any) error                 { return nil }
func (stubOperation) PreCheck(*operationspkg.OperationEnvironment) error { return nil }
func (stubOperation) Run(context.Context, *operationspkg.OperationEnvironment) error {
	return nil
}

type failingOperation struct{}

func (failingOperation) Name() string                                       { return "stub.fail" }
func (failingOperation) SetParameters(map[string]any) error                 { return nil }
func (failingOperation) PreCheck(*operationspkg.OperationEnvironment) error { return nil }
func (failingOperation) Run(context.Context, *operationspkg.OperationEnvironment) error {
	return errors.New("operation failed")
}

func newRegistry(t *testing.T) *operationspkg.Registry {
	t.Helper()

	registry := operationspkg.NewRegistry()
	if err := registry.Register("stub.ok", func() operationspkg.Operation { return stubOperation{} }); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register("stub.fail", func() operationspkg.Operation { return failingOperation{} }); err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestEventHandlerSeesStepStartAndFinish(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	seq := manifestcontracts.UnitOfWorkSequence{
		Name:  "smoke",
		Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}},
	}
	err := NewSequenceExecutor(nil, newRegistry(t), handler).Execute(context.Background(), seq)
	if err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}
	if len(handler.events) != 2 {
		t.Fatalf("events = %+v, want start and finish", handler.events)
	}
	if handler.events[0].Stage != StageStart || handler.events[1].Stage != StageFinish || handler.events[1].Err != nil {
		t.Fatalf("events = %+v", handler.events)
	}
	want := Event{Scope: ScopeStep, Sequence: "smoke", Step: "stub.ok", Index: 0}
	for _, event := range handler.events {
		event.Stage = ""
		event.Status = ""
		if event != want {
			t.Fatalf("event = %+v, want %+v", event, want)
		}
	}
	if handler.events[1].Status != StatusOK {
		t.Fatalf("finish status = %q, want %q", handler.events[1].Status, StatusOK)
	}
}

func TestEventHandlerSeesDispatchFromContext(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	dispatch := Dispatch{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}
	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "smoke", Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}}},
	}

	if err := NewWorkflowExecutor(nil, newRegistry(t), 1, handler).ExecuteAsDAG(WithDispatch(context.Background(), dispatch), seqs); err != nil {
		t.Fatalf("ExecuteAsDAG() = %v", err)
	}
	if len(handler.events) != 6 {
		t.Fatalf("events = %+v, want 6 lifecycle events", handler.events)
	}
	for i := range handler.events {
		if !handler.found[i] || handler.dispatches[i] != dispatch {
			t.Fatalf("event %d dispatch = %+v (ok=%t), want %+v", i, handler.dispatches[i], handler.found[i], dispatch)
		}
	}
}

func TestEventHandlerErrorFailsSuccessfulStep(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{finishErr: errors.New("event sink down")}
	seq := manifestcontracts.UnitOfWorkSequence{
		Name:  "smoke",
		Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}},
	}
	err := NewSequenceExecutor(nil, newRegistry(t), handler).Execute(context.Background(), seq)
	if err == nil || !errors.Is(err, handler.finishErr) {
		t.Fatalf("Execute() = %v, want handler error", err)
	}
}

func TestExecuteAsDAGEmitsLifecycleEvents(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "smoke", Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}}},
	}

	if err := NewWorkflowExecutor(nil, newRegistry(t), 1, handler).ExecuteAsDAG(context.Background(), seqs); err != nil {
		t.Fatalf("ExecuteAsDAG() = %v", err)
	}

	want := []Event{
		{Scope: ScopeDispatch, Stage: StageStart},
		{Scope: ScopeSequence, Sequence: "smoke", Stage: StageStart},
		{Scope: ScopeStep, Sequence: "smoke", Step: "stub.ok", Index: 0, Stage: StageStart},
		{Scope: ScopeStep, Sequence: "smoke", Step: "stub.ok", Index: 0, Stage: StageFinish, Status: StatusOK},
		{Scope: ScopeSequence, Sequence: "smoke", Stage: StageFinish, Status: StatusOK},
		{Scope: ScopeDispatch, Stage: StageFinish, Status: StatusOK},
	}
	if len(handler.events) != len(want) {
		t.Fatalf("events = %+v, want %+v", handler.events, want)
	}
	for i := range want {
		if handler.events[i] != want[i] {
			t.Fatalf("event %d = %+v, want %+v", i, handler.events[i], want[i])
		}
	}
}

func TestExecuteAsDAGMarksBlockedSequence(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "bad", Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.fail"}}},
		{Name: "after", DependsOn: []string{"bad"}},
	}

	err := NewWorkflowExecutor(nil, newRegistry(t), 1, handler).ExecuteAsDAG(context.Background(), seqs)
	if err == nil {
		t.Fatal("ExecuteAsDAG() = nil, want an error")
	}

	var blocked *Event
	for i := range handler.events {
		event := &handler.events[i]
		if event.Scope == ScopeSequence && event.Sequence == "after" {
			if event.Stage != StageFinish || event.Status != StatusBlocked {
				t.Fatalf("blocked event = %+v, want finish/blocked", event)
			}
			if blocked != nil {
				t.Fatalf("sequence after reported more than once")
			}
			blocked = event
		}
	}
	if blocked == nil {
		t.Fatalf("events = %+v, want a blocked event for sequence after", handler.events)
	}
}

func TestExecuteAsDAGMarksCancelled(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "smoke", Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewWorkflowExecutor(nil, newRegistry(t), 1, handler).ExecuteAsDAG(ctx, seqs)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteAsDAG() = %v, want context.Canceled", err)
	}

	want := []Event{
		{Scope: ScopeDispatch, Stage: StageStart},
		{Scope: ScopeSequence, Sequence: "smoke", Stage: StageStart},
		{Scope: ScopeSequence, Sequence: "smoke", Stage: StageFinish, Status: StatusCancelled},
		{Scope: ScopeDispatch, Stage: StageFinish, Status: StatusCancelled},
	}
	if len(handler.events) != len(want) {
		t.Fatalf("events = %+v, want %+v", handler.events, want)
	}
	for i := range want {
		got := handler.events[i]
		got.Err = nil
		if got != want[i] {
			t.Fatalf("event %d = %+v, want %+v", i, handler.events[i], want[i])
		}
	}
	if !errors.Is(handler.events[2].Err, context.Canceled) {
		t.Fatalf("sequence finish error = %v, want context.Canceled", handler.events[2].Err)
	}
}
