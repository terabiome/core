package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/terabiome/core/internal/models"
	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type UnitOfWork struct {
	baseService
	uowRepository UnitOfWorkRepository
}

func NewUnitOfWork(
	txManager TransactionManager,
	uowRepo UnitOfWorkRepository,
) *UnitOfWork {
	return &UnitOfWork{
		baseService:   newBaseService(txManager),
		uowRepository: uowRepo,
	}
}

func (s *UnitOfWork) Submit(ctx context.Context, uowContract manifestcontracts.UnitOfWork) error {
	// YAML is superset, and for our case, can use same structs interchangably
	uowBytes, err := json.Marshal(uowContract)
	if err != nil {
		return fmt.Errorf("failed to serialize UOW YAML input: %w", err)
	}

	uow := models.UnitOfWork{
		Name:   uowContract.Name,
		Config: uowBytes,
	}

	if err = s.uowRepository.Create(ctx, uow); err != nil {
		return fmt.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}
