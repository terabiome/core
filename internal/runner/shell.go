package runner

import (
	"context"
	"runtime"
	"time"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	shellexecutors "github.com/terabiome/core/pkg/executors/shell"
	journalerspkg "github.com/terabiome/core/pkg/journalers"
	loggingpkg "github.com/terabiome/core/pkg/logging"
	operationspkg "github.com/terabiome/core/pkg/operations"
	workflowpkg "github.com/terabiome/core/pkg/workflow"
)

type ShellRunner struct {
	logger         loggingpkg.Logger
	registry       *operationspkg.Registry
	maxConcurrency int
}

func NewShellRunner(
	logger loggingpkg.Logger,
	registry *operationspkg.Registry,
) *ShellRunner {
	return &ShellRunner{
		logger:         logger,
		registry:       registry,
		maxConcurrency: runtime.NumCPU(),
	}
}

func (r *ShellRunner) Run(ctx context.Context, uow manifestcontracts.UnitOfWork) error {
	journaler, _ := journalerspkg.NewJournaler(journalerspkg.JournalerParameters{
		MaxRecords:    50,
		FlushInterval: 1 * time.Second,
		BatchFn: func(entries []journalerspkg.JournalEntry) {
			for _, e := range entries {
				r.logger.InfoContext(ctx, e.Message,
					"timestamp", e.Timestamp.Format(time.RFC3339),
					"source", e.Source,
				)
			}
		},
	})
	journaler.Start()
	defer journaler.Close()

	env := &operationspkg.OperationEnvironment{
		NodeInfo:  r.probeNode(),
		Journaler: journaler,
		Executor:  shellexecutors.NewCommandExecutor(r.logger),
	}

	wfExec := workflowpkg.NewWorkflowExecutor(env, r.registry, r.maxConcurrency, nil)
	r.logger.InfoContext(ctx, "Starting local execution...")
	return wfExec.ExecuteAsDAG(ctx, uow.Sequences)
}

func (r *ShellRunner) probeNode() operationspkg.NodeInfo {
	return operationspkg.NodeInfo{
		MachineID: "local-shell",
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	}
}
