package services

import (
	"context"

	"github.com/terabiome/terabiome/internal/models"
)

// Not wrapper around units_of_works table, but for all around "Unit of Works"
type UnitOfWorkRepository interface {
	Create(ctx context.Context, uow models.UnitOfWork) error
}

type ServiceLeaseRepository interface {
	ClaimOrUpdate(ctx context.Context, agentID string, secretHash string, metadata map[string]any) error
	UpdateHeartbeat(ctx context.Context, agentID string) error
	GetActiveAgentsByCapability(ctx context.Context, capability string) ([]string, error)
	ExpireLease(ctx context.Context, agentID string) error
	GetStatus(ctx context.Context, agentID string) (string, error)
}
