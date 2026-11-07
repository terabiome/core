package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/terabiome/terabiome/internal/repositories"
)

type Dispatch struct {
	baseService
	uowRepository UnitOfWorkRepository
}

func NewDispatch(
	txManager TransactionManager,
	uowRepo UnitOfWorkRepository,
) *Dispatch {
	return &Dispatch{
		baseService:   newBaseService(txManager),
		uowRepository: uowRepo,
	}
}

func (s *Dispatch) Stage(ctx context.Context, limit int) error {
	txErr := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		// query "units_of_work" by "staged_at=NULL"
		uows, err := s.uowRepository.FilterByUnstaged(
			ctx,
			repositories.QueryParameters{
				SortBy: []repositories.SortField{{Name: "uuid", Order: repositories.SortOrderAscending}},
				Limit:  limit,
			},
			repositories.LockControlParameters{
				Type: repositories.LockTypeLockForNoKeyUpdate,
			})
		if err != nil {
			return fmt.Errorf("failed to get unstaged unit-of-work: %w", err)
		}

		// something else
		for _, uow := range uows {
			// TODO: fetch something and create "dispatch" record
			// question: do we resolve targets and place it in dispatch data?
			// design:
			// - 1 dispatch is swept later, set dispatched_at = [time.Now]
			// - given that, a dispatch can be tied to N targets, so it is hard to tell when it is considered fully dispatched
			// 	+ to simplify, when resolving NATS topics to emit the event to, it tracks the events. if emitted into 3 topics, and all successful, it counts as FULLY_DISPATCHED
			// 	+ rest of transitions of the downstream consumers have nothing to do with the status of the dispatch
			uow.StagedAt = time.Now().UTC()
			if subErr := s.uowRepository.UpdateByUUID(ctx, uow); subErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to update uow uuid=%s: %w", uow.UUID, err))
			}
		}

		// throw here
		if err != nil {
			return fmt.Errorf("failed to update uows, rolling back: %w", err)
		}

		return nil
	})
	return txErr
}
