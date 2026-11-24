package runner

import (
	"context"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type WorkflowService interface {
	PrepareExecutionPlan(ctx context.Context, uow manifestcontracts.UnitOfWork) ([]manifestcontracts.UnitOfWorkSequenceBatch, error)
}
