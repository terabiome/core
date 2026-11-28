package workflow

import (
	"context"
	"errors"
	"testing"
	"uuid"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type recordingHandler struct {
	events     []StepEvent
	finishErr  error
	dispatches []Dispatch
	found      []bool
}

func (h *recordingHandler) Handle(ctx context.Context, event StepEvent) error {
	dispatch, ok := DispatchFromContext(ctx)
	h.dispatches = append(h.dispatches, dispatch)
	h.found = append(h.found, ok)
	h.events = append(h.events, event)
	if event.Stage == StepStageFinish {
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

func TestStepHandlerSeesStartAndFinish(t *testing.T) {
	t.Parallel()

	registry := operationspkg.NewRegistry()
	if err := registry.Register("stub.ok", func() operationspkg.Operation { return stubOperation{} }); err != nil {
		t.Fatal(err)
	}

	handler := &recordingHandler{}
	seq := manifestcontracts.UnitOfWorkSequence{
		Name:  "smoke",
		Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}},
	}
	err := NewSequenceExecutor(nil, registry, handler).Execute(context.Background(), seq)
	if err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}
	if len(handler.events) != 2 {
		t.Fatalf("events = %+v, want start and finish", handler.events)
	}
	if handler.events[0].Stage != StepStageStart || handler.events[1].Stage != StepStageFinish || handler.events[1].Err != nil {
		t.Fatalf("events = %+v", handler.events)
	}
	want := Step{Sequence: "smoke", Name: "stub.ok", Index: 0}
	if handler.events[0].Step != want || handler.events[1].Step != want {
		t.Fatalf("events = %+v", handler.events)
	}
}

func TestStepHandlerSeesDispatchFromContext(t *testing.T) {
	t.Parallel()

	registry := operationspkg.NewRegistry()
	if err := registry.Register("stub.ok", func() operationspkg.Operation { return stubOperation{} }); err != nil {
		t.Fatal(err)
	}

	handler := &recordingHandler{}
	dispatch := Dispatch{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}
	seqs := []manifestcontracts.UnitOfWorkSequence{
		{Name: "smoke", Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}}},
	}

	if err := NewWorkflowExecutor(nil, registry, 1, handler).ExecuteAsDAG(WithDispatch(context.Background(), dispatch), seqs); err != nil {
		t.Fatalf("ExecuteAsDAG() = %v", err)
	}
	if len(handler.events) != 2 {
		t.Fatalf("events = %+v, want start and finish", handler.events)
	}
	for i := range handler.events {
		if !handler.found[i] || handler.dispatches[i] != dispatch {
			t.Fatalf("event %d dispatch = %+v (ok=%t), want %+v", i, handler.dispatches[i], handler.found[i], dispatch)
		}
	}
}

func TestStepHandlerErrorFailsSuccessfulStep(t *testing.T) {
	t.Parallel()

	registry := operationspkg.NewRegistry()
	if err := registry.Register("stub.ok", func() operationspkg.Operation { return stubOperation{} }); err != nil {
		t.Fatal(err)
	}

	handler := &recordingHandler{finishErr: errors.New("event sink down")}
	seq := manifestcontracts.UnitOfWorkSequence{
		Name:  "smoke",
		Steps: []manifestcontracts.UnitOfWorkStep{{Name: "stub.ok"}},
	}
	err := NewSequenceExecutor(nil, registry, handler).Execute(context.Background(), seq)
	if err == nil || !errors.Is(err, handler.finishErr) {
		t.Fatalf("Execute() = %v, want handler error", err)
	}
}
