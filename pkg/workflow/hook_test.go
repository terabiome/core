package workflow

import (
	"context"
	"errors"
	"testing"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type recordingHandler struct {
	events    []StepEvent
	finishErr error
}

func (h *recordingHandler) Handle(ctx context.Context, event StepEvent) error {
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
