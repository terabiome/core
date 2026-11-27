package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dispatchcontracts "github.com/terabiome/core/pkg/contracts/dispatch"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging"
	"github.com/terabiome/core/pkg/workflow"
)

type processorFunc func(ctx context.Context, message messaging.Message) error

func (f processorFunc) Handle(ctx context.Context, message messaging.Message) error {
	return f(ctx, message)
}

type messageHandler struct {
	executor *workflow.WorkflowExecutor
	logger   logging.Logger
}

func (h *messageHandler) Handle(
	ctx context.Context,
	message messaging.Message,
) error {
	var envelope dispatchcontracts.DispatchEnvelope
	if err := json.Unmarshal(message.Data(), &envelope); err != nil {
		return messaging.Terminal(fmt.Errorf("failed to decode dispatch envelope: %w", err))
	}

	sequences, err := envelope.Sequences()
	if err != nil {
		return messaging.Terminal(err)
	}

	if err := workflow.ValidateSequenceGraph(sequences); err != nil {
		return messaging.Terminal(fmt.Errorf("invalid sequence graph: %w", err))
	}

	h.logger.InfoContext(ctx, "executing dispatch",
		"dispatch_uuid", envelope.Metadata["dispatch_uuid"],
		"uow_uuid", envelope.Metadata["uow_uuid"],
		"sequences", len(sequences),
	)

	if err := h.executor.ExecuteAsDAG(ctx, sequences); err != nil {
		return fmt.Errorf("failed to execute unit of work: %w", err)
	}
	return nil
}

func startHeartbeat(ctx context.Context, message messaging.Message, ackWait time.Duration) func() {
	interval := ackWait / 3
	if interval <= 0 {
		interval = time.Second
	}

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = message.InProgress()
			case <-stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() { close(stop) }
}
