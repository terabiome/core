package runner

import (
	"context"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
)

type WorkflowService interface {
	PrepareExecutionPlan(ctx context.Context, uow yamlcontracts.UnitOfWork) ([]yamlcontracts.UnitOfWorkSequenceBatch, error)
}
