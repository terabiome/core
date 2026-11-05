package services

import (
	"context"

	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/repositories"
)

// Not wrapper around units_of_works table, but for all around "Unit of Works"
type UnitOfWorkRepository interface {
	Create(ctx context.Context, uow models.UnitOfWork) error
	FilterByUnstaged(ctx context.Context, queryParameters repositories.QueryParameters, lockControlParameters repositories.LockControlParameters) ([]models.UnitOfWork, error)
	UpdateByUUID(ctx context.Context, uow models.UnitOfWork) error
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
