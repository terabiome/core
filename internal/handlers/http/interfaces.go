package httphandlers

import (
	"context"

	manifestcontracts "github.com/terabiome/core/internal/contracts/manifest"
)

type UnitOfWorkService interface {
	Submit(ctx context.Context, request manifestcontracts.UnitOfWork) error
}
