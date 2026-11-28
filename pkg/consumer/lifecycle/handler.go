package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	eventcontracts "github.com/terabiome/core/pkg/contracts/events"
	"github.com/terabiome/core/pkg/messaging"
	"github.com/terabiome/core/pkg/workflow"
)

type EventHandler struct {
	publisher messaging.Publisher
	subject   string
	clock     func() time.Time
}

func NewEventHandler(publisher messaging.Publisher, subject string) (*EventHandler, error) {
	if publisher == nil {
		return nil, errors.New("lifecycle: nil publisher")
	}
	if subject == "" {
		return nil, errors.New("lifecycle: empty subject")
	}
	return &EventHandler{
		publisher: publisher,
		subject:   subject,
		clock:     time.Now,
	}, nil
}

func (h *EventHandler) Handle(ctx context.Context, event workflow.Event) error {
	dispatch, ok := workflow.DispatchFromContext(ctx)
	if !ok {
		return errors.New("lifecycle: event has no dispatch in context")
	}

	eventType, err := eventTypeFor(event)
	if err != nil {
		return err
	}

	metadata, err := json.Marshal(metadataFor(event))
	if err != nil {
		return fmt.Errorf("lifecycle: encode metadata: %w", err)
	}

	data, err := dataFor(event)
	if err != nil {
		return err
	}
	eventData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("lifecycle: encode data: %w", err)
	}

	payload, err := json.Marshal(eventcontracts.DispatchEvent{
		DispatchUUID: dispatch.UUID,
		UOWUUID:      dispatch.UOWUUID,
		Type:         eventType,
		Metadata:     metadata,
		Data:         eventData,
		GeneratedAt:  h.clock().UTC(),
	})
	if err != nil {
		return fmt.Errorf("lifecycle: encode event: %w", err)
	}

	// Events describe what already happened; a cancelled execution still reports it.
	if err := h.publisher.Publish(context.WithoutCancel(ctx), messaging.Destination{Subject: h.subject}, payload); err != nil {
		return fmt.Errorf("lifecycle: publish %s: %w", eventType, err)
	}
	return nil
}

func eventTypeFor(event workflow.Event) (eventcontracts.Type, error) {
	switch event.Scope {
	case workflow.ScopeDispatch:
		switch event.Stage {
		case workflow.StageStart:
			return eventcontracts.TypeDispatchStart, nil
		case workflow.StageFinish:
			return eventcontracts.TypeDispatchFinish, nil
		}
	case workflow.ScopeSequence:
		switch event.Stage {
		case workflow.StageStart:
			return eventcontracts.TypeSequenceStart, nil
		case workflow.StageFinish:
			return eventcontracts.TypeSequenceFinish, nil
		}
	case workflow.ScopeStep:
		switch event.Stage {
		case workflow.StageStart:
			return eventcontracts.TypeStepStart, nil
		case workflow.StageFinish:
			return eventcontracts.TypeStepFinish, nil
		}
	}

	return "", fmt.Errorf("lifecycle: unsupported event %s/%s", event.Scope, event.Stage)
}

func metadataFor(event workflow.Event) map[string]any {
	metadata := map[string]any{}
	switch event.Scope {
	case workflow.ScopeSequence:
		metadata["sequence"] = event.Sequence
	case workflow.ScopeStep:
		metadata["sequence"] = event.Sequence
		metadata["step"] = event.Step
		metadata["step_index"] = event.Index
	}
	return metadata
}

func dataFor(event workflow.Event) (map[string]any, error) {
	if event.Stage == workflow.StageStart {
		return map[string]any{}, nil
	}
	if event.Status == "" {
		return nil, errors.New("lifecycle: finish event has no status")
	}

	data := map[string]any{"status": string(event.Status)}
	if event.Err != nil {
		data["error"] = event.Err.Error()
	}
	return data, nil
}
