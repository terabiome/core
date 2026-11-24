package httphandlers

import (
	"context"

	"github.com/go-playground/validator/v10"
	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type UnitOfWorkService interface {
	Submit(ctx context.Context, request manifestcontracts.UnitOfWork) error
}

type Validatable interface {
	Validate(validator *validator.Validate) error
}
