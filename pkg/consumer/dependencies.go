package consumer

import (
	"fmt"

	shellexecutors "github.com/terabiome/core/pkg/executors/shell"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging/natsjetstream"
	"github.com/terabiome/core/pkg/operations"
	"github.com/terabiome/core/pkg/operations/builtin"
	"github.com/terabiome/core/pkg/workflow"
)

type Dependencies struct {
	Consumer  MessagingConsumer
	Registry  Registry
	Journaler Journaler
	Steps     workflow.StepHandler
	Wrap      func(Handler) Handler
}

func (d *Dependencies) orEmpty() *Dependencies {
	if d == nil {
		return &Dependencies{}
	}
	return d
}

func consumerFrom(config Config, deps *Dependencies, logger logging.Logger) (MessagingConsumer, bool, error) {
	if deps.Consumer != nil {
		return deps.Consumer, false, nil
	}

	created, err := natsjetstream.New(config.Jetstream, logger)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create jetstream client: %w", err)
	}
	return created, true, nil
}

func registryFrom(deps *Dependencies, logger logging.Logger) (Registry, error) {
	if deps.Registry != nil {
		return deps.Registry, nil
	}

	logger.Warn("no registry passed, using builtin one")
	return builtin.NewRegistry()
}

func journalerFrom(config Config, deps *Dependencies, logger logging.Logger) (Journaler, error) {
	if deps.Journaler != nil {
		return deps.Journaler, nil
	}
	return newJournaler(config.Journal, logger)
}

func handlerFrom(config Config, deps *Dependencies, journaler Journaler, logger logging.Logger) (Handler, error) {
	registry, err := registryFrom(deps, logger)
	if err != nil {
		return nil, err
	}

	env := &operations.OperationEnvironment{
		NodeInfo:  config.Node,
		Journaler: journaler,
		Executor:  shellexecutors.NewCommandExecutor(logger),
	}
	handler := Handler(&messageHandler{
		executor: workflow.NewWorkflowExecutor(env, registry, config.Concurrency, deps.Steps),
		logger:   logger,
	})
	if deps.Wrap != nil {
		handler = deps.Wrap(handler)
	}
	return handler, nil
}
