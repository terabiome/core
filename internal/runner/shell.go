package runner

import (
	"context"
	"runtime"
	"time"

	"github.com/terabiome/core/internal/executor"
	"github.com/terabiome/core/internal/writers"
	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	shellexecutors "github.com/terabiome/core/pkg/executors/shell"
	loggingpkg "github.com/terabiome/core/pkg/logging"
	operationspkg "github.com/terabiome/core/pkg/operations"
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
	journaler, _ := writers.NewJournaler(writers.JournalerParameters{
		MaxRecords:    50,
		FlushInterval: 1 * time.Second,
		BatchFn: func(entries []writers.JournalEntry) {
			for _, e := range entries {
				r.logger.InfoContext(ctx, e.Message,
					"timestamp", e.Timestamp.Format(time.RFC3339),
					"source", e.Source,
				)
			}
		},
	})
	defer journaler.Close()

	env := &operationspkg.OperationEnvironment{
		NodeInfo:  r.probeNode(),
		Journaler: journaler,
		Executor:  shellexecutors.NewCommandExecutor(r.logger),
	}

	wfExec := executor.NewWorkflowExecutor(env, r.registry, r.maxConcurrency)
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
