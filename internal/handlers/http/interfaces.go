package httphandlers

import (
	"context"

	yamlcontracts "github.com/terabiome/core/internal/contracts/yaml"
)

type UnitOfWorkService interface {
	Submit(ctx context.Context, request yamlcontracts.UnitOfWork) error
}
