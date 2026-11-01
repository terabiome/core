package services

import (
	"context"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/models"
	"go.yaml.in/yaml/v4"
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

func (s *UnitOfWork) Submit(ctx context.Context, uowContract yamlcontracts.UnitOfWork) error {
	uowBytes, err := yaml.Marshal(uowContract)
	if err != nil {
		return s.Errorf("failed to serialize UOW YAML input: %w", err)
	}

	uow := models.UnitOfWork{
		Name:       uowContract.Name,
		Status:     models.UnitOfWorkStatusNew,
		ConfigJSON: uowBytes,
	}

	if err = s.uowRepository.Create(ctx, uow); err != nil {
		return s.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}
