package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/terabiome/core/internal/models"
	"github.com/terabiome/core/internal/repositories"
	dispatchcontracts "github.com/terabiome/core/pkg/contracts/dispatch"
	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	"github.com/terabiome/core/pkg/helpers"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging"
	"golang.org/x/sync/errgroup"
)

type Dispatch struct {
	baseService
	uowRepository      UnitOfWorkRepository
	dispatchRepository DispatchRepository
	publisher          messaging.Publisher
}

func NewDispatch(
	txManager TransactionManager,
	uowRepo UnitOfWorkRepository,
	dispatchRepo DispatchRepository,
	publisher messaging.Publisher,
) *Dispatch {
	return &Dispatch{
		baseService:        newBaseService(txManager),
		uowRepository:      uowRepo,
		dispatchRepository: dispatchRepo,
		publisher:          publisher,
	}
}

// sweepCounts are outcomes committed by one pass, not rows touched and then rolled back.
type sweepCounts struct {
	staged            int
	dispatchesCreated int
	booked            int
	dispatched        int
	ambiguous         int
	failed            int
}

func (c sweepCounts) empty() bool {
	return c.staged == 0 && c.dispatchesCreated == 0 && c.booked == 0 &&
		c.dispatched == 0 && c.ambiguous == 0 && c.failed == 0
}

func (s *Dispatch) Stage(ctx context.Context, limit int) (sweepCounts, error) {
	var counts sweepCounts
	txErr := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		// query "units_of_work" by "staged_at=NULL"
		uows, err := s.uowRepository.FilterByUnstaged(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to get unstaged unit-of-work: %w", err)
		}

		// something else
		for _, uow := range uows {
			// TODO: fetch something and create "dispatch" record
			var uowContract manifestcontracts.UnitOfWork
			if subErr := json.Unmarshal(uow.Config, &uowContract); subErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to load config of unit-of-work uuid=%s: %w", uow.UUID, subErr))
				continue
			}

			// double-check just to make sure
			if uowContract.CheckIfContainCycle() {
				err = errors.Join(err, fmt.Errorf("config of unit-of-work uuid=%s contains cycle", uow.UUID))
				continue
			}

			// DB column is JSONB, unfortunately
			sequences, subErr := json.Marshal(uowContract.Sequences)
			if subErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to serialize sequence to [json.RawMessage] for unit-of-work uuid=%s: %w", uow.UUID, subErr))
				continue
			}

			// question: do we resolve targets and place it in dispatch data?
			// design:
			// - 1 dispatch is swept later, set dispatched_at = [time.Now]
			// - given that, a dispatch can be tied to N targets, so it is hard to tell when it is considered fully dispatched
			// 	+ to simplify, when resolving NATS topics to emit the event to, it tracks the events. if emitted into 3 topics, and all successful, it counts as FULLY_DISPATCHED
			// 	+ rest of transitions of the downstream consumers have nothing to do with the status of the dispatch

			dispatches := make([]models.Dispatch, 0, len(uowContract.Targets))
			for _, targetYAML := range uowContract.Targets {
				target, subErr := json.Marshal(targetYAML)
				if subErr != nil {
					err = errors.Join(err, fmt.Errorf("failed to serialize target to [json.RawMessage] for unit-of-work uuid=%s: %w", uow.UUID, subErr))
					break
				}

				dispatches = append(dispatches, models.Dispatch{
					UOWUUID: uow.UUID,
					Status:  models.DispatchStatusNew,
					Target:  target,
					Payload: sequences,
				})
			}

			// bulk insert smh
			if subErr := s.dispatchRepository.BulkInsert(ctx, dispatches); subErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to bulk insert dispatches for uow uuid=%s: %w", uow.UUID, subErr))
				continue
			}

			stagedAt := time.Now().UTC()
			uow.StagedAt = &stagedAt
			if subErr := s.uowRepository.UpdateByUUID(ctx, uow); subErr != nil {
				// inserts are already in this transaction; committing would restage the same unit
				return fmt.Errorf("failed to update uow uuid=%s: %w", uow.UUID, subErr)
			}
			counts.staged++
			counts.dispatchesCreated += len(dispatches)
		}

		// commit since partial failure is acceptable
		// as long as we don't write anything and let the next loop attempt to reconcile

		if err != nil {
			logging.GetGlobalSlogger().WarnContext(ctx, "encountered error updating uows, do partial commit anyway", "error", err)
		}

		return nil
	})
	if txErr != nil {
		// the stage transaction rolled back, including any units counted above
		return sweepCounts{}, txErr
	}
	return counts, nil
}

func (s *Dispatch) BookDispatches(ctx context.Context, limit int) (sweepCounts, error) {
	// build a KEY for this call to BOOK dispatches and PROCESS them
	// context should carry a timestamp for deadline if not exist, for later claim during catastrophic execution
	deadline, isDeadlineSet := ctx.Deadline()
	if !isDeadlineSet {
		// no need to use [context.WithDeadline]
		deadline = time.Now().UTC().Add(120 * time.Second)
	}

	idempotencyKey, err := loadIdempotencyKeyFromCtx(ctx)
	if err != nil || idempotencyKey == "" {
		// tolerate
		idempotencyKey = fmt.Sprintf("%s:%s", "dispatch:autogenerated", uuid.NewV7())
		ctx = setIdempotencyKeyInCtx(ctx, idempotencyKey)
	}

	// start a transaction to book the dispatch records
	var counts sweepCounts
	txErr := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		dispatchFilterOutput, err := s.dispatchRepository.FilterByStatus(ctx, models.DispatchStatusNew, limit)
		if err != nil {
			return fmt.Errorf("failed to get new dispatches: %w", err)
		}

		// mark dispatch as STAGED
		var dispatchUpdateInputBatch = make([]repositories.DispatchUpdateInput, 0, len(dispatchFilterOutput.Records))
		var bookErr error
		for _, dispatchFilterRecord := range dispatchFilterOutput.Records {
			metadataMap, unmarshalErr := helpers.JSON().Unmarshal[map[string]any](dispatchFilterRecord.Metadata)
			if unmarshalErr != nil {
				bookErr = errors.Join(bookErr, fmt.Errorf("failed to deserialize metadata from [json.RawMessage] to [%T]: %w", metadataMap, unmarshalErr))
				continue
			}

			subMap := map[string]any{
				"dispatch_key":          fmt.Sprintf("%s:%s", idempotencyKey, dispatchFilterRecord.UUID),
				"dispatch_key_deadline": deadline.UnixMilli(),
				"dispatch_retries":      0,
			}
			if v, ok := metadataMap["dispatch_retries"]; ok {
				retries, ok := anyToInt(v) // JSON objects decode numbers as float64
				if ok {
					subMap["dispatch_retries"] = retries + 1
				}
			}

			metadata, marshalErr := json.Marshal(helpers.Map().ImmutableMerge(metadataMap, subMap))
			if marshalErr != nil {
				bookErr = errors.Join(bookErr, fmt.Errorf("failed to serialize metadata to [json.RawMessage]: %w", marshalErr))
				continue
			}
			dispatchFilterRecord.Metadata = metadata

			dispatchUpdateInputBatch = append(dispatchUpdateInputBatch, repositories.DispatchUpdateInput{
				Filters: repositories.DispatchUpdateFilters{
					UUID: &dispatchFilterRecord.UUID,
				},
				Fields: repositories.DispatchUpdateFields{
					Metadata: &dispatchFilterRecord.Metadata,
					Status:   models.DispatchStatusStaged,
				},
			})
		}

		if bookErr != nil {
			return bookErr
		}

		// assume this always sets [models.Dispatch.UpdatedAt]
		if err = s.dispatchRepository.BulkUpdate(ctx, dispatchUpdateInputBatch); err != nil {
			return fmt.Errorf("failed to bulk update dispatches: %w", err)
		}
		counts.booked = len(dispatchUpdateInputBatch)

		return nil
	})
	if txErr != nil {
		return sweepCounts{}, txErr
	}
	return counts, nil
}

// RunSweep is one loop: stage, book, then publish. It does not start the next pass until the current one returns.
// ctx stops the loop between passes. An in-flight pass keeps its own deadline so a signal does not cut a publish in half.
func (s *Dispatch) RunSweep(ctx context.Context, limit int, interval time.Duration, passTimeout time.Duration) {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		passCtx, cancel := context.WithTimeout(context.Background(), passTimeout)
		s.sweepOnce(passCtx, limit)
		cancel()

		if ctx.Err() != nil {
			return
		}

		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func (s *Dispatch) sweepOnce(ctx context.Context, limit int) {
	var counts sweepCounts
	staged, err := s.Stage(ctx, limit)
	counts.staged = staged.staged
	counts.dispatchesCreated = staged.dispatchesCreated
	if err != nil {
		logging.GetGlobalSlogger().ErrorContext(ctx, "dispatch stage failed", "error", err)
	}
	if ctx.Err() != nil {
		s.logSweep(ctx, counts)
		return
	}
	booked, err := s.BookDispatches(ctx, limit)
	counts.booked = booked.booked
	if err != nil {
		logging.GetGlobalSlogger().ErrorContext(ctx, "dispatch book failed", "error", err)
	}
	if ctx.Err() != nil {
		s.logSweep(ctx, counts)
		return
	}
	processed, err := s.ProcessBookedDispatches(ctx, limit)
	counts.dispatched = processed.dispatched
	counts.ambiguous = processed.ambiguous
	counts.failed = processed.failed
	if err != nil {
		logging.GetGlobalSlogger().ErrorContext(ctx, "dispatch process failed", "error", err)
	}
	s.logSweep(ctx, counts)
}

func (s *Dispatch) logSweep(ctx context.Context, counts sweepCounts) {
	if counts.empty() {
		return
	}
	logging.GetGlobalSlogger().InfoContext(ctx, "dispatch sweep",
		"staged", counts.staged,
		"dispatches_created", counts.dispatchesCreated,
		"booked", counts.booked,
		"dispatched", counts.dispatched,
		"ambiguous", counts.ambiguous,
		"failed", counts.failed,
	)
}

type stagedPublish struct {
	uuid     uuid.UUID
	topics   []string
	envelope []byte
}

func (s *Dispatch) ProcessBookedDispatches(ctx context.Context, limit int) (sweepCounts, error) {
	// rule: do not allow retry on FAILED dispatch record
	// tx 1 claims rows and commits before any broker call, so the pool connection is not held across NATS.
	// AMBIGUOUS is the claim: another sweeper will not select these STAGED rows, and a crash mid-publish stays ambiguous.
	ready, counts, claimErr := s.claimStagedForPublish(ctx, limit)
	if claimErr != nil && len(ready) == 0 {
		return counts, claimErr
	}

	var publishErr error
	dispatched := make([]uuid.UUID, 0, len(ready))
	for _, job := range ready {
		if err := s.publishFanOut(ctx, job.topics, job.envelope); err != nil {
			// already claimed AMBIGUOUS; do not try to guess whether a subset of topics landed
			publishErr = errors.Join(publishErr, fmt.Errorf("failed to publish dispatch envelope for dispatch uuid=%v: %w", job.uuid, err))
			counts.ambiguous++
			continue
		}
		dispatched = append(dispatched, job.uuid)
	}

	// tx 2 records only the publishes that returned success.
	// a failed mark rolls the whole batch back, so those publishes stay AMBIGUOUS.
	markErr := s.markDispatched(ctx, dispatched)
	if markErr != nil {
		counts.ambiguous += len(dispatched)
	} else {
		counts.dispatched = len(dispatched)
	}
	return counts, errors.Join(claimErr, publishErr, markErr)
}

func (s *Dispatch) claimStagedForPublish(ctx context.Context, limit int) ([]stagedPublish, sweepCounts, error) {
	var ready []stagedPublish
	var counts sweepCounts
	var surfacingError error

	txErr := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		dispatchFilterOutput, err := s.dispatchRepository.FilterByStatus(ctx, models.DispatchStatusStaged, limit)
		if err != nil {
			return fmt.Errorf("failed to get staged dispatches: %w", err)
		}

		for _, dispatch := range dispatchFilterOutput.Records {
			dispatchMetadata, marshalErr := helpers.JSON().Unmarshal[map[string]any](dispatch.Metadata)
			if marshalErr != nil {
				subErr := fmt.Errorf("failed to deserialize dispatch metadata of uuid=%v: %w", dispatch.UUID, marshalErr)
				updateErr := s.markDispatchStatus(ctx, dispatch.UUID, models.DispatchStatusFailed)
				if updateErr != nil {
					subErr = errors.Join(subErr, updateErr)
				} else {
					counts.failed++
				}
				err = errors.Join(err, subErr)
				continue
			}

			envelopeBlob, marshalErr := json.Marshal(dispatchcontracts.DispatchEnvelope{
				Metadata: helpers.Map().ImmutableMerge(dispatchMetadata, map[string]any{
					"dispatch_uuid": dispatch.UUID,
					"uow_uuid":      dispatch.UOWUUID,
					"target":        dispatch.Target, // for workers wondering why it gets the message lol
				}),
				Payload: dispatch.Payload,
			})
			if marshalErr != nil {
				subErr := fmt.Errorf("failed to serialize NATS envelope for dispatch uuid=%v: %w", dispatch.UUID, marshalErr)
				updateErr := s.markDispatchStatus(ctx, dispatch.UUID, models.DispatchStatusFailed)
				if updateErr != nil {
					subErr = errors.Join(subErr, updateErr)
				} else {
					counts.failed++
				}
				err = errors.Join(err, subErr)
				continue
			}

			dispatchUnixDeadline, isSet, canCast := helpers.Map().GetInt64(dispatchMetadata, "dispatch_key_deadline")
			if !isSet {
				dispatchUnixDeadline = time.Now().Add(120 * time.Second).UTC().UnixMilli()
			} else if !canCast {
				logging.GetGlobalSlogger().ErrorContext(ctx, "could not cast dispatch_deadline to milli unix, ignored", "dispatch_uuid", dispatch.UUID)
				err = errors.Join(err, fmt.Errorf("failed to cast dispatch_deadline of uuid=%v", dispatch.UUID))
				continue
			} else if dispatchUnixDeadline-time.Now().UTC().UnixMilli() < 2000 {
				logging.GetGlobalSlogger().WarnContext(ctx, "exceed or too close to deadline, marking as AMBIGUOUS", "dispatch_uuid", dispatch.UUID)
				if updateErr := s.markDispatchStatus(ctx, dispatch.UUID, models.DispatchStatusAmbiguous); updateErr != nil {
					err = errors.Join(err, updateErr)
				} else {
					counts.ambiguous++
				}
				continue
			}

			topics, resolveErr := resolveTopics(dispatch.Target)
			if resolveErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to resolve topics for dispatch uuid=%v: %w", dispatch.UUID, resolveErr))
				continue
			}

			// claim before commit so a second sweeper cannot publish the same STAGED row
			if updateErr := s.markDispatchStatus(ctx, dispatch.UUID, models.DispatchStatusAmbiguous); updateErr != nil {
				err = errors.Join(err, updateErr)
				continue
			}

			ready = append(ready, stagedPublish{
				uuid:     dispatch.UUID,
				topics:   topics,
				envelope: envelopeBlob,
			})
		}

		if err != nil {
			// partial claim is acceptable; do not roll back rows already marked
			surfacingError = err
		}
		return nil
	})

	if txErr != nil {
		// claim did not commit, so nothing in ready is safe to publish and status writes rolled back
		return nil, sweepCounts{}, errors.Join(surfacingError, txErr)
	}
	return ready, counts, surfacingError
}

func (s *Dispatch) markDispatched(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return s.txManager.Wrap(ctx, func(ctx context.Context) error {
		var err error
		for _, id := range ids {
			if updateErr := s.markDispatchStatus(ctx, id, models.DispatchStatusDispatched); updateErr != nil {
				err = errors.Join(err, updateErr)
			}
		}
		// commit the ones that updated; surface the rest
		return err
	})
}

func (s *Dispatch) markDispatchStatus(ctx context.Context, id uuid.UUID, status models.DispatchStatus) error {
	updateErr := s.dispatchRepository.Update(ctx, repositories.DispatchUpdateInput{
		Filters: repositories.DispatchUpdateFilters{
			UUID: &id,
		},
		Fields: repositories.DispatchUpdateFields{
			Status: status,
		},
	})
	if updateErr != nil {
		return fmt.Errorf("failed to update dispatch uuid=%v status to %s: %w", id, status, updateErr)
	}
	return nil
}

// move this somewhere later
// can abstract [jetstream.Jetstream] but not recommended for now
func (s *Dispatch) publishFanOut(ctx context.Context, topics []string, data []byte) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(4) // magic number, define this in goroutine pool later

	for _, topic := range topics {
		g.Go(func(topic string) func() error {
			return func() error {
				err := s.publisher.Publish(ctx, messaging.Destination{Subject: topic}, data)
				if err != nil {
					return fmt.Errorf("failed to publish to %s: %w", topic, err)
				}
				return nil
			}
		}(topic))
	}

	return g.Wait()
}

// need to move this somewhere, but put it here for now
func resolveTopics(target json.RawMessage) ([]string, error) {
	targetYAML, err := helpers.JSON().Unmarshal[manifestcontracts.UnitOfWorkTarget](target)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize UOW target to YAML Go struct: %w", err)
	}
	logging.GetGlobalSlogger().Info("ignore all but topic matching (deliberate)")
	return targetYAML.Topics, nil
}

// anyToInt accepts the numeric types encoding/json stores in map[string]any.
func anyToInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}
