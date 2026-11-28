package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"uuid"

	eventcontracts "github.com/terabiome/core/pkg/contracts/events"
	"github.com/terabiome/core/pkg/messaging"
	"github.com/terabiome/core/pkg/workflow"
)

const testSubject = "dispatch.events"

var (
	testDispatch = workflow.Dispatch{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}
	testTime = time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
)

type recordingPublisher struct {
	calls        int
	subject      string
	payload      []byte
	err          error
	ctxErrAtCall error
}

func (p *recordingPublisher) Publish(
	ctx context.Context,
	destination messaging.Destination,
	data []byte,
) error {
	p.calls++
	p.subject = destination.Subject
	p.payload = append([]byte(nil), data...)
	p.ctxErrAtCall = ctx.Err()
	return p.err
}

func newTestHandler(t *testing.T, publisher messaging.Publisher) *EventHandler {
	t.Helper()

	handler, err := NewEventHandler(publisher, testSubject)
	if err != nil {
		t.Fatalf("NewHandler() = %v", err)
	}
	handler.clock = func() time.Time { return testTime }
	return handler
}

func decodeEvent(t *testing.T, publisher *recordingPublisher) eventcontracts.DispatchEvent {
	t.Helper()

	var event eventcontracts.DispatchEvent
	if err := json.Unmarshal(publisher.payload, &event); err != nil {
		t.Fatalf("json.Unmarshal(payload) = %v", err)
	}
	return event
}

func decodeObject(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()

	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("json.Unmarshal(%s) = %v", raw, err)
	}
	return object
}

func TestHandlerEmitsLifecycleEvents(t *testing.T) {
	t.Parallel()

	stepErr := errors.New("boom")
	tests := map[string]struct {
		event    workflow.Event
		wantType eventcontracts.Type
		wantMeta map[string]any
		wantData map[string]any
	}{
		"dispatch start": {
			event:    workflow.Event{Scope: workflow.ScopeDispatch, Stage: workflow.StageStart},
			wantType: eventcontracts.TypeDispatchStart,
			wantMeta: map[string]any{},
			wantData: map[string]any{},
		},
		"dispatch finish ok": {
			event:    workflow.Event{Scope: workflow.ScopeDispatch, Stage: workflow.StageFinish, Status: workflow.StatusOK},
			wantType: eventcontracts.TypeDispatchFinish,
			wantMeta: map[string]any{},
			wantData: map[string]any{"status": "ok"},
		},
		"dispatch finish failed": {
			event:    workflow.Event{Scope: workflow.ScopeDispatch, Stage: workflow.StageFinish, Status: workflow.StatusFailed, Err: stepErr},
			wantType: eventcontracts.TypeDispatchFinish,
			wantMeta: map[string]any{},
			wantData: map[string]any{"status": "failed", "error": "boom"},
		},
		"sequence start": {
			event:    workflow.Event{Scope: workflow.ScopeSequence, Sequence: "smoke", Stage: workflow.StageStart},
			wantType: eventcontracts.TypeSequenceStart,
			wantMeta: map[string]any{"sequence": "smoke"},
			wantData: map[string]any{},
		},
		"sequence finish blocked": {
			event:    workflow.Event{Scope: workflow.ScopeSequence, Sequence: "after", Stage: workflow.StageFinish, Status: workflow.StatusBlocked},
			wantType: eventcontracts.TypeSequenceFinish,
			wantMeta: map[string]any{"sequence": "after"},
			wantData: map[string]any{"status": "blocked"},
		},
		"step start": {
			event:    workflow.Event{Scope: workflow.ScopeStep, Sequence: "smoke", Step: "shell.cmd", Index: 1, Stage: workflow.StageStart},
			wantType: eventcontracts.TypeStepStart,
			wantMeta: map[string]any{"sequence": "smoke", "step": "shell.cmd", "step_index": float64(1)},
			wantData: map[string]any{},
		},
		"step finish ok": {
			event:    workflow.Event{Scope: workflow.ScopeStep, Sequence: "smoke", Step: "shell.cmd", Index: 1, Stage: workflow.StageFinish, Status: workflow.StatusOK},
			wantType: eventcontracts.TypeStepFinish,
			wantMeta: map[string]any{"sequence": "smoke", "step": "shell.cmd", "step_index": float64(1)},
			wantData: map[string]any{"status": "ok"},
		},
		"step finish failed": {
			event:    workflow.Event{Scope: workflow.ScopeStep, Sequence: "smoke", Step: "shell.cmd", Index: 1, Stage: workflow.StageFinish, Status: workflow.StatusFailed, Err: stepErr},
			wantType: eventcontracts.TypeStepFinish,
			wantMeta: map[string]any{"sequence": "smoke", "step": "shell.cmd", "step_index": float64(1)},
			wantData: map[string]any{"status": "failed", "error": "boom"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			publisher := &recordingPublisher{}
			handler := newTestHandler(t, publisher)
			ctx := workflow.WithDispatch(context.Background(), testDispatch)

			if err := handler.Handle(ctx, test.event); err != nil {
				t.Fatalf("Handle() = %v", err)
			}
			if publisher.calls != 1 || publisher.subject != testSubject {
				t.Fatalf("publish calls = %d subject = %q", publisher.calls, publisher.subject)
			}

			event := decodeEvent(t, publisher)
			if event.DispatchUUID != testDispatch.UUID || event.UOWUUID != testDispatch.UOWUUID {
				t.Fatalf("identity = %s/%s", event.DispatchUUID, event.UOWUUID)
			}
			if event.Type != test.wantType {
				t.Fatalf("type = %q, want %q", event.Type, test.wantType)
			}
			if !event.GeneratedAt.Equal(testTime) {
				t.Fatalf("generated_at = %s, want %s", event.GeneratedAt, testTime)
			}
			if got := decodeObject(t, event.Metadata); !equalObjects(got, test.wantMeta) {
				t.Fatalf("metadata = %+v, want %+v", got, test.wantMeta)
			}
			if got := decodeObject(t, event.Data); !equalObjects(got, test.wantData) {
				t.Fatalf("data = %+v, want %+v", got, test.wantData)
			}
		})
	}
}

func TestHandlerPublishesAfterCancellation(t *testing.T) {
	t.Parallel()

	publisher := &recordingPublisher{}
	handler := newTestHandler(t, publisher)

	ctx, cancel := context.WithCancel(workflow.WithDispatch(context.Background(), testDispatch))
	cancel()

	if err := handler.Handle(ctx, workflow.Event{Scope: workflow.ScopeSequence, Sequence: "smoke", Stage: workflow.StageFinish, Status: workflow.StatusCancelled}); err != nil {
		t.Fatalf("Handle() = %v", err)
	}
	if publisher.calls != 1 || publisher.ctxErrAtCall != nil {
		t.Fatalf("publish calls = %d ctx err = %v, want a live publish context", publisher.calls, publisher.ctxErrAtCall)
	}
}

func TestHandlerRejectsMissingDispatch(t *testing.T) {
	t.Parallel()

	publisher := &recordingPublisher{}
	handler := newTestHandler(t, publisher)

	err := handler.Handle(context.Background(), workflow.Event{Scope: workflow.ScopeStep, Stage: workflow.StageStart})
	if err == nil {
		t.Fatal("Handle() = nil, want error")
	}
	if publisher.calls != 0 {
		t.Fatalf("publish calls = %d, want 0", publisher.calls)
	}
}

func TestHandlerRejectsUnsupportedEvent(t *testing.T) {
	t.Parallel()

	tests := map[string]workflow.Event{
		"unknown scope": {Scope: workflow.Scope("middle"), Stage: workflow.StageStart},
		"unknown stage": {Scope: workflow.ScopeStep, Stage: workflow.Stage("middle")},
		"finish without status": {
			Scope: workflow.ScopeStep, Stage: workflow.StageFinish,
		},
	}

	for name, event := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			publisher := &recordingPublisher{}
			handler := newTestHandler(t, publisher)
			ctx := workflow.WithDispatch(context.Background(), testDispatch)

			if err := handler.Handle(ctx, event); err == nil {
				t.Fatal("Handle() = nil, want error")
			}
			if publisher.calls != 0 {
				t.Fatalf("publish calls = %d, want 0", publisher.calls)
			}
		})
	}
}

func TestHandlerReturnsPublishError(t *testing.T) {
	t.Parallel()

	publishErr := errors.New("nats down")
	publisher := &recordingPublisher{err: publishErr}
	handler := newTestHandler(t, publisher)
	ctx := workflow.WithDispatch(context.Background(), testDispatch)

	err := handler.Handle(ctx, workflow.Event{Scope: workflow.ScopeDispatch, Stage: workflow.StageStart})
	if !errors.Is(err, publishErr) {
		t.Fatalf("Handle() = %v, want %v", err, publishErr)
	}
}

func TestNewHandlerValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewEventHandler(nil, testSubject); err == nil {
		t.Fatal("NewHandler(nil publisher) = nil, want error")
	}
	if _, err := NewEventHandler(&recordingPublisher{}, ""); err == nil {
		t.Fatal("NewHandler(empty subject) = nil, want error")
	}
}

func equalObjects(got, want map[string]any) bool {
	if len(got) != len(want) {
		return false
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			return false
		}
	}
	return true
}
