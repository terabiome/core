package services

import (
	"context"

	"github.com/terabiome/core/internal/models"
	"github.com/terabiome/core/internal/repositories"
)

// Not wrapper around units_of_works table, but for all around "Unit of Works"
type UnitOfWorkRepository interface {
	Create(ctx context.Context, uow models.UnitOfWork) error
	FilterByUnstaged(ctx context.Context, limit int) ([]models.UnitOfWork, error)
	UpdateByUUID(ctx context.Context, uow models.UnitOfWork) error
}

type DispatchRepository interface {
	BulkInsert(ctx context.Context, dispatches []models.Dispatch) error
}

// orchestrate Repository call here, or whatever you like
type TransactionManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type ServiceLeaseRepository interface {
	ClaimOrUpdate(ctx context.Context, agentID string, secretHash string, metadata map[string]any) error
	UpdateHeartbeat(ctx context.Context, agentID string) error
	GetActiveAgentsByCapability(ctx context.Context, capability string) ([]string, error)
	ExpireLease(ctx context.Context, agentID string) error
	GetStatus(ctx context.Context, agentID string) (string, error)
}
