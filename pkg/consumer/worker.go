package consumer

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/terabiome/core/pkg/journalers"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging"
)

const stopTimeout = 5 * time.Second

type Worker struct {
	config    Config
	logger    logging.Logger
	consumer  MessagingConsumer
	journaler Journaler
	handler   Handler
	started   atomic.Bool
	closed    atomic.Bool
}

func NewWorker(config Config, deps *Dependencies, logger logging.Logger) (*Worker, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if err := config.normalize(); err != nil {
		return nil, err
	}
	deps = deps.orEmpty()

	consumer, owned, err := consumerFrom(config, deps, logger)
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*Worker, error) {
		if owned {
			_ = consumer.Close()
		}
		return nil, err
	}

	journaler, err := journalerFrom(config, deps, logger)
	if err != nil {
		return fail(err)
	}

	handler, err := handlerFrom(config, deps, journaler, logger)
	if err != nil {
		if deps.Journaler == nil {
			_ = journaler.Close()
		}
		return fail(err)
	}

	return &Worker{
		config:    config,
		logger:    logger,
		consumer:  consumer,
		journaler: journaler,
		handler:   handler,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	if !w.started.CompareAndSwap(false, true) {
		return errors.New("worker already started")
	}
	defer w.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	if startableJournaler, ok := w.journaler.(interface{ Start() }); ok {
		startableJournaler.Start()
	}

	// Handler work outlives the signal so draining can finish, until the
	// shutdown timeout cancels it explicitly.
	workCtx, cancelWork := context.WithCancel(context.WithoutCancel(ctx))
	defer cancelWork()

	subscriptions, err := w.subscribe(workCtx)
	if err != nil {
		w.shutdown(subscriptions, cancelWork, stopTimeout)
		return err
	}

	w.logger.InfoContext(ctx, "worker running",
		"subscriptions", len(subscriptions),
		"concurrency", w.config.Concurrency,
	)

	<-ctx.Done()
	w.shutdown(subscriptions, cancelWork, w.config.ShutdownTimeout)
	return ctx.Err()
}

func (w *Worker) Close() error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}

	var errs []error
	if w.consumer != nil {
		errs = append(errs, w.consumer.Close())
	}
	if w.journaler != nil {
		errs = append(errs, w.journaler.Close())
	}
	return errors.Join(errs...)
}

func (w *Worker) subscribe(ctx context.Context) ([]*messaging.Subscription, error) {
	subscriptions := make([]*messaging.Subscription, 0, len(w.config.Subscriptions))

	for _, subscription := range w.config.Subscriptions {
		source := messaging.Source{
			Stream:  subscription.Stream,
			Subject: subscription.Subject,
			Options: subscription.Options,
		}
		ackWait := subscription.Options.AckWait
		processor := processorFunc(func(ctx context.Context, message messaging.Message) error {
			if ackWait > 0 {
				stopHeartbeat := startHeartbeat(ctx, message, ackWait)
				defer stopHeartbeat()
			}
			return w.handler.Handle(ctx, message)
		})

		handle, err := w.consumer.Consume(ctx, source, processor)
		if err != nil {
			return subscriptions, fmt.Errorf(
				"failed to subscribe to %s on stream %s: %w",
				subscription.Subject, subscription.Stream, err,
			)
		}
		subscriptions = append(subscriptions, handle)
	}

	return subscriptions, nil
}

func (w *Worker) shutdown(
	subscriptions []*messaging.Subscription,
	cancelWork context.CancelFunc,
	timeout time.Duration,
) {
	drain(subscriptions)
	if w.await(subscriptions, timeout) {
		return
	}

	w.logger.Warn("worker shutdown timed out, stopping in-flight work",
		"timeout", timeout)
	cancelWork()
	stop(subscriptions)
	w.await(subscriptions, stopTimeout)
}

func (w *Worker) await(subscriptions []*messaging.Subscription, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for _, subscription := range subscriptions {
		done := subscription.Done()
		if done == nil {
			continue
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			select {
			case <-done:
				continue
			default:
				return false
			}
		}

		timer := time.NewTimer(remaining)
		select {
		case <-done:
			timer.Stop()
		case <-timer.C:
			return false
		}
	}
	return true
}

func newJournaler(config JournalConfig, logger logging.Logger) (*journalers.Journaler, error) {
	params := journalers.JournalerParameters{
		MaxRecords:    config.MaxRecords,
		FlushInterval: config.FlushInterval,
		BatchFn:       config.BatchFn,
	}
	if params.MaxRecords <= 0 {
		params.MaxRecords = 50
	}
	if params.FlushInterval <= 0 {
		params.FlushInterval = time.Second
	}
	if params.BatchFn == nil {
		params.BatchFn = func(entries []journalers.JournalEntry) {
			for _, entry := range entries {
				logger.Info(entry.Message,
					"timestamp", entry.Timestamp.Format(time.RFC3339),
					"source", entry.Source,
				)
			}
		}
	}
	return journalers.NewJournaler(params)
}

func drain(subscriptions []*messaging.Subscription) {
	for _, subscription := range subscriptions {
		subscription.Drain()
	}
}

func stop(subscriptions []*messaging.Subscription) {
	for _, subscription := range subscriptions {
		subscription.Stop()
	}
}
