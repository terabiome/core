package services

import (
	"context"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/planner"
)

type WorkflowService struct {
}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{}
}

// PrepareExecutionPlan takes a UOW and returns structured batches ready for the Executor.
func (s *WorkflowService) PrepareExecutionPlan(ctx context.Context, uow yamlcontracts.UnitOfWork) ([][]yamlcontracts.UnitOfWorkSequence, error) {
	// log the activity
	return planner.GenerateSequenceBatch(uow)
}
